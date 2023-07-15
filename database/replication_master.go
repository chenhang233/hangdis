package database

import (
	"hangdis/interface/redis"
	"hangdis/utils"
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

func (listener *replAofListener) Callback(cmdLines []CmdLine) {}

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
