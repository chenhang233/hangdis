package rdb

import (
	"encoding/json"
	"fmt"
	SortedSet "hangdis/datastruct/sortedset"
	"hangdis/utils/logs"
	"os"
	"strconv"
)

var CRLF = []byte("\r\n")

const (
	typeString = iota
	typeList
	typeSet
	typeHashMap
	typeZSet
	typeExpireTime
)

type TTLOption uint64

type Encoder struct {
	f *os.File
}

func NewEncoder(f *os.File) *Encoder {
	return &Encoder{
		f: f,
	}
}

func (e *Encoder) WriteDBHeader(dbIndex uint, keyCount, ttlCount uint64) error {
	sf := fmt.Sprintf("%s,%s,%s", strconv.FormatUint(uint64(dbIndex), 10), strconv.FormatUint(keyCount, 10), strconv.FormatUint(ttlCount, 10))
	e.Write([]byte(sf))
	return nil
}

func (e *Encoder) WriteCRLF() {
	_, err := e.f.Write(CRLF)
	if err != nil {
		logs.LOG.Debug.Println(err)
	}
}

func (e *Encoder) Write(p []byte) {
	_, err := e.f.Write(p)
	if err != nil {
		logs.LOG.Debug.Println(err)
	}
	e.WriteCRLF()
}

func (e *Encoder) WriteString(s string) {
	_, err := e.f.WriteString(s)
	if err != nil {
		logs.LOG.Debug.Println(err)
	}
	e.WriteCRLF()
}

func (e *Encoder) writeTTL(expiration TTLOption) {
	e.Write([]byte{typeExpireTime})
	e.Write([]byte{' '})
	e.Write([]byte(strconv.FormatUint(uint64(expiration), 10)))
	e.Write([]byte{' '})
}

func (e *Encoder) beforeWriteObject(options ...any) {
	for _, opt := range options {
		switch o := opt.(type) {
		case TTLOption:
			e.writeTTL(o)
		}
	}
}

func (e *Encoder) WriteStringObject(key string, value []byte, options ...any) error {
	e.beforeWriteObject(options)
	e.Write([]byte{typeString})
	e.WriteString(key)
	e.Write(value)
	return nil
}
func (e *Encoder) WriteListObject(key string, values [][]byte, options ...any) error {
	e.beforeWriteObject(options)
	e.Write([]byte{typeList})
	e.WriteString(key)
	for _, v := range values {
		e.Write(v)
	}
	e.WriteCRLF()
	return nil
}
func (e *Encoder) WriteSetObject(key string, values [][]byte, options ...any) error {
	e.beforeWriteObject(options)
	e.Write([]byte{typeSet})
	e.WriteString(key)
	for _, v := range values {
		e.Write(v)
	}
	e.WriteCRLF()
	return nil
}

func (e *Encoder) WriteHashMapObject(key string, hash map[string][]byte, options ...any) error {
	e.beforeWriteObject(options)
	e.Write([]byte{typeHashMap})
	e.WriteString(key)
	for k, v := range hash {
		e.WriteString(k)
		e.Write(v)
	}
	e.WriteCRLF()
	return nil
}

func (e *Encoder) WriteZSetObject(key string, entries []*SortedSet.Element, options ...any) error {
	e.beforeWriteObject(options)
	e.Write([]byte{typeZSet})
	e.WriteString(key)
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	e.Write(data)
	e.WriteCRLF()
	return nil
}
