package aof

import (
	"hangdis/config"
	"hangdis/datastruct/dict"
	List "hangdis/datastruct/list"
	"hangdis/datastruct/set"
	SortedSet "hangdis/datastruct/sortedset"
	"hangdis/interface/database"
	"hangdis/utils/logs"
	"hangdis/utils/rdb"
	"os"
	"time"
)

func (p *PerSister) startGenerateRDB(newListener Listener, hook func()) (*RewriteCtx, error) {
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
	if newListener != nil {
		p.listeners[newListener] = struct{}{}
	}
	if hook != nil {
		hook()
	}
	return &RewriteCtx{
		tmpFile:  tmpFile,
		fileSize: filesize,
		dbIdx:    p.currentDB,
	}, nil
}

func (p *PerSister) GenerateRDB(rdbFilename string) error {
	ctx, err := p.startGenerateRDB(nil, nil)
	if err != nil {
		return err
	}
	err = p.generateRDB(ctx)
	if err != nil {
		return err
	}
	err = ctx.tmpFile.Close()
	if err != nil {
		logs.LOG.Error.Println(err)
		return err
	}
	err = os.Rename(ctx.tmpFile.Name(), rdbFilename)
	if err != nil {
		return err
	}
	return nil
}

func (p *PerSister) generateRDB(ctx *RewriteCtx) error {
	tmpHandler := p.newRewriteHandler()
	tmpHandler.LoadAof(int(ctx.fileSize))
	encoder := rdb.NewEncoder(ctx.tmpFile)
	for i := 0; i < config.Properties.Databases; i++ {
		keyCount, ttlCount := tmpHandler.db.GetDBSize(i)
		if keyCount == 0 {
			continue
		}
		var err error
		err = encoder.WriteDBHeader(uint(i), uint64(keyCount), uint64(ttlCount))
		if err != nil {
			return err
		}
		tmpHandler.db.ForEach(i, func(key string, entity *database.DataEntity, expiration *time.Time) bool {
			var opts []interface{}
			if expiration != nil {
				opts = append(opts, uint64(expiration.Unix()))
			}
			switch obj := entity.Data.(type) {
			case []byte:
				err = encoder.WriteStringObject(key, obj, opts...)
			case List.List:
				vals := make([][]byte, 0, obj.Len())
				obj.ForEach(func(i int, v interface{}) bool {
					bytes, _ := v.([]byte)
					vals = append(vals, bytes)
					return true
				})
				err = encoder.WriteListObject(key, vals, opts...)
			case set.Set:
				vals := make([][]byte, 0, obj.Len())
				obj.ForEach(func(m string) bool {
					vals = append(vals, []byte(m))
					return true
				})
				err = encoder.WriteSetObject(key, vals, opts...)
			case dict.Dict:
				hash := make(map[string][]byte)
				obj.ForEach(func(key string, val interface{}) bool {
					bytes, _ := val.([]byte)
					hash[key] = bytes
					return true
				})
				err = encoder.WriteHashMapObject(key, hash, opts...)
			case SortedSet.SortedSet:
				var entries []*SortedSet.Element
				obj.ForEach(int64(0), obj.Len(), false, func(element *SortedSet.Element) bool {
					entries = append(entries, &SortedSet.Element{
						Member: element.Member,
						Score:  element.Score,
					})
					return true
				})
				err = encoder.WriteZSetObject(key, entries, opts...)
			}
			if err != nil {
				return false
			}
			return true
		})
	}
	return nil
}
