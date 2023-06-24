package database

import (
	"hangdis/aof"
	"hangdis/datastruct/bitmap"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"math/bits"
	"strconv"
	"strings"
	"time"
)

const (
	upsertPolicy  = iota // default
	insertPolicy         // set nx
	updatePolicy         // set xx
	greaterExpiry        // set GL
	lessExpiry           // set LT
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

func prepareMGet(args [][]byte) ([]string, []string) {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	return nil, keys
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

func execMSetNX(db *DB, args [][]byte) redis.Reply {
	n := len(args)
	if n%2 != 0 {
		return protocol.MakeSyntaxErrReply()
	}
	size := n / 2
	var reply redis.Reply
	reply = protocol.MakeIntReply(1)
	for i := 0; i < size; i++ {
		key := string(args[i*2])
		val := args[i*2+1]
		entity := &database.DataEntity{Data: val}
		absent := db.PutIfAbsent(key, entity)
		if absent == 0 {
			reply = protocol.MakeIntReply(0)
		}
	}
	db.addAof(utils.ToCmdLine3("msetnx", args...))
	return reply
}

func execGetEX(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	ttl := unlimitedTTL
	bys, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if bys == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	for i := 1; i < len(args); i++ {
		arg := strings.ToUpper(string(args[i]))
		if arg == "EX" || arg == "PX" {
			if ttl != unlimitedTTL {
				return protocol.MakeSyntaxErrReply()
			}
			if i+1 > len(args) {
				return &protocol.SyntaxErrReply{}
			}
			num, err := strconv.ParseInt(string(args[i+1]), 10, 64)
			if err != nil {
				return protocol.MakeSyntaxErrReply()
			}
			if num <= 0 {
				return protocol.MakeErrReply("ERR invalid expire time in getex")
			}
			if arg == "EX" {
				num *= 1000
			}
			ttl = num
			i++
		} else if arg == "PERSIST" {
			if ttl != unlimitedTTL {
				return protocol.MakeSyntaxErrReply()
			}
			if i+1 > len(args) {
				return &protocol.SyntaxErrReply{}
			}
		}
	}
	if len(args) > 1 {
		if ttl != unlimitedTTL {
			expireTime := time.Now().Add(time.Duration(ttl) * time.Millisecond)
			db.Expire(key, expireTime)
			db.addAof(aof.MakeExpireCmd(key, expireTime).Args)
		} else {
			db.Persist(key)
			db.addAof(utils.ToCmdLine3("persist", args[0]))
		}
	}

	return protocol.MakeBulkReply(bys)
}

func execMGet(db *DB, args [][]byte) redis.Reply {
	n := len(args)
	bys := make([][]byte, n)
	for i := 0; i < n; i++ {
		asString, err := db.getAsString(string(args[i]))
		if err != nil {
			return err
		}
		bys[i] = asString
	}
	return protocol.MakeMultiBulkReply(bys)
}

func execGetSet(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	value := args[1]
	old, err := db.getAsString(key)
	if err != nil {
		return err
	}
	db.PutEntity(key, &database.DataEntity{Data: value})
	db.Persist(key)
	db.addAof(utils.ToCmdLine3("getset", args...))
	if old == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	return protocol.MakeBulkReply(old)
}

func execGetDel(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	val, err := db.getAsString(key)
	if err != nil {
		return err
	}
	db.Remove(key)
	db.addAof(utils.ToCmdLine3("del", args...))
	return protocol.MakeBulkReply(val)
}

func execIncr(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	val, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if val != nil {
		num, err2 := strconv.ParseInt(string(val), 10, 64)
		if err2 != nil {
			return protocol.MakeErrReply("ERR value is not an integer or out of range")
		}
		num++
		db.PutEntity(key, &database.DataEntity{Data: []byte(strconv.FormatInt(num, 10))})
		db.addAof(utils.ToCmdLine3("incr", args...))
		return protocol.MakeIntReply(num)
	}
	db.PutEntity(key, &database.DataEntity{Data: []byte("1")})
	db.addAof(utils.ToCmdLine3("incr", []byte("1")))
	return protocol.MakeIntReply(1)
}

func execIncrBy(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	add, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	num, _ := db.getAsString(key)
	if num != nil {
		i, err := strconv.ParseInt(string(num), 10, 64)
		if err != nil {
			return protocol.MakeErrReply("ERR value is not an integer or out of range")
		}
		i += add
		formatInt := strconv.FormatInt(i, 10)
		db.PutEntity(key, &database.DataEntity{Data: []byte(formatInt)})
		db.addAof(utils.ToCmdLine3("incrby", args...))
		return protocol.MakeIntReply(i)
	}
	db.PutEntity(key, &database.DataEntity{Data: args[1]})
	db.addAof(utils.ToCmdLine3("incrby", args...))
	return protocol.MakeIntReply(add)
}

func execIncrByFloat(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	add := string(args[1])
	float, err := strconv.ParseFloat(add, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not a valid float")
	}
	val, _ := db.getAsString(key)
	if val != nil {
		old, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return protocol.MakeErrReply("ERR value is not a valid float")
		}
		old += float
		next := []byte(strconv.FormatFloat(old, 'f', -1, 64))
		db.PutEntity(key, &database.DataEntity{Data: next})
		db.addAof(utils.ToCmdLine3("incrbyflaot", args...))
		return protocol.MakeBulkReply(next)
	}
	db.PutEntity(key, &database.DataEntity{Data: args[1]})
	db.addAof(utils.ToCmdLine3("incrbyflaot", args...))
	return protocol.MakeBulkReply(args[1])
}

func execDecr(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	num, _ := db.getAsString(key)
	if num != nil {
		add, err := strconv.ParseInt(string(num), 10, 64)
		if err != nil {
			return protocol.MakeErrReply("ERR value is not an integer or out of range")
		}
		add--
		entity := &database.DataEntity{
			Data: []byte(strconv.FormatInt(add, 10)),
		}
		db.PutEntity(key, entity)
		db.addAof(utils.ToCmdLine3("decr", args...))
		return protocol.MakeIntReply(add)
	}
	entity := &database.DataEntity{
		Data: []byte("-1"),
	}
	db.PutEntity(key, entity)
	db.addAof(utils.ToCmdLine3("decr", args...))
	return protocol.MakeIntReply(-1)
}

func execDecrBy(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	value := string(args[1])
	float, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not a valid float")
	}
	val, _ := db.getAsString(key)
	if val != nil {
		old, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return protocol.MakeErrReply("ERR value is not a valid float")
		}
		old -= float
		next := []byte(strconv.FormatFloat(old, 'f', -1, 64))
		db.PutEntity(key, &database.DataEntity{Data: next})
		db.addAof(utils.ToCmdLine3("decrby", args...))
		return protocol.MakeBulkReply(next)
	}
	db.PutEntity(key, &database.DataEntity{Data: args[1]})
	db.addAof(utils.ToCmdLine3("decrby", args...))
	return protocol.MakeBulkReply(args[1])
}

