package database

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

type shard struct {
	m     map[string]interface{}
	mutex sync.RWMutex
}

func TestRWMutex(t *testing.T) {
	s := &shard{
		m:     map[string]interface{}{},
		mutex: sync.RWMutex{},
	}
	fmt.Println(s.mutex, " s.mutex")
	s.mutex.Lock()
}

func TestFormatFloat(t *testing.T) {
	f := 1.21
	f += 0.22
	float := strconv.FormatFloat(f, 'f', -1, 64)
	fmt.Println(float)
}

func TestUnixTime(t *testing.T) {
	num := int64(1687587028)
	unix := time.Unix(num, 0)
	fmt.Println(unix)
}
