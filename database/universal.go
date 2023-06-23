package database

import (
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
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
	return protocol.MakeOkReply()
}

func init() {
	RegisterCommand("Del", execDel, writeAllKeys, undoDel, -2, flagWrite)
	RegisterCommand("TTL", execTTL, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("Keys", execKeys, noPrepare, nil, 2, flagReadOnly)
}