func execStrLen(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	val, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if val == nil {
		return protocol.MakeIntReply(0)
	}
	return protocol.MakeIntReply(int64(len(val)))
}

func execAppend(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	old, err := db.getAsString(key)
	if err != nil {
		return err
	}
	old = append(old, args[1]...)
	db.PutEntity(key, &database.DataEntity{Data: old})
	db.addAof(utils.ToCmdLine3("append", args...))
	return protocol.MakeIntReply(int64(len(old)))
}

func execSetRange(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	offset, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply(err.Error())
	}
	old, err2 := db.getAsString(key)
	if err2 != nil {
		return err2
	}
	n1 := len(old)
	n2 := int(offset)
	n3 := len(args[2])
	if n1-1 < n2 {
		apd := make([]byte, n2-n1)
		old = append(old, apd...)
		n1 += len(apd)
	}
	for i, j := n2, 0; i < n2+n3; i++ {
		if i >= n1 {
			old = append(old, args[2][j])
		} else {
			old[i] = args[2][j]
		}
		j++
	}
	db.PutEntity(key, &database.DataEntity{
		Data: old,
	})
	db.addAof(utils.ToCmdLine3("setrange", args...))
	return protocol.MakeIntReply(int64(len(old)))
}

func execGetRange(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	startIdx, err2 := strconv.ParseInt(string(args[1]), 10, 64)
	if err2 != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	endIdx, err2 := strconv.ParseInt(string(args[2]), 10, 64)
	if err2 != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	bs, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if bs == nil {
		return protocol.MakeEmptyMultiBulkReply()
	}
	bytesLen := int64(len(bs))
	beg, end := utils.ConvertRange(startIdx, endIdx, bytesLen)
	if beg < 0 {
		return protocol.MakeEmptyMultiBulkReply()
	}
	return protocol.MakeBulkReply(bs[beg:end])
}
func execSetBit(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	offset, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR bit offset is not an integer or out of range")
	}
	value := string(args[2])
	var v byte
	if value == "0" {
		v = 0
	} else if value == "1" {
		v = 1
	} else {
		return protocol.MakeErrReply("ERR bit is not an integer or out of range")
	}
	old, err2 := db.getAsString(key)
	if err2 != nil {
		return err2
	}
	bys := bitmap.FromBytes(old)
	get := bys.Get(offset)
	bys.Set(offset, v)
	db.PutEntity(key, &database.DataEntity{Data: bys.ToBytes()})
	db.addAof(utils.ToCmdLine3("setbit", args...))
	return protocol.MakeIntReply(int64(get))
}

