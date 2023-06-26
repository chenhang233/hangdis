package database

import (
	List "hangdis/datastruct/list"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"strconv"
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

var lPushCmd = []byte("LPUSH")

func undoLPop(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	list, errReply := db.getAsList(key)
	if errReply != nil {
		return nil
	}
	if list == nil || list.Len() == 0 {
		return nil
	}
	element, _ := list.Get(0).([]byte)
	return []CmdLine{
		{
			lPushCmd,
			args[0],
			element,
		},
	}
}

var rPushCmd = []byte("RPUSH")

func undoRPop(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	list, errReply := db.getAsList(key)
	if errReply != nil {
		return nil
	}
	if list == nil || list.Len() == 0 {
		return nil
	}
	element, _ := list.Get(list.Len() - 1).([]byte)
	return []CmdLine{
		{
			rPushCmd,
			args[0],
			element,
		},
	}
}

func undoRPopLPush(db *DB, args [][]byte) []CmdLine {
	sourceKey := string(args[0])
	list, errReply := db.getAsList(sourceKey)
	if errReply != nil {
		return nil
	}
	if list == nil || list.Len() == 0 {
		return nil
	}
	element, _ := list.Get(list.Len() - 1).([]byte)
	return []CmdLine{
		{
			rPushCmd,
			args[0],
			element,
		},
		{
			[]byte("LPOP"),
			args[1],
		},
	}
}

func undoLSet(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	index64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return nil
	}
	index := int(index64)
	list, errReply := db.getAsList(key)
	if errReply != nil {
		return nil
	}
	if list == nil {
		return nil
	}
	size := list.Len()
	if index >= size || index < -size {
		return nil
	} else if index < 0 {
		index = size + index
	}
	value, _ := list.Get(index).([]byte)
	return []CmdLine{
		{
			[]byte("LSET"),
			args[0],
			args[1],
			value,
		},
	}
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

func execLPop(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	list, err := db.getAsList(key)
	if err != nil {
		return err
	}
	if list == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	rn := 1
	if len(args) > 1 {
		n, err2 := strconv.ParseInt(string(args[1]), 10, 64)
		if err2 != nil {
			return protocol.MakeErrReply("ERR invalid count")
		}
		rn = int(n)
		if rn <= 0 {
			return protocol.MakeErrReply("ERR count less than 0")
		}
	}
	if rn == 1 {
		val := list.Remove(0).([]byte)
		if list.Len() == 0 {
			db.Remove(key)
		}
		db.addAof(utils.ToCmdLine3("lpop", args...))
		return protocol.MakeBulkReply(val)
	}
	res := make([][]byte, rn)
	for i := 0; i < rn; i++ {
		if list.Len() == 0 {
			break
		}
		val := list.Remove(0).([]byte)
		res[i] = val
	}
	db.addAof(utils.ToCmdLine3("lpop", args...))
	return protocol.MakeMultiBulkReply(res)
}

func execRPop(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	list, err := db.getAsList(key)
	if err != nil {
		return err
	}
	if list == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	rn := 1
	if len(args) > 1 {
		n, err2 := strconv.ParseInt(string(args[1]), 10, 64)
		if err2 != nil {
			return protocol.MakeErrReply("ERR invalid count")
		}
		rn = int(n)
		if rn <= 0 {
			return protocol.MakeErrReply("ERR count less than 0")
		}
	}
	if rn == 1 {
		val := list.RemoveLast().([]byte)
		if list.Len() == 0 {
			db.Remove(key)
		}
		db.addAof(utils.ToCmdLine3("rpop", args...))
		return protocol.MakeBulkReply(val)
	}
	res := make([][]byte, rn)
	for i := 0; i < rn; i++ {
		if list.Len() == 0 {
			break
		}
		val := list.RemoveLast().([]byte)
		res[i] = val
	}
	db.addAof(utils.ToCmdLine3("rpop", args...))
	return protocol.MakeMultiBulkReply(res)
}

func prepareRPopLPush(args [][]byte) ([]string, []string) {
	return []string{
		string(args[0]),
		string(args[1]),
	}, nil
}

func execRPopLPush(db *DB, args [][]byte) redis.Reply {
	sourceKey := string(args[0])
	destKey := string(args[1])
	list1, err := db.getAsList(sourceKey)
	if err != nil {
		return err
	}
	if list1 == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	list2, _, err := db.getOrInitList(destKey)
	if err != nil {
		return err
	}
	last := list1.RemoveLast().([]byte)
	list2.Insert(0, last)
	if list1.Len() == 0 {
		db.Remove(sourceKey)
	}
	db.addAof(utils.ToCmdLine3("rpoplpush", args...))
	return protocol.MakeBulkReply(last)
}

func execLRem(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	count64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	count := int(count64)
	value := args[2]
	list, err2 := db.getAsList(key)
	if err2 != nil {
		return err2
	}
	if list == nil {
		return protocol.MakeIntReply(0)
	}
	var removed int
	if count < 0 {
		removed = list.ReverseRemoveByVal(func(a any) bool {
			return utils.Equals(a, value)
		}, -count)
	} else if count > 0 {
		removed = list.RemoveByVal(func(a any) bool {
			return utils.Equals(a, value)
		}, count)
	} else {
		removed = list.RemoveAllByVal(func(a any) bool {
			return utils.Equals(a, value)
		})
	}
	if list.Len() == 0 {
		db.Remove(key)
	}
	if removed > 0 {
		db.addAof(utils.ToCmdLine3("lrem", args...))
	}
	return protocol.MakeIntReply(int64(removed))
}

func execLLen(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	list, err := db.getAsList(key)
	if err != nil {
		return err
	}
	if list == nil {
		return protocol.MakeIntReply(0)
	}
	return protocol.MakeIntReply(int64(list.Len()))
}

func execLIndex(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	index64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	index := int(index64)
	list, err2 := db.getAsList(key)
	if err2 != nil {
		return err2
	}
	if list == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	size := list.Len()
	if index >= size || index < -size {
		return protocol.MakeEmptyMultiBulkReply()
	} else if index < 0 {
		index = size + index
	}
	res := list.Get(index).([]byte)
	return protocol.MakeBulkReply(res)
}

func execLSet(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	index64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	index := int(index64)
	list, err2 := db.getAsList(key)
	if err2 != nil {
		return err2
	}
	if list == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	size := list.Len()
	if index >= size || index < -size {
		return protocol.MakeEmptyMultiBulkReply()
	} else if index < 0 {
		index = size + index
	}
	list.Set(index, args[2])
	db.addAof(utils.ToCmdLine3("lset", args...))
	return protocol.MakeOkReply()
}

func execLRange(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	start64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	stop64, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	list, err2 := db.getAsList(key)
	if err2 != nil {
		return err2
	}
	if list == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	size := list.Len()
	start, end := utils.ConvertRange(start64, stop64, int64(size))
	if start < 0 {
		return protocol.MakeEmptyMultiBulkReply()
	}
	slice := list.Range(start, end)
	res := make([][]byte, len(slice))
	for i, a := range slice {
		res[i] = a.([]byte)
	}
	if len(res) == 0 {
		return protocol.MakeEmptyMultiBulkReply()
	}
	return protocol.MakeMultiBulkReply(res)
}

func init() {
	RegisterCommand("LPUSH", execLPush, writeFirstKey, undoLPush, -3, flagWrite)
	RegisterCommand("LPUSHX", execLPushX, writeFirstKey, undoLPush, -3, flagWrite)
	RegisterCommand("RPUSH", execRPush, writeFirstKey, undoRPush, -3, flagWrite)
	RegisterCommand("RPUSHX", execRPushX, writeFirstKey, undoRPush, -3, flagWrite)
	RegisterCommand("LPOP", execLPop, writeFirstKey, undoLPop, -2, flagWrite)
	RegisterCommand("RPOP", execRPop, writeFirstKey, undoRPop, -2, flagWrite)
	RegisterCommand("RPOPLPUSH", execRPopLPush, prepareRPopLPush, undoRPopLPush, 3, flagWrite)
	RegisterCommand("LREM", execLRem, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("LLEN", execLLen, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("LINDEX", execLIndex, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("LSET", execLSet, writeFirstKey, undoLSet, 4, flagWrite)
	RegisterCommand("LRANGE", execLRange, readFirstKey, nil, 4, flagReadOnly)
	//registerCommand("LTrim", execLTrim, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	//registerCommand("LInsert", execLInsert, writeFirstKey, rollbackFirstKey, 5, flagWrite)
}
