package database

import (
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
)

func Ping(c redis.Connection, args [][]byte) redis.Reply {
	if len(args) == 0 {
		return protocol.MakePongReply()
	} else if len(args) == 1 {
		return protocol.MakeErrReply(string(args[0]))
	} else {
		return protocol.MakeEmptyMultiBulkReply()
	}
}
