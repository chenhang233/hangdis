package database

import (
	"hangdis/aof"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"strconv"
	"strings"
	"time"
)

const (
	upsertPolicy = iota // default
	insertPolicy        // set nx
	updatePolicy        // set xx
)
const unlimitedTTL int64 = 0

func execSet(db *DB, args [][]byte) redis.Reply {
	policy := upsertPolicy
	ttl := unlimitedTTL
	n := len(args)
	key := string(args[0])
	value := args[1]
	if n > 2 {
		for i := 2; i < n; i++ {
			arg := strings.ToUpper(string(args[i]))
			if arg == "NX" {
				if policy == updatePolicy {
					return protocol.MakeSyntaxErrReply()
				}
				policy = insertPolicy
			} else if arg == "XX" {
				if policy == insertPolicy {
					return protocol.MakeSyntaxErrReply()
				}
				policy = updatePolicy
			} else if arg == "EX" || arg == "PX" {
				if ttl != unlimitedTTL || i+1 >= n {
					return protocol.MakeSyntaxErrReply()
				}
				tv, err := strconv.ParseInt(string(arg[i+1]), 10, 64)
				if err != nil {
					return protocol.MakeSyntaxErrReply()
				}
				if tv <= 0 {
					return protocol.MakeErrReply("ERR invalid expire time in set")
				}
				if arg == "EX" {
					tv *= 1000
				}
				ttl = tv
			} else {
				return protocol.MakeSyntaxErrReply()
			}
		}
	}
	entity := &database.DataEntity{Data: value}
	var result int
	switch policy {
	case upsertPolicy:
		db.PutEntity(key, entity)
		result = 1
	case updatePolicy:
		result = db.PutIfExists(key, entity)
	case insertPolicy:
		result = db.PutIfAbsent(key, entity)
	}
	if result > 0 {
		if ttl != unlimitedTTL {
			expireTime := time.Now().Add(time.Duration(ttl) * time.Millisecond)
			db.Expire(key, expireTime)
			db.addAof(CmdLine{
				[]byte("SET"),
				args[0],
				args[1],
			})
			db.addAof(aof.MakeExpireCmd(key, expireTime).Args)
		} else {
			db.Persist(key)
			db.addAof(utils.ToCmdLine3("set", args...))
		}
		return protocol.MakeOkReply()
	}
	return protocol.MakeEmptyMultiBulkReply()
}

func (db *DB) getAsString(key string) ([]byte, *protocol.SyntaxErrReply) {
	entity, ok := db.GetEntity(key)
	if !ok {
		return nil, nil
	}
	bytes, ok := entity.Data.([]byte)
	if !ok {
		return nil, &protocol.SyntaxErrReply{}
	}
	return bytes, nil
}

func execGet(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	bys, err := db.getAsString(key)
	if err != nil {
		return err
	}
	return protocol.MakeBulkReply(bys)
}

func init() {
	RegisterCommand("SET", execSet, writeFirstKey, rollbackFirstKey, -3, flagWrite)
	RegisterCommand("GET", execGet, readFirstKey, nil, 2, flagReadOnly)
}
