package database

import (
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/wildcard"
	"time"
)

func undoDel(db *DB, args [][]byte) []CmdLine {
	keys, _ := writeAllKeys(args)
	return rollbackGivenKeys(db, keys...)
}

func execDel(db *DB, args [][]byte) redis.Reply {
	keys, _ := writeAllKeys(args)
	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.addAof(utils.ToCmdLine3("del", args...))
	}
	return protocol.MakeIntReply(int64(deleted))
}

func execTTL(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	_, e := db.GetEntity(key)
	if !e {
		return protocol.MakeIntReply(-2)
	}
	val, exists := db.ttlMap.Get(key)
	if !exists {
		return protocol.MakeIntReply(-1)
	}
	t := val.(time.Time)
	ttl := t.Sub(time.Now())
	return protocol.MakeIntReply(int64(ttl / time.Second))
}

func execKeys(db *DB, args [][]byte) redis.Reply {
	pattern, err := wildcard.CompilePattern(string(args[0]))
	if err != nil {
		return protocol.MakeErrReply("ERR illegal wildcard")
	}
	keys := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			keys = append(keys, []byte(key))
			return true
		}
		return false
	})
	return protocol.MakeMultiBulkReply(keys)
}

func init() {
	RegisterCommand("DEL", execDel, writeAllKeys, undoDel, -2, flagWrite)
	RegisterCommand("TTL", execTTL, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("KEYS", execKeys, noPrepare, nil, 2, flagReadOnly)
	registerCommand("Expire", execExpire, writeFirstKey, undoExpire, 3, flagWrite)
	registerCommand("ExpireAt", execExpireAt, writeFirstKey, undoExpire, 3, flagWrite)
	registerCommand("ExpireTime", execExpireTime, readFirstKey, nil, 2, flagReadOnly)
	registerCommand("PExpire", execPExpire, writeFirstKey, undoExpire, 3, flagWrite)
}
