package database

import (
	List "hangdis/datastruct/list"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
)

func (db *DB) getAsList(key string) (List.List, redis.ErrorReply) {
	entity, exist := db.GetEntity(key)
	if !exist {
		return nil, nil
	}
	list, ok := entity.Data.(List.List)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return list, nil
}

func (db *DB) getOrInitList(key string) (List.List, bool, redis.ErrorReply) {
	list, err := db.getAsList(key)
	if err != nil {
		return nil, false, err
	}
	isNew := false
	if list == nil {
		list = List.NewQuickList()
		db.PutEntity(key, &database.DataEntity{Data: list})
		isNew = true
	}
	return list, isNew, nil
}

func undoLPush(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	count := len(args) - 1
	cmdLines := make([]CmdLine, 0, count)
	for i := 0; i < count; i++ {
		cmdLines = append(cmdLines, utils.ToCmdLine("LPOP", key))
	}
	return cmdLines
}

func undoRPush(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	count := len(args) - 1
	cmdLines := make([]CmdLine, 0, count)
	for i := 0; i < count; i++ {
		cmdLines = append(cmdLines, utils.ToCmdLine("RPOP", key))
	}
	return cmdLines
}

func execLPush(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	values := args[1:]
	list, _, err := db.getOrInitList(key)
	if err != nil {
		return err
	}
	for _, value := range values {
		list.Insert(0, value)
	}
	db.addAof(utils.ToCmdLine3("lpush", args...))
	return protocol.MakeIntReply(int64(list.Len()))
}

func execLPushX(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	values := args[1:]
	list, err := db.getAsList(key)
	if err != nil {
		return err
	}
	if list == nil {
		return protocol.MakeIntReply(0)
	}
	for _, value := range values {
		list.Insert(0, value)
	}
	db.addAof(utils.ToCmdLine3("lpushx", args...))
	return protocol.MakeIntReply(int64(list.Len()))
}

func execRPush(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	values := args[1:]
	list, _, err := db.getOrInitList(key)
	if err != nil {
		return err
	}
	for _, value := range values {
		list.Add(value)
	}
	db.addAof(utils.ToCmdLine3("rpush", args...))
	return protocol.MakeIntReply(int64(list.Len()))
}

func execRPushX(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	values := args[1:]
	list, err := db.getAsList(key)
	if err != nil {
		return err
	}
	if list == nil {
		return protocol.MakeIntReply(0)
	}
	for _, value := range values {
		list.Add(value)
	}
	db.addAof(utils.ToCmdLine3("rpushx", args...))
	return protocol.MakeIntReply(int64(list.Len()))
}

func init() {
	RegisterCommand("LPush", execLPush, writeFirstKey, undoLPush, -3, flagWrite)
	RegisterCommand("LPushX", execLPushX, writeFirstKey, undoLPush, -3, flagWrite)
	RegisterCommand("RPush", execRPush, writeFirstKey, undoRPush, -3, flagWrite)
	RegisterCommand("RPushX", execRPushX, writeFirstKey, undoRPush, -3, flagWrite)
}
