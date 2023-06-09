package database

import (
	"fmt"
	"hangdis/datastruct/dict"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/pubsub"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/logs"
	"hangdis/utils/timewheel"
	"strings"
	"sync"
	"time"
)

const (
	dataDictSize = 1 << 16
	ttlDictSize  = 1 << 10
)

type DB struct {
	index      int
	data       *dict.ConcurrentDict
	ttlMap     *dict.ConcurrentDict
	versionMap *dict.ConcurrentDict
	addAof     func(CmdLine)
	lock       sync.RWMutex
}

type ExecFunc func(db *DB, args [][]byte) redis.Reply

type SysExecFunc func(db redis.Connection, args [][]byte) redis.Reply

type SubExecFunc func(hub *pubsub.Hub, c redis.Connection, args [][]byte) redis.Reply

type PreFunc func(args [][]byte) ([]string, []string)

type CmdLine = [][]byte

type UndoFunc func(db *DB, args [][]byte) []CmdLine

func makeBasicDB() *DB {
	return &DB{
		index:      0,
		data:       dict.MakeConcurrent(dataDictSize),
		ttlMap:     dict.MakeConcurrent(ttlDictSize),
		versionMap: dict.MakeConcurrent(dataDictSize),
		addAof:     func(line CmdLine) {},
	}
}

func makeDB() *DB {
	return &DB{
		index:      0,
		data:       dict.MakeConcurrent(dataDictSize),
		ttlMap:     dict.MakeConcurrent(ttlDictSize),
		versionMap: dict.MakeConcurrent(dataDictSize),
		addAof:     func(line CmdLine) {},
	}
}

func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	row, exists := db.data.GetWithLock(key)
	if !exists || db.IsExpired(key) {
		return nil, false
	}
	entity := row.(*database.DataEntity)
	return entity, true
}

func (db *DB) GetExpiredTime(key string) time.Time {
	val, exists := db.ttlMap.Get(key)
	if !exists {
		return time.Now()
	}
	t := val.(time.Time)
	return t
}

func (db *DB) IsExpired(key string) bool {
	val, exists := db.ttlMap.Get(key)
	if !exists {
		return false
	}
	t := val.(time.Time)
	after := time.Now().After(t)
	if after {
		db.Remove(key)
	}
	return after
}

func (db *DB) Expire(key string, expireTime time.Time) {
	db.ttlMap.Put(key, expireTime)
	name := utils.GetExpireTaskName(key)
	timewheel.At(expireTime, name, func() {
		keys := []string{key}
		db.RWLocks(keys, nil)
		defer db.RWUnLocks(keys, nil)
		val, exists := db.ttlMap.Get(key)
		if !exists {
			return
		}
		logs.LOG.Debug.Println(fmt.Sprintf("Expire callback key: %s", utils.Green(name)))
		expireTime := val.(time.Time)
		expired := time.Now().After(expireTime)
		if expired {
			db.Remove(key)
		}
	})
}

func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

func (db *DB) Remove(key string) {
	db.data.Remove(key)
	db.ttlMap.Remove(key)
	name := utils.GetExpireTaskName(key)
	timewheel.Cancel(name)
}

func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.GetWithLock(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

func (db *DB) Persist(key string) {
	db.ttlMap.Remove(key)
	name := utils.GetExpireTaskName(key)
	timewheel.Cancel(name)
}

func (db *DB) AfterClientClose(c redis.Connection) {

}

func (db *DB) Close() {

}
func (db *DB) Exec(c redis.Connection, cmdLine [][]byte) redis.Reply {
	cmdName := strings.ToLower(string(cmdLine[0]))
	if cmdName == "exec" {

	}
	if c != nil && c.InMultiState() {
		return EnqueueCmd(c, cmdLine)
	}
	return db.execNormalCommand(cmdLine)
}

func validateArity(arity int, cmdArgs [][]byte) bool {
	n := len(cmdArgs)
	if arity >= 0 {
		return arity == n
	}
	return -arity <= n
}

func validateParity(parity int, cmdArgs [][]byte) bool {
	if parity == -1 {
		return true
	}
	n := len(cmdArgs)
	if n%2 == 0 {
		if parity == odd {
			return false
		}
	} else {
		if parity == even {
			return false
		}
	}
	return true
}

func (db *DB) execNormalCommand(cmdLine [][]byte) redis.Reply {
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return protocol.MakeErrReply("ERR command not found")
	}
	if !validateArity(cmd.arity, cmdLine) {
		return protocol.MakeErrReply("Check error  wrong number of arguments")
	}
	if !validateParity(cmd.parity, cmdLine) {
		return protocol.MakeErrReply("Check error Parity")
	}
	prepare := cmd.prepare
	write, read := prepare(cmdLine[1:])
	db.addVersion(write...)
	db.RWLocks(write, read)
	defer db.RWUnLocks(write, read)
	executor := cmd.executor
	return executor(db, cmdLine[1:])
}

func EnqueueCmd(conn redis.Connection, cmdLine [][]byte) redis.Reply {
	return protocol.MakeEmptyMultiBulkReply()
}

func (db *DB) ForEach(cb func(key string, data *database.DataEntity, expiration *time.Time) bool) {
	db.data.ForEach(func(key string, raw interface{}) bool {
		entity, _ := raw.(*database.DataEntity)
		var expiration *time.Time
		rawExpireTime, ok := db.ttlMap.Get(key)
		if ok {
			expireTime, _ := rawExpireTime.(time.Time)
			expiration = &expireTime
		}

		return cb(key, entity, expiration)
	})
}
func (db *DB) RWLocks(writeKeys []string, readKeys []string) {
	//if len(readKeys) > 0 {
	//	db.lock.RLock()
	//	defer db.lock.RUnlock()
	//}
	//if len(writeKeys) > 0 {
	//	db.lock.Lock()
	//	defer db.lock.Unlock()
	//}
}
func (db *DB) RWUnLocks(writeKeys []string, readKeys []string) {

}

func (db *DB) addVersion(keys ...string) {
	for _, key := range keys {
		v := db.GetVersion(key)
		db.versionMap.Put(key, v+1)
	}
}

func (db *DB) GetVersion(key string) uint32 {
	val, exists := db.versionMap.Get(key)
	if !exists {
		return 0
	}
	return val.(uint32)
}

//func (db *DB) ExecWithLock(conn redis.Connection, cmdLine [][]byte) redis.Reply {
//	return nil
//}
//func (db *DB) ExecMulti(conn redis.Connection, watching map[string]uint32, cmdLines []CmdLine) redis.Reply {
//	return nil
//}
//func (db *DB) GetUndoLogs(cmdLine [][]byte) []CmdLine {
//	return nil
//}
