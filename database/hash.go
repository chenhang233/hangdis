package database

import (
	Dict "hangdis/datastruct/dict"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
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
	for i, j := 0, 0; i < n; i++ {
		fields[j] = string(args[i])
		i++
		j++
	}
	return rollbackHashFields(db, key, fields...)
}

func execHSet(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	args = args[1:]
}

func init() {
	RegisterCommand("HSet", execHSet, writeFirstKey, undoHSet, -4, flagWrite).addParity(even)
	//registerCommand("HSetNX", execHSetNX, writeFirstKey, undoHSet, 4, flagWrite)
	//registerCommand("HGet", execHGet, readFirstKey, nil, 3, flagReadOnly)
	//registerCommand("HExists", execHExists, readFirstKey, nil, 3, flagReadOnly)
}
