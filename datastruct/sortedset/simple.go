package sortedset

type SimpleSortedSet struct {
	dict map[string]*Element
}

func Make() *SimpleSortedSet {
	return &SimpleSortedSet{}
}

func (s *SimpleSortedSet) Add(key string, score float64) bool {
	_, ok := s.dict[key]
	s.dict[key] = &Element{
		key:   key,
		Score: score,
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
