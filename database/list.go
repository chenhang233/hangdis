package database

import (
	List "hangdis/datastruct/list"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
)

func (db *DB) getAsList(key string) (List.List, redis.ErrorReply) {
	entity, exist := db.GetEntity(key)
	if !exist {
		return nil, nil
	}
	list, ok := entity.Data.(List.List)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return list, nil
}

func (db *DB) getOrInitList(key string) (List.List, bool, redis.ErrorReply) {
	list, err := db.getAsList(key)
	if err != nil {
		return nil, false, err
	}
	isNew := false
	if list == nil {
		list = List.NewQuickList()
		db.PutEntity(key, &database.DataEntity{Data: list})
		isNew = true
	}
	return list, isNew, nil
}
