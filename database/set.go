package database

import (
	Set "hangdis/datastruct/set"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
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

func init() {
	RegisterCommand("SADD", execSAdd, writeFirstKey, undoSetChange, -3, flagWrite)
	RegisterCommand("SISMEMBER", execSIsMember, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("SREM", execSRem, writeFirstKey, undoSetChange, -3, flagWrite)
}
