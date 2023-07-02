package sortedset

import "fmt"

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
		for i := stop; i >= start; i-- {
			if !consumer(slice[i]) {
				break
			}
		}
	} else {
		for i := start; i <= stop; i++ {
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
	s.ForEach(start, stop-1, desc, func(element *Element) bool {
		slice[i] = element
		i++
		return true
	})
	return slice
}

func (s *SimpleSortedSet) Count(min *ScoreBorder, max *ScoreBorder) int64 {
	var i int64 = 0
	s.ForEach(0, s.Len()-1, false, func(element *Element) bool {
		fmt.Println(element.Score, min.Value, max.Value, "------144")
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
	if offset < 0 || limit == 0 {
		return make([]*Element, 0)
	}
	var slice []*Element
	s.ForEach(0, s.Len()-1, desc, func(element *Element) bool {
		if element.Score > max.Value {
			return false
		}
		if element.Score >= min.Value {
			slice = append(slice, element)
		}
		return true
	})
	if limit > 0 {
		return slice[offset:limit]
	}
	return slice[offset:]
}

func (s *SimpleSortedSet) RemoveByScore(min *ScoreBorder, max *ScoreBorder) int64 {
	slice := s.RangeByScore(min, max, 0, 0, false)
	res := 0
	for _, element := range slice {
		if s.Remove(element.Member) {
			res++
		}
	}
	return int64(res)
}

func (s *SimpleSortedSet) PopMin(count int) []*Element {
	var removed []*Element
	s.ForEach(0, s.Len()-1, false, func(element *Element) bool {
		if count == 0 {
			return false
		}
		count--
		removed = append(removed, element)
		return true
	})
	for _, member := range removed {
		s.Remove(member.Member)
	}
	return removed
}

func (s *SimpleSortedSet) RemoveByRank(start int64, stop int64) int64 {
	elements := s.Range(start, stop, true)
	i := 0
	for _, element := range elements {
		if s.Remove(element.Member) {
			i++
		}
	}
	return int64(i)
}
