package database

import (
	"hangdis/config"
	"hangdis/interface/database"
	"hangdis/interface/redis"
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

func NewStandaloneServer() *Server {
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}

	if config.Properties.AppendOnly {
		// aof
	}

}

func (server *Server) Exec(c redis.Connection, cmdLine [][]byte) (result redis.Reply) {}

func (server *Server) AfterClientClose(c redis.Connection) {

}

func (server *Server) Close() {}

func (server *Server) ExecWithLock(conn redis.Connection, cmdLine [][]byte) redis.Reply {

}
func (server *Server) ExecMulti(conn redis.Connection, watching map[string]uint32, cmdLines []database.CmdLine) redis.Reply {

}
func (server *Server) GetUndoLogs(dbIndex int, cmdLine [][]byte) []database.CmdLine {

}
func (server *Server) ForEach(dbIndex int, cb func(key string, data *database.DataEntity, expiration *time.Time) bool) {

}
func (server *Server) RWLocks(dbIndex int, writeKeys []string, readKeys []string) {

}
func (server *Server) RWUnLocks(dbIndex int, writeKeys []string, readKeys []string) {

}
func (server *Server) GetDBSize(dbIndex int) (int, int) {

}
