package database

import (
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
)

func undoDel(db *DB, args [][]byte) []CmdLine {
	keys, _ := writeAllKeys(args)
	return rollbackGivenKeys(db, keys...)
}

func execDel(db *DB, args [][]byte) redis.Reply {
	return protocol.MakeOkReply()
}

func execTTL(db *DB, args [][]byte) redis.Reply {
	return protocol.MakeOkReply()
}

func execKeys(db *DB, args [][]byte) redis.Reply {
	return protocol.MakeOkReply()
}

func init() {
	RegisterCommand("Del", execDel, writeAllKeys, undoDel, -2, flagWrite)
	RegisterCommand("TTL", execTTL, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("Keys", execKeys, noPrepare, nil, 2, flagReadOnly)
}
