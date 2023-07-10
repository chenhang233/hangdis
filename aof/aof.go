package aof

import (
	"context"
	"hangdis/interface/database"
	"os"
	"sync"
)

const (
	aofQueueSize = 1 << 16
)

type Listener interface {
	Callback([]CmdLine)
}

type CmdLine = [][]byte

type payload struct {
	cmdLine CmdLine
	dbIndex int
	wg      *sync.WaitGroup
}

type PerSister struct {
	ctx         context.Context
	cancel      context.CancelFunc
	db          database.DBEngine
	tmpDBMaker  func() database.DBEngine
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	aofFsync    string
	aofFinished chan struct{}
	pausingAof  sync.Mutex
	currentDB   int
	listeners   map[Listener]struct{}
	buffer      []CmdLine
}

func NewPerSister(db database.DBEngine, filename string, load bool, fsync string, tmpDBMaker func() database.DBEngine) (*PerSister, error) {

}

func (p *PerSister) LoadAof(maxBytes int) {

}
