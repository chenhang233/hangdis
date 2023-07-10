package database

import (
	"hangdis/aof"
	"hangdis/config"
	"hangdis/interface/database"
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
		singleDB.addAof = func(line CmdLine) {
			if config.Properties.AppendOnly { // config may be changed during runtime
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
	// ----
	return nil
}
