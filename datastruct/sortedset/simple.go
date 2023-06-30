package sortedset

type SimpleSortedSet struct {
	dict map[string]*Element
}

func SimpleMake() *SimpleSortedSet {
	return &SimpleSortedSet{
		dict: make(map[string]*Element),
	}
}

func (s *SimpleSortedSet) Add(key string, score float64) bool {
	_, ok := s.dict[key]
	s.dict[key] = &Element{
		Member: key,
		Score:  score,
	}
	if ok {
		return false
	} else {
		return true
	}
}

func (s *SimpleSortedSet) Len() int64 {
	return int64(len(s.dict))
}

func (s *SimpleSortedSet) Get(key string) (element *Element, ok bool) {
	v, ok := s.dict[key]
	if !ok {
		return nil, false
	}
	return v, true
}

func (s *SimpleSortedSet) Remove(key string) bool {
	_, ok := s.dict[key]
	if !ok {
		return false
	}
	delete(s.dict, key)
	return true
}

func (s *SimpleSortedSet) GetRank(member string, desc bool) (rank int64) {
	return 0
}
func (s *SimpleSortedSet) ForEach(start int64, stop int64, desc bool, consumer func(element *Element) bool) {
	return
}
func (s *SimpleSortedSet) Range(start int64, stop int64, desc bool) []*Element {
	return nil
}
func (s *SimpleSortedSet) Count(min *ScoreBorder, max *ScoreBorder) int64 {
	return 0
}
func (s *SimpleSortedSet) RangeByScore(min *ScoreBorder, max *ScoreBorder, offset int64, limit int64, desc bool) []*Element {
	return nil
}
func (s *SimpleSortedSet) RemoveByScore(min *ScoreBorder, max *ScoreBorder) int64 {
	return 0
}
func (s *SimpleSortedSet) PopMin(count int) []*Element {
	return nil
}
func (s *SimpleSortedSet) RemoveByRank(start int64, stop int64) int64 {
	return 0
}
