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

func execSetNX(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{Data: value}
	ok := db.PutIfAbsent(key, entity)
	db.addAof(utils.ToCmdLine3("setnx", args...))
	return protocol.MakeIntReply(int64(ok))
}

func execSetEX(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	value := args[2]
	num, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeSyntaxErrReply()
	}
	if num <= 0 {
		return protocol.MakeErrReply("ERR invalid expire time in setex")
	}
	entity := &database.DataEntity{Data: value}
	db.PutEntity(key, entity)
	expireTime := time.Now().Add(time.Duration(num*1000) * time.Millisecond)
	db.Expire(key, expireTime)
	db.addAof(utils.ToCmdLine3("setex", args...))
	return &protocol.OkReply{}
}

func execPSetEX(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	value := args[2]
	num, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeSyntaxErrReply()
	}
	if num <= 0 {
		return protocol.MakeErrReply("ERR invalid expire time in psetex")
	}
	entity := &database.DataEntity{Data: value}
	db.PutEntity(key, entity)
	expireTime := time.Now().Add(time.Duration(num) * time.Millisecond)
	db.Expire(key, expireTime)
	db.addAof(utils.ToCmdLine3("psetex", args...))
	return &protocol.OkReply{}
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

func prepareMSet(args [][]byte) ([]string, []string) {
	size := len(args) / 2
	keys := make([]string, size)
	for i := 0; i < size; i++ {
		keys[i] = string(args[2*i])
	}
	return keys, nil
}

func undoMSet(db *DB, args [][]byte) []CmdLine {
	writeKeys, _ := prepareMSet(args)
	return rollbackGivenKeys(db, writeKeys...)
}

func execMSet(db *DB, args [][]byte) redis.Reply {
	n := len(args)
	if n%2 != 0 {
		return protocol.MakeSyntaxErrReply()
	}
	size := n / 2
	for i := 0; i < size; i++ {
		key := string(args[i*2])
		val := args[i*2+1]
		entity := &database.DataEntity{Data: val}
		db.PutEntity(key, entity)
	}
	db.addAof(utils.ToCmdLine3("mset", args...))
	return &protocol.OkReply{}
}

func init() {
	RegisterCommand("SET", execSet, writeFirstKey, rollbackFirstKey, -3, flagWrite)
	RegisterCommand("SETNx", execSetNX, writeFirstKey, rollbackFirstKey, -3, flagWrite)
	RegisterCommand("SETEx", execSetEX, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("PSetEX ", execPSetEX, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("MSet", execMSet, prepareMSet, undoMSet, -3, flagWrite)
	RegisterCommand("GET", execGet, readFirstKey, nil, 2, flagReadOnly)
}
