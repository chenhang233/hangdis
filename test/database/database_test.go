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

func TestStrToInt(t *testing.T) {
	a := "5"
	delta, err := strconv.ParseInt(a, 10, 64)
	if err != nil {
		panic(err)
	}
	fmt.Println(delta, " d")
}

func TestRange(t *testing.T) {
	arr := []int{1, 2, 3}
	for _, v := range arr {
		if v == 2 {
			continue
		}
		fmt.Println(v, "v")
	}
}
