package database

import (
	SortedSet "hangdis/datastruct/sortedset"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
)

func (db *DB) getAsSortedSet(key string) (*SortedSet.SortedSet, redis.ErrorReply) {
	entity, exists := db.GetEntity(key)
	if !exists {
		return nil, nil
	}
	sortedSet, ok := entity.Data.(*SortedSet.SortedSet)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return sortedSet, nil
}

func (db *DB) getOrInitSortedSet(key string) (*SortedSet.SortedSet, bool, redis.ErrorReply) {
	set, err := db.getAsSortedSet(key)
	if err != nil {
		return nil, false, err
	}
	isNew := false
	if set == nil {
		*set = SortedSet.SimpleMake()
		db.PutEntity(key, &database.DataEntity{Data: set})
		isNew = true
	}
	return set, isNew, nil
}
