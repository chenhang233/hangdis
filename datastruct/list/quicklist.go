package list

import "container/list"

const pageSize = 1024

type QuickList struct {
	data *list.List
	size int // data size
}

func NewQuickList() *QuickList {
	l := &QuickList{
		data: list.New(),
	}
	return l
}

func (ql *QuickList) Add(val any) {

}
func (ql *QuickList) Get(index int) (val any) {

}
func (ql *QuickList) Set(index int, val any) {

}
func (ql *QuickList) Insert(index int, val any) {

}
func (ql *QuickList) Remove(index int) (val any) {

}
func (ql *QuickList) RemoveLast() (val any) {

}
func (ql *QuickList) RemoveAllByVal(expected Expected) int {

}
func (ql *QuickList) RemoveByVal(expected Expected, count int) int {

}
func (ql *QuickList) ReverseRemoveByVal(expected Expected, count int) int {

}
func (ql *QuickList) Len() int {

}
func (ql *QuickList) ForEach(consumer Consumer) {

}
func (ql *QuickList) Contains(expected Expected) bool {

}
func (ql *QuickList) Range(start int, stop int) []any {

}