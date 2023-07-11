package aof

import (
	"hangdis/config"
	"hangdis/interface/database"
	"hangdis/utils/logs"
	"os"
	"time"
)

func (p *PerSister) newRewriteHandler() *PerSister {
	h := &PerSister{}
	h.aofFilename = p.aofFilename
	h.db = p.tmpDBMaker()
	return h
}

func (p *PerSister) GenerateRDB(rdbFilename string) error {
	p.pausingAof.Lock()
	defer p.pausingAof.Unlock()
	err := p.aofFile.Sync()
	if err != nil {
		logs.LOG.Error.Println(err)
		return err
	}
	fileInfo, _ := os.Stat(p.aofFilename)
	filesize := fileInfo.Size()
	tmpFile, err := os.CreateTemp(config.GetTmpDir(), "*.aof")
	if err != nil {
		logs.LOG.Debug.Println("tmp file create failed")
		return err
	}

	tmpHandler := p.newRewriteHandler()
	tmpHandler.LoadAof(int(filesize))
	for i := 0; i < config.Properties.Databases; i++ {
		tmpHandler.db.ForEach(i, func(key string, data *database.DataEntity, expiration *time.Time) bool {

		})
	}
	tmpFile.Write()
	err = tmpFile.Close()
	if err != nil {
		logs.LOG.Error.Println(err)
		return err
	}
	err = os.Rename(tmpFile.Name(), rdbFilename)
	if err != nil {
		return err
	}
	return nil
}