func execGetBit(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	offset, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR bit offset is not an integer or out of range")
	}
	old, err2 := db.getAsString(key)
	if err2 != nil {
		return err2
	}
	if old == nil {
		return protocol.MakeIntReply(0)
	}
	bys := bitmap.FromBytes(old)
	get := bys.Get(offset)
	return protocol.MakeIntReply(int64(get))
}

func execBitCount(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	bys, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if bys == nil {
		return protocol.MakeIntReply(0)
	}
	byteMode := true
	if len(args) > 3 {
		mode := strings.ToLower(string(args[3]))
		if mode == "byte" {
			byteMode = true
		} else if mode == "bit" {
			byteMode = false
		}
	}
	var size int64
	bs := bitmap.FromBytes(bys)
	if byteMode {
		size = int64(len(*bs))
	} else {
		size = int64(bs.BitSize())
	}
	var beg, end int
	if len(args) > 1 {
		var err2 error
		var startIdx, endIdx int64
		startIdx, err2 = strconv.ParseInt(string(args[1]), 10, 64)
		if err2 != nil {
			return protocol.MakeErrReply("ERR value is not an integer or out of range")
		}
		endIdx, err2 = strconv.ParseInt(string(args[2]), 10, 64)
		if err2 != nil {
			return protocol.MakeErrReply("ERR value is not an integer or out of range")
		}
		beg, end = utils.ConvertRange(startIdx, endIdx, size)
		if beg < 0 {
			return protocol.MakeIntReply(0)
		}
	}
	var count int64
	if byteMode {
		bs.ForEachByte(beg, end, func(offset int64, val byte) bool {
			count += int64(bits.OnesCount8(val))
			return true
		})
	} else {
		bs.ForEachBit(int64(beg), int64(end), func(offset int64, val byte) bool {
			if val > 0 {
				count++
			}
			return true
		})
	}
	return protocol.MakeIntReply(count)
}

