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

func (s *SimpleSortedSet) QuickSort(elements *[]*Element, l int, r int) {
	if l < r {
		i, j, x := l, r, (*elements)[l]
		for i < j {
			for i < j && (*elements)[j].Score >= x.Score {
				j--
			}
			if i < j {
				(*elements)[i] = (*elements)[j]
				i++
			}
			for i < j && (*elements)[i].Score < x.Score {
				i++
			}
			if i < j {
				(*elements)[j] = (*elements)[i]
				j--
			}
		}
		(*elements)[i] = x
		s.QuickSort(elements, l, i-1)
		s.QuickSort(elements, i+1, r)
	}
}

func (s *SimpleSortedSet) ToSlice() []*Element {
	arr := make([]*Element, s.Len())
	i := 0
	for _, element := range s.dict {
		arr[i] = element
		i++
	}
	return arr
}

func (s *SimpleSortedSet) findElement(element *Element, arr []*Element) int {
	for i, e := range arr {
		if e.Member == element.Member {
			return i
		}
	}
	return -1
}

func (s *SimpleSortedSet) GetRank(member string, desc bool) (rank int64) {
	element, ok := s.dict[member]
	if !ok {
		return -1
	}
	slice := s.ToSlice()
	s.QuickSort(&slice, 0, int(s.Len())-1)
	fe := s.findElement(element, slice)
	if desc {
		return s.Len() - 1 - int64(fe)
	}
	return int64(fe)
}
func (s *SimpleSortedSet) ForEach(start int64, stop int64, desc bool, consumer func(element *Element) bool) {
	n := s.Len()
	if start >= n || start < 0 || stop >= n || stop < start {
		panic("illegal  start or stop")
	}
	slice := s.ToSlice()
	s.QuickSort(&slice, 0, int(s.Len())-1)
	if desc {
		for i := stop - 1; i >= start; i-- {
			if !consumer(slice[i]) {
				break
			}
		}
	} else {
		for i := start; i < stop; i++ {
			if !consumer(slice[i]) {
				break
			}
		}
	}
}
func (s *SimpleSortedSet) Range(start int64, stop int64, desc bool) []*Element {
	sliceSize := int(stop - start)
	slice := make([]*Element, sliceSize)
	i := 0
	s.ForEach(start, stop, desc, func(element *Element) bool {
		slice[i] = element
		i++
		return true
	})
	return slice
}
func (s *SimpleSortedSet) Count(min *ScoreBorder, max *ScoreBorder) int64 {
	var i int64 = 0
	s.ForEach(0, s.Len(), false, func(element *Element) bool {
		gtMin := min.less(element.Score)
		if !gtMin {
			return true
		}
		ltMax := max.greater(element.Score)
		if !ltMax {
			return false
		}
		i++
		return true
	})
	return i
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
