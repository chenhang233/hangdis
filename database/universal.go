package database

import (
	"hangdis/aof"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/wildcard"
	"strconv"
	"time"
)

func toTTLCmd(db *DB, key string) *protocol.MultiBulkReply {
	raw, exists := db.ttlMap.Get(key)
	if !exists {
		// has no TTL
		return protocol.MakeMultiBulkReply(utils.ToCmdLine("PERSIST", key))
	}
	expireTime, _ := raw.(time.Time)
	timestamp := strconv.FormatInt(expireTime.UnixNano()/1000/1000, 10)
	return protocol.MakeMultiBulkReply(utils.ToCmdLine("PEXPIREAT", key, timestamp))
}

func undoExpire(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	return []CmdLine{
		toTTLCmd(db, key).Args,
	}
}

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

func execExpire(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	num, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	_, exist := db.GetEntity(key)
	if !exist {
		return protocol.MakeIntReply(0)
	}
	policy := upsertPolicy
	if len(args) > 2 {
		op := string(args[2])
		if op == "NX" {
			policy = insertPolicy
		} else if op == "XX" {
			policy = updatePolicy
		} else if op == "GT" {
			policy = greaterExpiry
		} else if op == "LT" {
			policy = lessExpiry
		} else {
			return protocol.MakeSyntaxErrReply()
		}
	}
	expireAt := time.Now().Add(time.Duration(num) * time.Second)
	switch policy {
	case insertPolicy:
		if db.IsExpired(key) {
			return protocol.MakeIntReply(0)
		}
	case updatePolicy:
		if !db.IsExpired(key) {
			return protocol.MakeIntReply(0)
		}
	case greaterExpiry:
		if !expireAt.After(db.GetExpiredTime(key)) {
			return protocol.MakeIntReply(0)
		}
	case lessExpiry:
		if expireAt.After(db.GetExpiredTime(key)) {
			return protocol.MakeIntReply(0)
		}
	}
	db.Expire(key, expireAt)
	db.addAof(aof.MakeExpireCmd(key, expireAt).Args)
	return protocol.MakeIntReply(1)
}

func execExpireAt(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	num, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	_, exist := db.GetEntity(key)
	if !exist {
		return protocol.MakeIntReply(0)
	}
	policy := upsertPolicy
	if len(args) > 2 {
		op := string(args[2])
		if op == "NX" {
			policy = insertPolicy
		} else if op == "XX" {
			policy = updatePolicy
		} else if op == "GT" {
			policy = greaterExpiry
		} else if op == "LT" {
			policy = lessExpiry
		} else {
			return protocol.MakeSyntaxErrReply()
		}
	}
	expireAt := time.Unix(num, 0)
	switch policy {
	case insertPolicy:
		if db.IsExpired(key) {
			return protocol.MakeIntReply(0)
		}
	case updatePolicy:
		if !db.IsExpired(key) {
			return protocol.MakeIntReply(0)
		}
	case greaterExpiry:
		if !expireAt.After(db.GetExpiredTime(key)) {
			return protocol.MakeIntReply(0)
		}
	case lessExpiry:
		if expireAt.After(db.GetExpiredTime(key)) {
			return protocol.MakeIntReply(0)
		}
	}
	db.Expire(key, expireAt)
	db.addAof(aof.MakeExpireCmd(key, expireAt).Args)
	return protocol.MakeIntReply(1)
}

func execExpireTime(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	_, exist := db.GetEntity(key)
	if !exist {
		return protocol.MakeIntReply(-2)
	}
	val, exists := db.ttlMap.Get(key)
	if !exists {
		return protocol.MakeIntReply(-1)
	}
	t := val.(time.Time)
	return protocol.MakeIntReply(t.Unix())
}

func execPExpire(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	num, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	_, exist := db.GetEntity(key)
	if !exist {
		return protocol.MakeIntReply(0)
	}
	policy := upsertPolicy
	if len(args) > 2 {
		op := string(args[2])
		if op == "NX" {
			policy = insertPolicy
		} else if op == "XX" {
			policy = updatePolicy
		} else if op == "GT" {
			policy = greaterExpiry
		} else if op == "LT" {
			policy = lessExpiry
		} else {
			return protocol.MakeSyntaxErrReply()
		}
	}
	expireAt := time.Now().Add(time.Duration(num))
	switch policy {
	case insertPolicy:
		if db.IsExpired(key) {
			return protocol.MakeIntReply(0)
		}
	case updatePolicy:
		if !db.IsExpired(key) {
			return protocol.MakeIntReply(0)
		}
	case greaterExpiry:
		if !expireAt.After(db.GetExpiredTime(key)) {
			return protocol.MakeIntReply(0)
		}
	case lessExpiry:
		if expireAt.After(db.GetExpiredTime(key)) {
			return protocol.MakeIntReply(0)
		}
	}
	db.Expire(key, expireAt)
	db.addAof(aof.MakeExpireCmd(key, expireAt).Args)
	return protocol.MakeIntReply(1)
}

func execPExpireAt(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	num, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	_, exist := db.GetEntity(key)
	if !exist {
		return protocol.MakeIntReply(0)
	}
	policy := upsertPolicy
	if len(args) > 2 {
		op := string(args[2])
		if op == "NX" {
			policy = insertPolicy
		} else if op == "XX" {
			policy = updatePolicy
		} else if op == "GT" {
			policy = greaterExpiry
		} else if op == "LT" {
			policy = lessExpiry
		} else {
			return protocol.MakeSyntaxErrReply()
		}
	}
	expireAt := time.UnixMicro(num)
	switch policy {
	case insertPolicy:
		if db.IsExpired(key) {
			return protocol.MakeIntReply(0)
		}
	case updatePolicy:
		if !db.IsExpired(key) {
			return protocol.MakeIntReply(0)
		}
	case greaterExpiry:
		if !expireAt.After(db.GetExpiredTime(key)) {
			return protocol.MakeIntReply(0)
		}
	case lessExpiry:
		if expireAt.After(db.GetExpiredTime(key)) {
			return protocol.MakeIntReply(0)
		}
	}
	db.Expire(key, expireAt)
	db.addAof(aof.MakeExpireCmd(key, expireAt).Args)
	return protocol.MakeIntReply(1)
}

func init() {
	RegisterCommand("DEL", execDel, writeAllKeys, undoDel, -2, flagWrite)
	RegisterCommand("TTL", execTTL, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("KEYS", execKeys, noPrepare, nil, 2, flagReadOnly)
	RegisterCommand("EXPIRE", execExpire, writeFirstKey, undoExpire, -3, flagWrite)
	RegisterCommand("EXPIREAT", execExpireAt, writeFirstKey, undoExpire, -3, flagWrite)
	RegisterCommand("EXPIRETIME", execExpireTime, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("PEXPIRE", execPExpire, writeFirstKey, undoExpire, 3, flagWrite)
	RegisterCommand("PEXPIREAT", execPExpireAt, writeFirstKey, undoExpire, 3, flagWrite)
	//registerCommand("PTTL", execPTTL, readFirstKey, nil, 2, flagReadOnly)
	//registerCommand("Persist", execPersist, writeFirstKey, undoExpire, 2, flagWrite)
	//registerCommand("Exists", execExists, readAllKeys, nil, -2, flagReadOnly)
	//registerCommand("Type", execType, readFirstKey, nil, 2, flagReadOnly)
	//registerCommand("Rename", execRename, prepareRename, undoRename, 3, flagReadOnly)
	//registerCommand("RenameNx", execRenameNx, prepareRename, undoRename, 3, flagReadOnly)
}
