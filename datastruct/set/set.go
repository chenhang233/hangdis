package set

import (
	"hangdis/datastruct/dict"
)

type InstanceSet struct {
	dict dict.Dict
}

func Make(members ...string) *InstanceSet {
	set := &InstanceSet{
		dict: dict.MakeInstanceDict(),
	}
	for _, member := range members {
		set.Add(member)
	}
	return set
}

func (set *InstanceSet) Add(val string) int {
	return set.dict.Put(val, nil)
}

func (set *InstanceSet) Remove(val string) int {
	return set.dict.Remove(val)
}

func (set *InstanceSet) Has(val string) bool {
	if set == nil || set.dict == nil {
		return false
	}
	_, exists := set.dict.Get(val)
	return exists
}
func (set *InstanceSet) Len() int {
	if set == nil || set.dict == nil {
		return 0
	}
	return set.dict.Len()
}

func (set *InstanceSet) ToSlice() []string {
	n := set.dict.Len()
	slice := make([]string, n)
	i := 0
	q := len(slice)
	set.dict.ForEach(func(key string, val any) bool {
		if i < q {
			slice[i] = key
		} else {
			slice = append(slice, key)
		}
		i++
		return true
	})
	return slice
}

func (set *InstanceSet) ForEach(consumer Consumer) {
	if set == nil || set.dict == nil {
		return
	}
	set.dict.ForEach(func(key string, val interface{}) bool {
		return consumer(key)
	})
}

func (set *InstanceSet) ShallowCopy() *InstanceSet {
	result := Make()
	set.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})
	return result
}

func (set *InstanceSet) RandomMembers(limit int) []string {
	if set == nil || set.dict == nil {
		return nil
	}
	return set.dict.RandomKeys(limit)
}

func (set *InstanceSet) RandomDistinctMembers(limit int) []string {
	return set.dict.RandomDistinctKeys(limit)
}

func Intersect(sets ...*Set) Set {
	var res Set
	res = Make()
	m := make(map[string]int)
	for _, set := range sets {
		(*set).ForEach(func(member string) bool {
			m[member]++
			return true
		})
	}
	for k, i := range m {
		if i == len(sets) {
			res.Add(k)
		}
	}
	return res
}

func Union(sets ...*Set) Set {
	var res Set
	res = Make()
	for _, set := range sets {
		(*set).ForEach(func(member string) bool {
			res.Add(member)
			return true
		})
	}
	return res
}

func Diff(sets ...*InstanceSet) *InstanceSet {
	res := sets[0].ShallowCopy()
	for i := 1; i < len(sets); i++ {
		sets[i].ForEach(func(member string) bool {
			res.Remove(member)
			return true
		})
		if res.Len() == 0 {
			break
		}
	}
	return res
}
