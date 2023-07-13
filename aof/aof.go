package aof

import (
	"context"
	"fmt"
	"hangdis/config"
	"hangdis/interface/database"
	"hangdis/redis/connection"
	"hangdis/redis/parser"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/logs"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	aofQueueSize = 1 << 16
)

const (
	FsyncAlways   = "always"
	FsyncNo       = "no"
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
	//listen buffer
	listeners map[Listener]struct{}
	buffer    []CmdLine
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
	aofChan := p.aofChan
	p.aofChan = nil
	defer func(aofChan chan *payload) {
		p.aofChan = aofChan
	}(aofChan)
	file, err := os.Open(p.aofFilename)
	if err != nil {
		logs.LOG.Error.Println(err)
		return
	}
	defer file.Close()
	stream := parser.ParseStream(file)
	fakeConn := connection.NewFakeConn()
	for sp := range stream {
		if sp.Err != nil {
			if sp.Err == io.EOF {
				break
			}
			logs.LOG.Warn.Println("parse error: " + sp.Err.Error())
			continue
		}
		if sp.Data == nil {
			logs.LOG.Warn.Println("empty payload")
			continue
		}
		r, ok := sp.Data.(*protocol.MultiBulkReply)
		if !ok {
			logs.LOG.Warn.Println("require multi bulk protocol")
			continue
		}
		res := p.db.Exec(fakeConn, r.Args)
		if protocol.IsErrorReply(res) {
			logs.LOG.Error.Println("exec err", string(res.ToBytes()))
			continue
		}
		if strings.ToLower(string(r.Args[0])) == "select" {
			dbIndex, err := strconv.Atoi(string(r.Args[1]))
			if err != nil {
				logs.LOG.Debug.Println(err, " db index :", dbIndex)
			}
			p.currentDB = dbIndex
		}
	}
}

func (p *PerSister) listenCmd() {
	for pc := range p.aofChan {
		p.writeAof(pc)
	}
	p.aofFinished <- struct{}{}
}

func (p *PerSister) writeAof(pd *payload) {
	p.buffer = p.buffer[0:]
	p.pausingAof.Lock()
	defer p.pausingAof.Unlock()
	if p.currentDB != pd.dbIndex {
		selectCmd := utils.ToCmdLine("SELECT", strconv.Itoa(pd.dbIndex))
		p.buffer = append(p.buffer, selectCmd)
		data := protocol.MakeMultiBulkReply(selectCmd).ToBytes()
		_, err := p.aofFile.Write(data)
		if err != nil {
			logs.LOG.Warn.Println(err)
			return
		}
		p.currentDB = pd.dbIndex
	}
	data := protocol.MakeMultiBulkReply(pd.cmdLine).ToBytes()
	p.buffer = append(p.buffer, pd.cmdLine)
	_, err := p.aofFile.Write(data)
	if err != nil {
		logs.LOG.Warn.Println(err)
		return
	}
	for listener := range p.listeners {
		listener.Callback(p.buffer)
	}
	if p.aofFsync == FsyncAlways {
		_ = p.aofFile.Sync()
	}
}

func (p *PerSister) Fsync() {
	p.pausingAof.Lock()
	if err := p.aofFile.Sync(); err != nil {
		logs.LOG.Error.Println(utils.Red(fmt.Sprintf("fsync failed: %v", err)))
	}
	p.pausingAof.Unlock()
}

func (p *PerSister) SaveCmdLine(dbIndex int, cmdLine CmdLine) {
	if p.aofChan == nil {
		logs.LOG.Warn.Println(utils.Red("aofChan not found"))
		return
	}
	pd := &payload{
		dbIndex: dbIndex,
		cmdLine: cmdLine,
	}
	if p.aofFsync == FsyncAlways {
		p.writeAof(pd)
		return
	}
	p.aofChan <- pd
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

func (p *PerSister) generateAof(ctx *RewriteCtx) error {
	tempFile := ctx.tmpFile
	tmpAof := p.newRewriteHandler()
	tmpAof.LoadAof(int(ctx.fileSize))
	for i := 0; i < config.Properties.Databases; i++ {
		data := protocol.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(i))).ToBytes()
		_, err := tempFile.Write(data)
		if err != nil {
			return err
		}
		tmpAof.db.ForEach(i, func(key string, data *database.DataEntity, expiration *time.Time) bool {
			cmd := EntityToCmd(key, data)
			if cmd != nil {
				_, _ = tempFile.Write(cmd.ToBytes())
			}
			if expiration != nil {
				cmd := MakeExpireCmd(key, *expiration)
				if cmd != nil {
					_, _ = tempFile.Write(cmd.ToBytes())
				}
			}
			return true
		})
	}
	return nil
}
