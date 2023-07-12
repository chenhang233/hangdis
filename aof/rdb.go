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
	time.Sleep(time.Second * 5)
	if err != nil {
		logs.LOG.Debug.Println("tmp file create failed")
		return err
	}

	tmpHandler := p.newRewriteHandler()
	tmpHandler.LoadAof(int(filesize))
	encoder := rdb.NewEncoder(tmpFile)
	for i := 0; i < config.Properties.Databases; i++ {
		var err error
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
