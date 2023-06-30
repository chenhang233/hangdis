package sortedset

type Element struct {
	key   string
	Score float64
}

type SortedSet interface {
	Add(key string, score float64) bool
	Remove(key string) bool
	Get(key string) (element *Element, ok bool)
	Len() int64
}
