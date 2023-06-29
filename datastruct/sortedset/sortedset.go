package sortedset

type Element struct {
	key   string
	Score float64
}

type SortedSet struct {
	dict map[string]*Element
}

func Make() *SortedSet {
	return &SortedSet{}
}

func (sortedSet *SortedSet) Add(key string, score float64) bool {
	return true
}

func (sortedSet *SortedSet) Len() int64 {
	return 0
}

func (sortedSet *SortedSet) Get(key string) (element *Element, ok bool) {
	return nil, false
}

func (sortedSet *SortedSet) Remove(key string) bool {
	return true
}
