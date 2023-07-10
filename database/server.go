package database

import (
	"fmt"
	"hangdis/aof"
	"hangdis/config"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/pubsub"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/logs"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	dbSet     []*atomic.Value // *DB
	hub       *pubsub.Hub
	perSister *aof.PerSister
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func NewStandaloneServer() *Server {
	server := &Server{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	err := os.MkdirAll(config.GetTmpDir(), os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("create tmp dir failed: %v", err))
	}
	server.dbSet = make([]*atomic.Value, config.Properties.Databases)
	for i := range server.dbSet {
		db := makeDB()
		db.index = i
		holder := &atomic.Value{}
		holder.Store(db)
		server.dbSet[i] = holder
	}
	server.hub = pubsub.MakeHub()
	validAof := false
	if config.Properties.AppendOnly {
		validAof = fileExists(config.Properties.AppendFilename)
		aofHandler, err := NewPerSister(server,
			config.Properties.AppendFilename, true, config.Properties.AppendFsync)
		if err != nil {
			panic(err)
		}
		server.bindPerSister(aofHandler)
	}
	if config.Properties.RDBFilename != "" && !validAof {
		err := server.loadRdbFile()
		if err != nil {
			logs.LOG.Error.Println(err)
		}
	}
	return server
}

func (server *Server) Exec(c redis.Connection, cmdLine [][]byte) (result redis.Reply) {
	cmdName := strings.ToLower(string(cmdLine[0]))
	logs.LOG.Debug.Println(utils.Yellow(fmt.Sprintf("client info: %s  current execution command: %s", c.Name(), cmdName)))
	if !isAuthenticated(c, cmdName) {
		return protocol.MakeErrReply("Authentication required")
	}
	if sysCmd, ok := systemTable[cmdName]; ok {
		exec := sysCmd.executor
		return exec(c, cmdLine)
	}
	if p, ok := pubSubTable[cmdName]; ok {
		exec := p.executor
		return exec(server.hub, c, cmdLine[1:])
	}
	index := c.GetDBIndex()
	db, err := server.selectDB(index)
	if err != nil {
		logs.LOG.Error.Println(err)
		return err
	}
	return db.Exec(c, cmdLine)
}

func (server *Server) selectDB(dbIndex int) (*DB, *protocol.StandardErrReply) {
	if dbIndex >= len(server.dbSet) || dbIndex < 0 {
		return nil, protocol.MakeErrReply("ERR DB index is out of range")
	}
	return server.dbSet[dbIndex].Load().(*DB), nil
}

func isAuthenticated(c redis.Connection, cmdName string) bool {
	if config.Properties.Password == "" || cmdName == "auth" {
		return true
	}
	return c.GetPassword() == config.Properties.Password
}

func (server *Server) AfterClientClose(c redis.Connection) {
	pubsub.UnsubscribeAll(server.hub, c)
}

func (server *Server) Close() {
	if server.perSister != nil {
		//server.perSister.Close()
	}
}

func (server *Server) mustSelectDB(dbIndex int) *DB {
	selectedDB, err := server.selectDB(dbIndex)
	if err != nil {
		panic(err)
	}
	return selectedDB
}
func (server *Server) GetDBSize(dbIndex int) (int, int) {
	db := server.mustSelectDB(dbIndex)
	return db.data.Len(), db.ttlMap.Len()
}

func (server *Server) ExecWithLock(conn redis.Connection, cmdLine [][]byte) redis.Reply {
	return nil
}
func (server *Server) ExecMulti(conn redis.Connection, watching map[string]uint32, cmdLines []database.CmdLine) redis.Reply {
	return nil
}
func (server *Server) GetUndoLogs(dbIndex int, cmdLine [][]byte) []database.CmdLine {
	return nil
}
func (server *Server) ForEach(dbIndex int, cb func(key string, data *database.DataEntity, expiration *time.Time) bool) {
}
func (server *Server) RWLocks(dbIndex int, writeKeys []string, readKeys []string) {

}
func (server *Server) RWUnLocks(dbIndex int, writeKeys []string, readKeys []string) {

}