func execBitPos(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	bs, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if bs == nil {
		return protocol.MakeIntReply(-1)
	}
	valStr := string(args[1])
	var v byte
	if valStr == "1" {
		v = 1
	} else if valStr == "0" {
		v = 0
	} else {
		return protocol.MakeErrReply("ERR bit is not an integer or out of range")
	}
	byteMode := true
	if len(args) > 4 {
		mode := strings.ToLower(string(args[4]))
		if mode == "bit" {
			byteMode = false
		} else if mode == "byte" {
			byteMode = true
		} else {
			return protocol.MakeErrReply("ERR syntax error")
		}
	}
	var size int64
	bm := bitmap.FromBytes(bs)
	if byteMode {
		size = int64(len(*bm))
	} else {
		size = int64(bm.BitSize())
	}
	var beg, end int
	if len(args) > 2 {
		var err2 error
		var startIdx, endIdx int64
		startIdx, err2 = strconv.ParseInt(string(args[2]), 10, 64)
		if err2 != nil {
			return protocol.MakeErrReply("ERR value is not an integer or out of range")
		}
		endIdx, err2 = strconv.ParseInt(string(args[3]), 10, 64)
		if err2 != nil {
			return protocol.MakeErrReply("ERR value is not an integer or out of range")
		}
		beg, end = utils.ConvertRange(startIdx, endIdx, size)
		if beg < 0 {
			return protocol.MakeIntReply(0)
		}
	}
	if byteMode {
		beg *= 8
		end *= 8
	}
	var offset = int64(-1)
	bm.ForEachBit(int64(beg), int64(end), func(o int64, val byte) bool {
		if val == v {
			offset = o
			return false
		}
		return true
	})
	return protocol.MakeIntReply(offset)
}

func getRandomKey(db *DB, args [][]byte) redis.Reply {
	keys := db.data.RandomKeys(1)
	if len(keys) == 0 {
		return protocol.MakeEmptyMultiBulkReply()
	}
	var key []byte
	return protocol.MakeBulkReply(strconv.AppendQuote(key, keys[0]))
}

func init() {
	RegisterCommand("SET", execSet, writeFirstKey, rollbackFirstKey, -3, flagWrite)
	RegisterCommand("SETNx", execSetNX, writeFirstKey, rollbackFirstKey, -3, flagWrite)
	RegisterCommand("SETEx", execSetEX, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("PSetEX ", execPSetEX, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("MSet", execMSet, prepareMSet, undoMSet, -3, flagWrite)
	RegisterCommand("MSetNX", execMSetNX, prepareMSet, undoMSet, -3, flagWrite)
	RegisterCommand("GET", execGet, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("GetEX", execGetEX, writeFirstKey, rollbackFirstKey, -2, flagReadOnly)
	RegisterCommand("MGet", execMGet, prepareMGet, nil, -2, flagReadOnly)
	RegisterCommand("GetSet", execGetSet, writeFirstKey, rollbackFirstKey, 3, flagWrite)
	RegisterCommand("GetDel", execGetDel, writeFirstKey, rollbackFirstKey, 2, flagWrite)
	RegisterCommand("INCR", execIncr, writeFirstKey, rollbackFirstKey, 2, flagWrite)
	RegisterCommand("IncrBy", execIncrBy, writeFirstKey, rollbackFirstKey, 3, flagWrite)
	RegisterCommand("IncrByFloat", execIncrByFloat, writeFirstKey, rollbackFirstKey, 3, flagWrite)
	RegisterCommand("Decr", execDecr, writeFirstKey, rollbackFirstKey, 2, flagWrite)
	RegisterCommand("DecrBy", execDecrBy, writeFirstKey, rollbackFirstKey, 3, flagWrite)
	RegisterCommand("StrLen", execStrLen, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("Append", execAppend, writeFirstKey, rollbackFirstKey, 3, flagWrite)
	RegisterCommand("SetRange", execSetRange, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("GetRange", execGetRange, readFirstKey, nil, 4, flagReadOnly)
	RegisterCommand("SetBit", execSetBit, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("GetBit", execGetBit, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("BitCount", execBitCount, readFirstKey, nil, -2, flagReadOnly)
	RegisterCommand("BitPos", execBitPos, readFirstKey, nil, -3, flagReadOnly)
	RegisterCommand("RandomKey", getRandomKey, readAllKeys, nil, 1, flagReadOnly)
}
