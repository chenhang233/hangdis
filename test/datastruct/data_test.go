package datastruct

import (
	"fmt"
	"testing"
)

const prime32 = uint32(16777619)

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func TestDictSpread(t *testing.T) {
	size := 50
	u := fnv32("runoobkey")
	fmt.Println("code: ", u)
	test := uint32(size-1) & u
	fmt.Println("index: ", test)
}

func TestBit(t *testing.T) {
	b := byte(2)
	fmt.Println(b)
	c := byte(1 << 2)
	fmt.Println(c, "c")
	b &^= c
	fmt.Println(b, 1<<3)
}
