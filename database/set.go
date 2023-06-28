package database

import (
	Set "hangdis/datastruct/set"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
)

func (db *DB) getAsSet(key string) (Set.Set, redis.ErrorReply) {
	entity, exist := db.GetEntity(key)
	if !exist {
		return nil, nil
	}
	set, ok := entity.Data.(Set.Set)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return set, nil
}

func (db *DB) getOrInitSet(key string) (Set.Set, bool, redis.ErrorReply) {
	set, err := db.getAsSet(key)
	if err != nil {
		return nil, false, err
	}
	isNew := false
	if set == nil {
		set = Set.Make()
		db.PutEntity(key, &database.DataEntity{Data: set})
		isNew = true
	}
	return set, isNew, nil
}
