package sortedset

import (
	"errors"
	"strconv"
)

type Element struct {
	Member string
	Score  float64
}

type ScoreBorder struct {
	Inf     int8
	Value   float64
	Exclude bool
}

func ParseScoreBorder(s string) (*ScoreBorder, error) {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, errors.New("ERR min or max is not a float")
	}
	return &ScoreBorder{
		Inf:     0,
		Value:   value,
		Exclude: false,
	}, nil
}
func (border *ScoreBorder) greater(value float64) bool {
	return border.Value >= value
}
func (border *ScoreBorder) less(value float64) bool {
	return border.Value <= value
}

type SortedSet interface {
	Add(member string, score float64) bool
	Remove(member string) bool
	Get(member string) (element *Element, ok bool)
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
