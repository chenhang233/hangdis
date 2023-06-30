package datastruct

import (
	"fmt"
	SortedSet "hangdis/datastruct/sortedset"
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

func TestQuickSort(t *testing.T) {
	simpleMake := SortedSet.SimpleMake()
	simpleMake.Add("A", 1)
	simpleMake.Add("B", 2)
	simpleMake.Add("C", 3)
	simpleMake.Add("D", 4)
	slice := simpleMake.ToSlice()
	fmt.Println(slice, "slice before")
	simpleMake.QuickSort(&slice, 0, int(simpleMake.Len()-1))
	fmt.Println(slice[0].Score, slice[1].Score, "slice after")
}

func TestGetRnak(t *testing.T) {
	simpleMake := SortedSet.SimpleMake()
	simpleMake.Add("A", 1)
	simpleMake.Add("B", 2)
	simpleMake.Add("C", 3)
	simpleMake.Add("D", 4)
	rank := simpleMake.GetRank("B", false)
	fmt.Println(rank, "rank")
	rank = simpleMake.GetRank("B", true)
	fmt.Println(rank, "rank")
}
