package database

import "hangdis/datastruct/dict"

type DB struct {
	index      int
	data       dict.Dict
	ttlMap     dict.Dict
	versionMap dict.Dict
}

func makeDB() *DB {
	return &DB{}
}
