package database

import (
	"fmt"
	"hangdis/aof"
	"hangdis/config"
	"hangdis/interface/database"
	"hangdis/utils/logs"
	"hangdis/utils/rdb"
	"os"
	"sync/atomic"
)

func NewPerSister(db database.DBEngine, filename string, load bool, fsync string) (*aof.PerSister, error) {
	return aof.NewPerSister(db, filename, load, fsync, func() database.DBEngine {
		return MakeAuxiliaryServer()
	})
}

func MakeAuxiliaryServer() *Server {
	mdb := &Server{}
	mdb.dbSet = make([]*atomic.Value, config.Properties.Databases)
	for i := range mdb.dbSet {
		holder := &atomic.Value{}
		holder.Store(makeBasicDB())
		mdb.dbSet[i] = holder
	}
	return mdb
}

func (server *Server) bindPerSister(aofHandler *aof.PerSister) {
	server.perSister = aofHandler
	for _, db := range server.dbSet {
		singleDB := db.Load().(*DB)
		logs.LOG.Info.Println(fmt.Sprintf("Database %d, Start listening to addAof", singleDB.index))
		singleDB.addAof = func(line CmdLine) {
			if config.Properties.AppendOnly {
				server.perSister.SaveCmdLine(singleDB.index, line)
			}
		}
	}
}

func (server *Server) loadRdbFile() error {
	rdbFile, err := os.Open(config.Properties.RDBFilename)
	if err != nil {
		return err
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	decoder := rdb.NewDecoder(rdbFile)
	err = server.LoadRDB(decoder)
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) LoadRDB(dec *rdb.Decoder) error {
	return dec.Parse(func(obj rdb.RedisObject) bool {
		db := server.mustSelectDB(obj.GetDBIndex())
		var entity *database.DataEntity
		switch obj.GetType() {
		case rdb.TypeString:
			entity = &database.DataEntity{
				Data: obj.V,
			}
		}
		if entity != nil {
			db.PutEntity(obj.GetKey(), entity)
			if obj.GetExpiration() != nil {
				db.Expire(obj.GetKey(), *obj.GetExpiration())
			}
			db.addAof(aof.EntityToCmd(obj.GetKey(), entity).Args)
		}
		return true
	})
}
