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
	ql.size++
	if ql.data.Len() == 0 {
		newpage := make([]any, 0, pageSize)
		newpage = append(newpage, val)
		ql.data.PushBack(newpage)
		return
	}
	back := ql.data.Back()
	values := back.Value.([]any)
	if len(values) == cap(values) {
		newpage := make([]any, 0, pageSize)
		newpage = append(newpage, val)
		ql.data.PushBack(newpage)
		return
	}
	values = append(values, val)
	back.Value = values
	return
}

func (ql *QuickList) find(index int) *iterator {
	if ql == nil {
		panic("list nil")
	}
	if index > ql.size || index < 0 {
		panic("index out of bound")
	}
	var n *list.Element
	var pageBegin, pageLen int
	if index < ql.size/2 {
		n = ql.data.Front()
		pageBegin = 0
		for {
			pageLen = len(n.Value.([]any))
			if pageBegin+pageLen > index {
				break
			}
			pageBegin += pageLen
			n = n.Next()
		}
	} else {
		n = ql.data.Back()
		pageBegin = ql.size
		for {
			pageLen = len(n.Value.([]any))
			pageBegin -= pageLen
			if pageBegin <= index {
				break
			}
			n = n.Prev()
		}
	}
	offset := index - pageBegin
	return &iterator{
		ql:     ql,
		offset: offset,
		node:   n,
	}
}
func (ql *QuickList) Get(index int) (val any) {
	find := ql.find(index)
	return find.get()
}
func (ql *QuickList) Set(index int, val any) {
	find := ql.find(index)
	find.set(val)
}
func (ql *QuickList) Insert(index int, val any) {
	if index == ql.size {
		ql.Add(val)
		return
	}
	find := ql.find(index)
	page := find.page()
	if len(page) < pageSize {
		page = append(page[:find.offset+1], page[find.offset:]...)
		page[find.offset] = val
		find.node.Value = page
		ql.size++
		return
	}
	after := page[pageSize/2:]
	newpage := make([]any, 0, len(after))
	newpage = append(newpage, after...)
	page = page[:pageSize/2]
	if find.offset < len(page) {
		page = append(page[:find.offset+1], page[find.offset:]...)
		page[find.offset] = val
	} else {
		i := find.offset - pageSize/2
		newpage = append(newpage[:i+1], page[i:]...)
		newpage[i] = val
	}
	find.node.Value = page
	find.ql.data.InsertAfter(newpage, find.node)
	ql.size++
}
func (ql *QuickList) Remove(index int) (val any) {
	iter := ql.find(index)
	return iter.remove()
}
func (ql *QuickList) RemoveLast() (val any) {
	if ql.Len() == 0 {
		return nil
	}
	ql.size--
	back := ql.data.Back()
	values := back.Value.([]any)
	n := len(values)
	if n == 1 {
		ql.data.Remove(back)
		return values[0]
	}
	val = values[n-1]
	values = append(values[:n-1])
	back.Value = values
	return
}
func (ql *QuickList) RemoveAllByVal(expected Expected) int {
	iter := ql.find(0)
	remove := 0
	for !iter.atEnd() {
		if expected(iter.get()) {
			iter.remove()
			remove++
		} else {
			iter.next()
		}
	}
	return remove
}
func (ql *QuickList) RemoveByVal(expected Expected, count int) int {
	if ql.size == 0 {
		return 0
	}
	iter := ql.find(0)
	remove := 0
	for !iter.atEnd() {
		if expected(iter.get()) {
			iter.remove()
			remove++
			if remove == count {
				break
			}
		} else {
			iter.next()
		}
	}
	return remove
}
func (ql *QuickList) ReverseRemoveByVal(expected Expected, count int) int {
	if ql.size == 0 {
		return 0
	}
	iter := ql.find(ql.size - 1)
	remove := 0
	for !iter.atBegin() {
		if expected(iter.get()) {
			iter.remove()
			remove++
			if remove == count {
				break
			}
		} else {
			iter.prev()
		}
	}
	return remove
}
func (ql *QuickList) Len() int {
	return ql.size
}
func (ql *QuickList) ForEach(consumer Consumer) {
	if ql == nil {
		panic("list is nil")
	}
	if ql.Len() == 0 {
		return
	}
	iter := ql.find(0)
	i := 0
	for {
		callback := consumer(i, iter.get())
		if !callback {
			break
		}
		i++
		if !iter.next() {
			break
		}
	}
}
func (ql *QuickList) Contains(expected Expected) bool {
	contains := false
	ql.ForEach(func(i int, actual interface{}) bool {
		if expected(actual) {
			contains = true
			return false
		}
		return true
	})
	return contains
}
func (ql *QuickList) Range(start int, stop int) []any {
	if start < 0 || start >= ql.Len() {
		panic("start out of range")
	}
	if stop < start || stop > ql.Len() {
		panic("stop out of range")
	}
	size := stop - start
	slice := make([]any, 0, size)
	iter := ql.find(start)
	i := 0
	for i < size {
		slice = append(slice, iter.get())
		i++
		iter.next()
	}
	return slice
}
