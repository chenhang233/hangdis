package set

import "hangdis/datastruct/dict"

type InstanceSet struct {
	dict dict.Dict
}

func Make(members ...string) *InstanceSet {
	set := &InstanceSet{
		dict: dict.MakeInstanceDict(),
	}

	return set
}

func (set *InstanceSet) Add(val string) int {}

func (set *InstanceSet) Remove(val string) int {}

func (set *InstanceSet) Has(val string) bool {}
func (set *InstanceSet) Len() int            {}

func (set *InstanceSet) ToSlice() []string {}

func (set *InstanceSet) ForEach(consumer Consumer) {}

func (set *InstanceSet) ShallowCopy() *InstanceSet {}

func Intersect(sets ...*InstanceSet) *InstanceSet {}

func Union(sets ...*InstanceSet) *InstanceSet {}

func Diff(sets ...*InstanceSet) *InstanceSet {}

func (set *InstanceSet) RandomMembers(limit int) []string {}

func (set *InstanceSet) RandomDistinctMembers(limit int) []string {}
