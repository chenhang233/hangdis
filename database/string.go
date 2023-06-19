package database

import (
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
)

const (
	upsertPolicy = iota // default
	insertPolicy        // set nx
	updatePolicy        // set ex
)
const unlimitedTTL int64 = 0

func execSet(db *DB, args [][]byte) redis.Reply {
	return protocol.MakeEmptyMultiBulkReply()
}

func init() {
	RegisterCommand("SET", execSet, writeFirstKey, rollbackFirstKey, -3, flagWrite)
}
