package database

import (
	"fmt"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/logs"
	"os"
	"sync"
	"time"
)

const (
	masterRole = iota
	slaveRole
)

const (
	bgSaveIdle = uint8(iota)
	bgSaveRunning
	bgSaveFinish
)

type slaveClient struct {
	conn         redis.Connection
	state        uint8
	offset       int64
	lastAckTime  time.Time
	announceIp   string
	announcePort int
	capacity     uint8
}
type replBacklog struct {
	buf           []byte
	beginOffset   int64
	currentOffset int64
}

func (backlog *replBacklog) appendBytes(bin []byte) {
	backlog.buf = append(backlog.buf, bin...)
	backlog.currentOffset += int64(len(bin))
}
func (backlog *replBacklog) getSnapshot() ([]byte, int64) {
	return backlog.buf[:], backlog.currentOffset
}

func (backlog *replBacklog) getSnapshotAfter(beginOffset int64) ([]byte, int64) {
	beg := beginOffset - backlog.beginOffset
	return backlog.buf[beg:], backlog.currentOffset
}

func (backlog *replBacklog) isValidOffset(offset int64) bool {
	return offset >= backlog.beginOffset && offset < backlog.currentOffset
}

type replAofListener struct {
	mdb         *Server
	backlog     *replBacklog
	readyToSend bool
}

func (listener *replAofListener) Callback(cmdLines []CmdLine) {
	listener.mdb.masterStatus.mu.Lock()
	for _, cmdLine := range cmdLines {
		reply := protocol.MakeMultiBulkReply(cmdLine)
		listener.backlog.appendBytes(reply.ToBytes())
	}
	listener.mdb.masterStatus.mu.Unlock()
	if listener.readyToSend {
		if err := listener.mdb.masterSendUpdatesToSlave(); err != nil {
			logs.LOG.Error.Println(err)
		}
	}
}

type masterStatus struct {
	mu           sync.RWMutex
	replId       string
	slaveMap     map[redis.Connection]*slaveClient
	waitSlaves   map[*slaveClient]struct{}
	onlineSlaves map[*slaveClient]struct{}
	bgSaveState  uint8
	rdbFilename  string
	aofListener  *replAofListener
	backlog      *replBacklog
	rewriting    bool
}

func (server *Server) masterSendUpdatesToSlave() error {
	onlineSlaves := make(map[*slaveClient]struct{})
	server.masterStatus.mu.RLock()
	beginOffset := server.masterStatus.backlog.beginOffset
	backlog, currentOffset := server.masterStatus.backlog.getSnapshot()
	for slave := range server.masterStatus.onlineSlaves {
		onlineSlaves[slave] = struct{}{}
	}
	server.masterStatus.mu.RUnlock()
	for slave := range onlineSlaves {
		slaveBeginOffset := slave.offset - beginOffset
		_, err := slave.conn.Write(backlog[slaveBeginOffset:])
		if err != nil {
			logs.LOG.Error.Println(fmt.Sprintf("send updates backlog to slave failed: %v", err))
			server.removeSlave(slave)
			continue
		}
		slave.offset = currentOffset
	}
	return nil
}

func (server *Server) removeSlave(slave *slaveClient) {
	server.masterStatus.mu.Lock()
	defer server.masterStatus.mu.Unlock()
	_ = slave.conn.Close()
	delete(server.masterStatus.slaveMap, slave.conn)
	delete(server.masterStatus.waitSlaves, slave)
	delete(server.masterStatus.onlineSlaves, slave)
	logs.LOG.Info.Println(fmt.Sprintf("disconnect with slave  %s", slave.conn.Name()))
}

func (server *Server) initMasterStatus() {
	server.masterStatus = &masterStatus{
		mu:           sync.RWMutex{},
		replId:       utils.RandomUUID(),
		backlog:      &replBacklog{},
		slaveMap:     make(map[redis.Connection]*slaveClient),
		waitSlaves:   make(map[*slaveClient]struct{}),
		onlineSlaves: make(map[*slaveClient]struct{}),
		bgSaveState:  bgSaveIdle,
		rdbFilename:  "",
	}
}

func (server *Server) stopMasterStatus() {
	server.masterStatus.mu.Lock()
	defer server.masterStatus.mu.Unlock()
	for _, slave := range server.masterStatus.slaveMap {
		_ = slave.conn.Close()
		delete(server.masterStatus.slaveMap, slave.conn)
		delete(server.masterStatus.waitSlaves, slave)
		delete(server.masterStatus.onlineSlaves, slave)
	}
	if server.perSister != nil {
		server.perSister.RemoveListener(server.masterStatus.aofListener)
	}
	_ = os.Remove(server.masterStatus.rdbFilename)
	server.masterStatus.rdbFilename = ""
	server.masterStatus.replId = ""
	server.masterStatus.backlog = &replBacklog{}
	server.masterStatus.slaveMap = make(map[redis.Connection]*slaveClient)
	server.masterStatus.waitSlaves = make(map[*slaveClient]struct{})
	server.masterStatus.onlineSlaves = make(map[*slaveClient]struct{})
	server.masterStatus.bgSaveState = bgSaveIdle
}
