package database

import (
	"fmt"
	"hangdis/config"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/pubsub"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/logs"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	dbSet []*atomic.Value // *DB
	hub   *pubsub.Hub
	//persister *aof.Persister
	//role int32
	//slaveStatus  *slaveStatus
	//masterStatus *masterStatus
}

func NewStandaloneServer() *Server {
	server := &Server{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
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
	if config.Properties.AppendOnly {
		// aof
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

}

func (server *Server) Close() {}

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
func (server *Server) GetDBSize(dbIndex int) (int, int) {
	return 0, 0
}
