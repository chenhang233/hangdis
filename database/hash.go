package database

import (
	Dict "hangdis/datastruct/dict"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
)

func (db *DB) getAsDict(key string) (Dict.Dict, redis.ErrorReply) {
	entity, exist := db.GetEntity(key)
	if !exist {
		return nil, nil
	}
	d, ok := entity.Data.(Dict.Dict)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return d, nil
}

func (db *DB) getOrInitDict(key string) (Dict.Dict, bool, redis.ErrorReply) {
	dict, err := db.getAsDict(key)
	if err != nil {
		return nil, false, &protocol.WrongTypeErrReply{}
	}
	isNew := false
	if dict == nil {
		dict = Dict.MakeInstanceDict()
		db.PutEntity(key, &database.DataEntity{Data: dict})
		isNew = true
		return dict, isNew, nil
	}
	return dict, isNew, nil
}

func undoHSet(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	args = args[1:]
	n := len(args)
	fields := make([]string, n/2)
	for i, j := 0, 0; i < n; {
		fields[j] = string(args[i])
		i += 2
		j++
	}
	return rollbackHashFields(db, key, fields...)
}

func undoHDel(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	fields := make([]string, len(args)-1)
	fieldArgs := args[1:]
	for i, v := range fieldArgs {
		fields[i] = string(v)
	}
	return rollbackHashFields(db, key, fields...)
}

func execHSet(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	dict, _, err := db.getOrInitDict(key)
	if err != nil {
		return err
	}
	args = args[1:]
	n := len(args)
	res := 0
	for i, j := 0, 1; i < n; {
		k := string(args[i])
		v := args[j]
		res += dict.Put(k, v)
		i += 2
		j += 2
	}
	db.addAof(utils.ToCmdLine3("hset", args...))
	return protocol.MakeIntReply(int64(res))
}

func execHSetNX(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	field := string(args[1])
	dict, _, err := db.getOrInitDict(key)
	if err != nil {
		return err
	}
	result := dict.PutIfAbsent(field, args[2])
	if result > 0 {
		db.addAof(utils.ToCmdLine3("hsetnx", args...))
	}
	return protocol.MakeIntReply(int64(result))
}

func execHGet(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	field := string(args[1])
	dict, err := db.getAsDict(key)
	if err != nil {
		return err
	}
	if dict == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	val, exists := dict.Get(field)
	if !exists {
		return protocol.MakeEmptyMultiBulkReply()
	}
	return protocol.MakeBulkReply(val.([]byte))
}

func execHExists(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	field := string(args[1])
	dict, err := db.getAsDict(key)
	if err != nil {
		return err
	}
	if dict == nil {
		return protocol.MakeIntReply(0)
	}
	_, exists := dict.Get(field)
	if !exists {
		return protocol.MakeIntReply(0)
	}
	return protocol.MakeIntReply(1)
}

func execHDel(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	fields := args[1:]
	dict, err := db.getAsDict(key)
	if err != nil {
		return err
	}
	if dict == nil {
		return protocol.MakeIntReply(0)
	}
	deleted := 0
	for _, field := range fields {
		deleted += dict.Remove(string(field))
	}
	if dict.Len() == 0 {
		db.Remove(key)
	}
	if deleted > 0 {
		db.addAof(utils.ToCmdLine3("hdel", args...))
	}
	return protocol.MakeIntReply(int64(deleted))
}

func init() {
	RegisterCommand("HSET", execHSet, writeFirstKey, undoHSet, -4, flagWrite).addParity(even)
	RegisterCommand("HSETNX", execHSetNX, writeFirstKey, undoHSet, 4, flagWrite)
	RegisterCommand("HGET", execHGet, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("HEXISTS", execHExists, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("HDEL", execHDel, writeFirstKey, undoHDel, -3, flagWrite)
	//registerCommand("HLen", execHLen, readFirstKey, nil, 2, flagReadOnly)
	//registerCommand("HStrlen", execHStrlen, readFirstKey, nil, 3, flagReadOnly)
	//registerCommand("HMSet", execHMSet, writeFirstKey, undoHMSet, -4, flagWrite)
	//registerCommand("HMGet", execHMGet, readFirstKey, nil, -3, flagReadOnly)
	//registerCommand("HKeys", execHKeys, readFirstKey, nil, 2, flagReadOnly)
	//registerCommand("HVals", execHVals, readFirstKey, nil, 2, flagReadOnly)
	//registerCommand("HGetAll", execHGetAll, readFirstKey, nil, 2, flagReadOnly)
	//registerCommand("HIncrBy", execHIncrBy, writeFirstKey, undoHIncr, 4, flagWrite)
	//registerCommand("HIncrByFloat", execHIncrByFloat, writeFirstKey, undoHIncr, 4, flagWrite)
	//registerCommand("HRandField", execHRandField, readFirstKey, nil, -2, flagReadOnly)
}
