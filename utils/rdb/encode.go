package rdb

import (
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

func (e *Encoder) Write(p []byte) {
	_, err := e.f.Write(p)
	if err != nil {
		logs.LOG.Debug.Println(err)
	}
}

func (e *Encoder) WriteString(s string) {
	_, err := e.f.WriteString(s)
	if err != nil {
		logs.LOG.Debug.Println(err)
	}
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
	e.Write(CRLF)
}
func (e *Encoder) WriteListObject(key string, values [][]byte, options ...any) error {

}
func (e *Encoder) WriteSetObject(key string, value [][]byte, options ...any) error {

}

func (e *Encoder) WriteHashMapObject(key string, hash map[string][]byte, options ...any) error {

}

func (e *Encoder) WriteZSetObject(key string, entries []*model.ZSetEntry, options ...any) error {

}
