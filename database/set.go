package database

import (
	Set "hangdis/datastruct/set"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"strconv"
)

func (db *DB) getAsSet(key string) (Set.Set, redis.ErrorReply) {
	entity, exist := db.GetEntity(key)
	if !exist {
		return nil, nil
	}
	set, ok := entity.Data.(Set.Set)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return set, nil
}

func (db *DB) getOrInitSet(key string) (Set.Set, bool, redis.ErrorReply) {
	set, err := db.getAsSet(key)
	if err != nil {
		return nil, false, err
	}
	isNew := false
	if set == nil {
		set = Set.Make()
		db.PutEntity(key, &database.DataEntity{Data: set})
		isNew = true
	}
	return set, isNew, nil
}

func undoSetChange(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	memberArgs := args[1:]
	members := make([]string, len(memberArgs))
	for i, mem := range memberArgs {
		members[i] = string(mem)
	}
	return rollbackSetMembers(db, key, members...)
}

func execSAdd(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	members := args[1:]
	count := 0
	set, _, err := db.getOrInitSet(key)
	if err != nil {
		return err
	}
	for _, member := range members {
		count += set.Add(string(member))
	}
	db.addAof(utils.ToCmdLine3("sadd", args...))
	return protocol.MakeIntReply(int64(count))
}

func execSIsMember(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	member := string(args[1])
	set, err := db.getAsSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeIntReply(0)
	}
	if set.Has(member) {
		return protocol.MakeIntReply(1)
	}
	return protocol.MakeIntReply(0)
}

func execSRem(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	members := args[1:]
	set, err := db.getAsSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeIntReply(0)
	}
	count := 0
	for _, member := range members {
		count += set.Remove(string(member))
	}
	if count == 0 {
		db.Remove(key)
	}
	if count > 0 {
		db.addAof(utils.ToCmdLine3("srem", args...))
	}
	return protocol.MakeIntReply(int64(count))
}

func execSPop(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	set, err := db.getAsSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	count := 1
	if len(args) > 1 {
		count64, err2 := strconv.ParseInt(string(args[1]), 10, 64)
		if err2 != nil || count64 <= 0 {
			return protocol.MakeErrReply("ERR value is out of range, must be positive")
		}
		count = int(count64)
	}
	n := set.Len()
	if count > n {
		count = n
	}
	if count == 0 {
		return protocol.MakeEmptyMultiBulkReply()
	}
	slice := make([][]byte, count)
	members := set.RandomDistinctMembers(count)
	for i := 0; i < count; i++ {
		slice[i] = []byte(members[i])
		set.Remove(members[i])
	}
	if count > 0 {
		db.addAof(utils.ToCmdLine3("spop", args...))
	}
	return protocol.MakeMultiBulkReply(slice)
}

func execSCard(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	set, err := db.getAsSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeIntReply(0)
	}
	return protocol.MakeIntReply(int64(set.Len()))
}

func execSMembers(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	set, err := db.getAsSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	res := make([][]byte, set.Len())
	for i, s := range set.ToSlice() {
		res[i] = []byte(s)
	}
	return protocol.MakeMultiBulkReply(res)
}

func setReply(set *Set.Set) redis.Reply {
	res := make([][]byte, (*set).Len())
	i := 0
	(*set).ForEach(func(member string) bool {
		res[i] = []byte(member)
		i++
		return true
	})
	return protocol.MakeMultiBulkReply(res)
}

func execSInter(db *DB, args [][]byte) redis.Reply {
	n := len(args)
	sets := make([]*Set.Set, n)
	for i := 0; i < n; i++ {
		key := string(args[i])
		set, err := db.getAsSet(key)
		if err != nil {
			return err
		}
		if set == nil || set.Len() == 0 {
			return protocol.MakeEmptyMultiBulkReply()
		}
		sets[i] = &set
	}
	intersect := Set.Intersect(sets...)
	return setReply(&intersect)
}

func execSInterStore(db *DB, args [][]byte) redis.Reply {
	dest := string(args[0])
	n := len(args) - 1
	sets := make([]*Set.Set, n)
	for i := 1; i <= n; i++ {
		key := string(args[i])
		set, err := db.getAsSet(key)
		if err != nil {
			return err
		}
		if set == nil || set.Len() == 0 {
			return protocol.MakeIntReply(0)
		}
		sets[i-1] = &set
	}
	intersect := Set.Intersect(sets...)
	db.Remove(dest)
	db.PutEntity(dest, &database.DataEntity{Data: intersect})
	db.addAof(utils.ToCmdLine3("sinterstore", args...))
	return protocol.MakeIntReply(int64(intersect.Len()))
}

