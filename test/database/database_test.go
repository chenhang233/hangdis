package database

import (
	"fmt"
	"sync"
	"testing"
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
