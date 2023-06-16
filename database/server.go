package database

import (
	"hangdis/config"
	"hangdis/interface/database"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	dbSet []*atomic.Value // *DB

	//hub *pubsub.Hub
	//persister *aof.Persister

	//role int32
	//slaveStatus  *slaveStatus
	//masterStatus *masterStatus
}

func init() {
	RegisterSystemCommand("PING", Ping)
	RegisterSystemCommand("INFO", Info)
}

func NewStandaloneServer() *Server {
	server := &Server{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}

	if config.Properties.AppendOnly {
		// aof
	}
	server.dbSet = make([]*atomic.Value, config.Properties.Databases)
	for i := range server.dbSet {
		db := makeDB()
		db.index = i
		holder := &atomic.Value{}
		holder.Store(db)
		server.dbSet[i] = holder
	}
	return server
}

func (server *Server) Exec(c redis.Connection, cmdLine [][]byte) (result redis.Reply) {
	cmdName := strings.ToLower(string(cmdLine[0]))
	if sysCmd, ok := systemTable[cmdName]; ok {
		exec := sysCmd.executor
		return exec(c, cmdLine)
	}
	return protocol.MakeEmptyMultiBulkReply()
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
