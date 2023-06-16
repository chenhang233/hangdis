package dict

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

const prime32 = uint32(16777619)

type ConcurrentDict struct {
	table      []*shard
	count      int32 // key count
	shardCount int   // table count
}

type shard struct {
	m     map[string]interface{}
	mutex sync.RWMutex
}

func computeCapacity(param int) (size int) {
	if param <= 16 {
		return 16
	}
	n := param - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if n < 0 {
		return math.MaxInt32
	}
	return n + 1
}

// GetHashCode32 return hashCode
func GetHashCode32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

// return index
func (dict *ConcurrentDict) spread(hashCode uint32) uint32 {
	if dict == nil {
		panic("dict is nil")
	}
	tableSize := uint32(len(dict.table))
	return (tableSize - 1) & hashCode
}

// return *shard table
func (dict *ConcurrentDict) getShard(index uint32) *shard {
	if dict == nil {
		panic("dict is nil")
	}
	return dict.table[index]
}

func MakeConcurrent(shardCount int) *ConcurrentDict {
	shardCount = computeCapacity(shardCount)
	tables := make([]*shard, shardCount)
	for i := 0; i < shardCount; i++ {
		tables[i] = &shard{
			m: make(map[string]interface{}),
		}
	}
	return &ConcurrentDict{
		shardCount: shardCount,
		count:      0,
		table:      tables,
	}
}

func (dict *ConcurrentDict) Get(key string) (val interface{}, exists bool) {
	if dict == nil {
		panic("dict is nil")
	}
	hash := GetHashCode32(key)
	index := dict.spread(hash)
	table := dict.getShard(index)
	table.mutex.RLock()
	defer table.mutex.RUnlock()
	val, exists = table.m[key]
	return
}
func (dict *ConcurrentDict) Len() int {
	if dict == nil {
		panic("dict is nil")
	}
	return int(dict.count)
}
func (dict *ConcurrentDict) Put(key string, val interface{}) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hash := GetHashCode32(key)
	index := dict.spread(hash)
	table := dict.getShard(index)
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if _, ok := table.m[key]; ok {
		table.m[key] = val
		return 0
	}
	dict.count++
	table.m[key] = val
	return 1
}
func (dict *ConcurrentDict) PutIfAbsent(key string, val interface{}) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hash := GetHashCode32(key)
	index := dict.spread(hash)
	table := dict.getShard(index)
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if _, ok := table.m[key]; ok {
		return 0
	}
	dict.count++
	table.m[key] = val
	return 1
}
func (dict *ConcurrentDict) PutIfExists(key string, val interface{}) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hash := GetHashCode32(key)
	index := dict.spread(hash)
	table := dict.getShard(index)
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if _, ok := table.m[key]; ok {
		table.m[key] = val
		return 1
	}
	return 0
}
func (dict *ConcurrentDict) Remove(key string) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hash := GetHashCode32(key)
	index := dict.spread(hash)
	table := dict.getShard(index)
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if _, ok := table.m[key]; ok {
		dict.count--
		delete(table.m, key)
		return 1
	}
	return 0
}
func (dict *ConcurrentDict) ForEach(consumer Consumer) {
	if dict == nil {
		panic("dict is nil")
	}
	for _, v := range dict.table {
		v.mutex.RLock()
		f := func() bool {
			defer v.mutex.RUnlock()
			for k, v := range v.m {
				if !consumer(k, v) {
					return false
				}
			}
			return true
		}
		if !f() {
			continue
		}
	}
}
func (dict *ConcurrentDict) Keys() []string {
	if dict == nil {
		panic("dict is nil")
	}
	keys := make([]string, dict.Len())
	i := 0
	dict.ForEach(func(key string, val interface{}) bool {
		if i < len(keys) {
			keys[i] = key
			i++
		} else {
			keys = append(keys, key)
		}
		return true
	})
	return keys
}

func (s *shard) RandomKey() string {
	if s == nil {
		panic("shard is nil")
	}
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for k, _ := range s.m {
		return k
	}
	return ""
}

func (dict *ConcurrentDict) RandomKeys(limit int) []string {
	if dict == nil {
		panic("dict is nil")
	}
	if limit >= dict.Len() {
		return dict.Keys()
	}
	shardCount := len(dict.table)
	keys := make([]string, limit)
	nR := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < limit; {
		table := dict.getShard(uint32(nR.Intn(shardCount)))
		if table == nil {
			continue
		}
		key := table.RandomKey()
		if key != "" {
			keys = append(keys, key)
			i++
		}
	}
	return keys
}
func (dict *ConcurrentDict) RandomDistinctKeys(limit int) []string {
	if dict == nil {
		panic("dict is nil")
	}
	if limit >= dict.Len() {
		return dict.Keys()
	}
	m := make(map[string]bool)
	shardCount := len(dict.table)
	nR := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < limit; {
		table := dict.getShard(uint32(nR.Intn(shardCount)))
		if table == nil {
			continue
		}
		key := table.RandomKey()
		if key != "" {
			if _, ok := m[key]; !ok {
				m[key] = true
				i++
			}
		}
	}
	keys := make([]string, limit)
	i := 0
	for k := range m {
		keys[i] = k
	}
	return keys
}
func (dict *ConcurrentDict) Clear() {
	*dict = *MakeConcurrent(dict.shardCount)
}
