package sortedset

type Element struct {
	key   string
	Score float64
}

type ScoreBorder struct {
	Inf     int8
	Value   float64
	Exclude bool
}

type SortedSet interface {
	Add(key string, score float64) bool
	Remove(key string) bool
	Get(key string) (element *Element, ok bool)
	Len() int64
	GetRank(member string, desc bool) (rank int64)
	ForEach(start int64, stop int64, desc bool, consumer func(element *Element) bool)
	Range(start int64, stop int64, desc bool) []*Element
	Count(min *ScoreBorder, max *ScoreBorder) int64
	RangeByScore(min *ScoreBorder, max *ScoreBorder, offset int64, limit int64, desc bool) []*Element
	RemoveByScore(min *ScoreBorder, max *ScoreBorder) int64
	PopMin(count int) []*Element
	RemoveByRank(start int64, stop int64) int64
}
