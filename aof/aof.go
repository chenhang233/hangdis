package aof

import (
	"context"
	"fmt"
	"hangdis/interface/database"
	"hangdis/utils"
	"hangdis/utils/logs"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	aofQueueSize = 1 << 16
)

const (
	FsyncEverySec = "everysec"
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
	p := &PerSister{}
	p.aofFilename = filename
	p.aofFsync = strings.ToLower(fsync)
	p.db = db
	p.tmpDBMaker = tmpDBMaker
	p.currentDB = 0
	if load {
		p.LoadAof(0)
	}
	aofFile, err := os.OpenFile(p.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	p.aofFile = aofFile
	p.aofChan = make(chan *payload, aofQueueSize)
	p.aofFinished = make(chan struct{})
	p.listeners = make(map[Listener]struct{})
	go func() {
		p.listenCmd()
	}()
	ctx, cancel := context.WithCancel(context.Background())
	p.ctx = ctx
	p.cancel = cancel
	if p.aofFsync == FsyncEverySec {
		p.fsyncEverySecond()
	}
	return p, nil
}

func (p *PerSister) LoadAof(maxBytes int) {

}

func (p *PerSister) listenCmd() {
	for pc := range p.aofChan {
		p.writeAof(pc)
	}
	p.aofFinished <- struct{}{}
}

func (p *PerSister) writeAof(payload *payload) {

}

func (p *PerSister) Fsync() {
	p.pausingAof.Lock()
	if err := p.aofFile.Sync(); err != nil {
		logs.LOG.Error.Println(utils.Red(fmt.Sprintf("fsync failed: %v", err)))
	}
	p.pausingAof.Unlock()
}

func (p *PerSister) fsyncEverySecond() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				p.Fsync()
			case <-p.ctx.Done():
				return
			}
		}
	}()
}
