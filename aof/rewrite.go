package aof

import (
	"hangdis/config"
	"hangdis/utils/logs"
	"io"
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
	if config.Properties.AofUseRdbPreamble {
		logs.LOG.Info.Println("generateRDB")
		err = p.generateRDB(ctx)
	} else {
		logs.LOG.Info.Println("generateAof")
		err = p.generateAof(ctx)
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

func (p *PerSister) FinishRewrite(ctx *RewriteCtx) {
	p.pausingAof.Lock()
	defer p.pausingAof.Unlock()
	file := ctx.tmpFile
	src, err := os.Open(p.aofFilename)
	if err != nil {
		logs.LOG.Error.Println(err)
		return
	}
	defer func() {
		_ = src.Close()
		_ = file.Close()
	}()
	_, err = src.Seek(ctx.fileSize, 0)
	if err != nil {
		logs.LOG.Error.Println(err)
		return
	}
	//data := protocol.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(ctx.dbIdx))).ToBytes()
	//_, err = file.Write(data)
	//if err != nil {
	//	logs.LOG.Error.Println(err)
	//	return
	//}
	_, err = io.Copy(file, src)
	if err != nil {
		logs.LOG.Error.Println(err)
		return
	}
	_ = p.aofFile.Close()
	err = os.Rename(file.Name(), p.aofFilename)
	if err != nil {
		logs.LOG.Error.Println(err)
	}
	aofFile, err := os.OpenFile(p.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		panic(err)
	}
	p.aofFile = aofFile
	//data = protocol.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.currentDB))).ToBytes()
	//_, err = p.aofFile.Write(data)
	//if err != nil {
	//	panic(err)
	//}
}
