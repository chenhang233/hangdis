package list

import "container/list"

const pageSize = 1024

type QuickList struct {
	data *list.List
	size int // User data quantity
}

type iterator struct {
	node   *list.Element
	offset int
	ql     *QuickList
}

func (iter *iterator) page() []any {
	return iter.node.Value.([]any)
}

func (iter *iterator) get() any {
	return iter.page()[iter.offset]
}
func (iter *iterator) next() bool {
	page := iter.page()
	if iter.offset < len(page)-1 {
		iter.offset++
		return true
	}
	if iter.node == iter.ql.data.Back() {
		iter.offset = len(page)
		return false
	}
	iter.node = iter.node.Next()
	iter.offset = 0
	return true
}
func (iter *iterator) prev() bool {
	page := iter.page()
	if iter.offset > 0 {
		iter.offset--
		return true
	}
	if iter.node == iter.ql.data.Front() {
		iter.offset = -1
		return false
	}
	iter.node = iter.node.Prev()
	iter.offset = len(iter.node.Value.([]any)) - 1
	return true
}
func (iter *iterator) atEnd() bool {
	if iter.ql.data.Len() == 0 {
		return true
	}
	if iter.node != iter.ql.data.Back() {
		return false
	}
	return iter.offset == len(iter.page())
}
func (iter *iterator) atBegin() bool {
	if iter.ql.data.Len() == 0 {
		return true
	}
	if iter.node != iter.ql.data.Front() {
		return false
	}
	return iter.offset == -1
}
func (iter *iterator) set(val any) {
	page := iter.page()
	page[iter.offset] = val
}
func (iter *iterator) remove() any {
	page := iter.page()
	val := page[iter.offset]
	page = append(page[:iter.offset], page[iter.offset+1:]...)
	if len(page) > 0 {
		iter.node.Value = page
		if len(page) == iter.offset {
			if iter.node != iter.ql.data.Back() {
				iter.node = iter.node.Next()
				iter.offset = 0
			}
		}
	} else {
		if iter.node == iter.ql.data.Back() {
			iter.ql.data.Remove(iter.node)
			iter.node = nil
			iter.offset = 0
		} else {
			next := iter.node.Next()
			iter.ql.data.Remove(iter.node)
			iter.node = next
			iter.offset = 0
		}
	}
	iter.ql.size--
	return val
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