func execSUnion(db *DB, args [][]byte) redis.Reply {
	n := len(args)
	sets := make([]*Set.Set, n)
	for i := 0; i < n; i++ {
		key := string(args[i])
		set, err := db.getAsSet(key)
		if err != nil {
			return err
		}
		if set == nil || set.Len() == 0 {
			return protocol.MakeEmptyMultiBulkReply()
		}
		sets[i] = &set
	}
	union := Set.Union(sets...)
	return setReply(&union)
}

func execSUnionStore(db *DB, args [][]byte) redis.Reply {
	dest := string(args[0])
	n := len(args) - 1
	sets := make([]*Set.Set, n)
	for i := 1; i <= n; i++ {
		key := string(args[i])
		set, err := db.getAsSet(key)
		if err != nil {
			return err
		}
		if set == nil {
			return protocol.MakeIntReply(0)
		}
		sets[i-1] = &set
	}
	union := Set.Union(sets...)
	db.Remove(dest)
	if union.Len() == 0 {
		return protocol.MakeIntReply(0)
	}
	db.PutEntity(dest, &database.DataEntity{Data: union})
	db.addAof(utils.ToCmdLine3("sunionstore", args...))
	return protocol.MakeIntReply(int64(union.Len()))
}

func execSDiff(db *DB, args [][]byte) redis.Reply {
	n := len(args)
	sets := make([]*Set.Set, n)
	for i := 0; i < n; i++ {
		key := string(args[i])
		set, err := db.getAsSet(key)
		if err != nil {
			return err
		}
		sets[i] = &set
	}
	diff := Set.Diff(sets...)
	return setReply(&diff)
}

func execSDiffStore(db *DB, args [][]byte) redis.Reply {
	dest := string(args[0])
	n := len(args) - 1
	sets := make([]*Set.Set, n)
	for i := 1; i <= n; i++ {
		key := string(args[i])
		set, err := db.getAsSet(key)
		if err != nil {
			return err
		}
		if set == nil {
			return protocol.MakeIntReply(0)
		}
		sets[i-1] = &set
	}
	diff := Set.Diff(sets...)
	db.Remove(dest)
	if diff.Len() == 0 {
		return protocol.MakeIntReply(0)
	}
	db.PutEntity(dest, &database.DataEntity{Data: diff})
	db.addAof(utils.ToCmdLine3("sdiffstore", args...))
	return protocol.MakeIntReply(int64(diff.Len()))
}

func execSRandMember(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	set, err := db.getAsSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	count := 1
	if len(args) > 1 {
		count64, err2 := strconv.ParseInt(string(args[1]), 10, 64)
		if err2 != nil || count64 <= 0 {
			return protocol.MakeErrReply("ERR value is out of range, must be positive")
		}
		count = int(count64)
	}
	n := set.Len()
	if count > n {
		count = n
	}
	if count == 0 {
		return protocol.MakeEmptyMultiBulkReply()
	}
	slice := make([][]byte, count)
	members := set.RandomDistinctMembers(count)
	for i := 0; i < count; i++ {
		slice[i] = []byte(members[i])
	}
	return protocol.MakeMultiBulkReply(slice)
}

func init() {
	RegisterCommand("SADD", execSAdd, writeFirstKey, undoSetChange, -3, flagWrite)
	RegisterCommand("SISMEMBER", execSIsMember, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("SREM", execSRem, writeFirstKey, undoSetChange, -3, flagWrite)
	RegisterCommand("SPOP", execSPop, writeFirstKey, undoSetChange, -2, flagWrite)
	RegisterCommand("SCARD", execSCard, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("SMEMBERS", execSMembers, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("SINTER", execSInter, prepareSetCalculate, nil, -2, flagReadOnly)
	RegisterCommand("SINTERSTORE", execSInterStore, prepareSetCalculateStore, rollbackFirstKey, -3, flagWrite)
	RegisterCommand("SUNION", execSUnion, prepareSetCalculate, nil, -2, flagReadOnly)
	RegisterCommand("SUNIONSTORE", execSUnionStore, prepareSetCalculateStore, rollbackFirstKey, -3, flagWrite)
	RegisterCommand("SDIFF", execSDiff, prepareSetCalculate, nil, -2, flagReadOnly)
	RegisterCommand("SDIFFSTORE", execSDiffStore, prepareSetCalculateStore, rollbackFirstKey, -3, flagWrite)
	RegisterCommand("SRANDMEMBER", execSRandMember, readFirstKey, nil, -2, flagReadOnly)
}
