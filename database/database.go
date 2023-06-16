package database

import (
	"hangdis/datastruct/dict"
	"hangdis/interface/redis"
)

type DB struct {
	index      int
	data       dict.Dict
	ttlMap     dict.Dict
	versionMap dict.Dict
}

type ExecFunc func(db *DB, args [][]byte) redis.Reply

type PreFunc func(args [][]byte) ([]string, []string)

type CmdLine = [][]byte

type UndoFunc func(db *DB, args [][]byte) []CmdLine

func makeDB() *DB {
	return &DB{}
}
