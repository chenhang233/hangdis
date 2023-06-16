package database

import (
	"hangdis/datastruct/dict"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"time"
)

const (
	dataDictSize = 1 << 16
	ttlDictSize  = 1 << 10
	lockerSize   = 1024
)

type DB struct {
	index      int
	data       dict.Dict
	ttlMap     dict.Dict
	versionMap dict.Dict
}

type ExecFunc func(db *DB, args [][]byte) redis.Reply

type SysExecFunc func(db redis.Connection, args [][]byte) redis.Reply

type PreFunc func(args [][]byte) ([]string, []string)

type CmdLine = [][]byte

type UndoFunc func(db *DB, args [][]byte) []CmdLine

func makeDB() *DB {
	return &DB{
		index: 0,
		data:  dict.MakeConcurrent(dataDictSize),
	}
}

func (db *DB) AfterClientClose(c redis.Connection) {

}

func (db *DB) Close() {

}
func (db *DB) Exec(c redis.Connection, cmdLine [][]byte) redis.Reply {
	return nil
}

func (db *DB) execNormalCommand(cmdLine [][]byte) redis.Reply {
	return nil
}

func (db *DB) ExecWithLock(conn redis.Connection, cmdLine [][]byte) redis.Reply {
	return nil
}
func (db *DB) ExecMulti(conn redis.Connection, watching map[string]uint32, cmdLines []CmdLine) redis.Reply {
	return nil
}
func (db *DB) GetUndoLogs(dbIndex int, cmdLine [][]byte) []CmdLine {
	return nil
}
func (db *DB) ForEach(dbIndex int, cb func(key string, data *database.DataEntity, expiration *time.Time) bool) {

}
func (db *DB) RWLocks(dbIndex int, writeKeys []string, readKeys []string) {

}
func (db *DB) RWUnLocks(dbIndex int, writeKeys []string, readKeys []string) {

}
func (db *DB) GetDBSize(dbIndex int) (int, int) {
	return 0, 0
}
