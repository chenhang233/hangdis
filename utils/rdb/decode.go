package rdb

import (
	"os"
	"time"
)

type Decoder struct {
	f *os.File
}

func NewDecoder(f *os.File) *Decoder {
	return &Decoder{
		f: f,
	}
}

func (d *Decoder) Read() {
	//reader := bufio.NewReader(d.f)
	//reader
}

func (d *Decoder) Parse(fn func(RedisObject) bool) error {
	return nil
}

type RedisObject struct {
	V any
}

func (o *RedisObject) GetKey() string {
	return ""
}
func (o *RedisObject) GetExpiration() *time.Time {
	return &time.Time{}
}

func (o *RedisObject) GetDBIndex() int {
	return 0
}

func (o *RedisObject) GetType() int {
	return TypeString
}
