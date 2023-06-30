package database

import (
	SortedSet "hangdis/datastruct/sortedset"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"strconv"
)

func (db *DB) getAsSortedSet(key string) (SortedSet.SortedSet, redis.ErrorReply) {
	entity, exists := db.GetEntity(key)
	if !exists {
		return nil, nil
	}
	sortedSet, ok := entity.Data.(SortedSet.SortedSet)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return sortedSet, nil
}

func (db *DB) getOrInitSortedSet(key string) (SortedSet.SortedSet, bool, redis.ErrorReply) {
	set, err := db.getAsSortedSet(key)
	if err != nil {
		return nil, false, err
	}
	isNew := false
	if set == nil {
		set = SortedSet.SimpleMake()
		db.PutEntity(key, &database.DataEntity{Data: set})
		isNew = true
	}
	return set, isNew, nil
}

func undoZAdd(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	size := (len(args) - 1) / 2
	fields := make([]string, size)
	for i := 0; i < size; i++ {
		fields[i] = string(args[2*i+2])
	}
	return rollbackZSetFields(db, key, fields...)
}

func execZAdd(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	tp := string(args[1])
	policy := upsertPolicy
	if tp == "NX" {
		policy = insertPolicy
	} else if tp == "XX" {
		policy = updatePolicy
	}
	var n, j int
	if policy != upsertPolicy {
		n = (len(args) - 2) / 2
		j = 1
	} else {
		n = (len(args) - 1) / 2
		j = 0
	}
	elements := make([]*SortedSet.Element, n)
	for i := 0; i < n; i++ {
		scoreValue := string(args[i*2+1+j])
		score, err := strconv.ParseFloat(scoreValue, 64)
		if err != nil {
			return protocol.MakeErrReply("ERR value is not a valid float")
		}
		member := string(args[i*2+2+j])
		elements[i] = &SortedSet.Element{
			Member: member,
			Score:  score,
		}
	}
	set, _, err := db.getOrInitSortedSet(key)
	if err != nil {
		return err
	}
	i := 0
	for _, element := range elements {
		if policy != upsertPolicy {
			_, ok := set.Get(element.Member)
			if policy == insertPolicy && ok {
				continue
			}
			if policy == updatePolicy && !ok {
				continue
			}
		}
		if set.Add(element.Member, element.Score) {
			i++
		}
	}
	db.addAof(utils.ToCmdLine3("zadd", args...))
	return protocol.MakeIntReply(int64(i))
}

func execZScore(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	member := string(args[1])
	set, err := db.getAsSortedSet(key)
	if err != nil {
		return err
	}
	if set == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	element, ok := set.Get(member)
	if !ok {
		return protocol.MakeEmptyMultiBulkReply()
	}
	value := strconv.FormatFloat(element.Score, 'f', -1, 64)
	return protocol.MakeBulkReply([]byte(value))
}

func init() {
	RegisterCommand("ZADD", execZAdd, writeFirstKey, undoZAdd, -4, flagWrite)
	RegisterCommand("ZSCORE", execZScore, readFirstKey, nil, 3, flagReadOnly)
	//RegisterCommand("ZCard", execZCard, readFirstKey, nil, 2, flagReadOnly)
}
