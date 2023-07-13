package aof

import (
	"hangdis/config"
	"hangdis/utils/logs"
	"os"
)

func (p *PerSister) newRewriteHandler() *PerSister {
	h := &PerSister{}
	h.aofFilename = p.aofFilename
	h.db = p.tmpDBMaker()
	return h
}

type RewriteCtx struct {
	tmpFile  *os.File
	fileSize int64
	dbIdx    int
}

func (p *PerSister) Rewrite() error {
	ctx, err := p.StartRewrite()
	if err != nil {
		return err
	}
	err = p.DoRewrite(ctx)
	if err != nil {
		return err
	}
	p.FinishRewrite(ctx)
	return nil
}

func (p *PerSister) DoRewrite(ctx *RewriteCtx) (err error) {
	if !config.Properties.AofUseRdbPreamble {
		logs.LOG.Info.Println("generateAof")
		err = p.generateAof(ctx)
	} else {
		logs.LOG.Info.Println("generateRDB")
		err = p.generateRDB(ctx)
	}
	return err
}

func (p *PerSister) StartRewrite() (*RewriteCtx, error) {
	p.pausingAof.Lock()
	defer p.pausingAof.Unlock()
	err := p.aofFile.Sync()
	if err != nil {
		logs.LOG.Error.Println(err)
		return nil, err
	}
	fileInfo, _ := os.Stat(p.aofFilename)
	filesize := fileInfo.Size()
	tmpFile, err := os.CreateTemp(config.GetTmpDir(), "*.aof")
	if err != nil {
		logs.LOG.Debug.Println("tmp file create failed")
		return nil, err
	}
	return &RewriteCtx{
		tmpFile:  tmpFile,
		fileSize: filesize,
		dbIdx:    p.currentDB,
	}, nil
}

func (p *PerSister) FinishRewrite(ctx *RewriteCtx) {}
