package list

type LinkedList struct {
	first *node
	last  *node
	size  int
}

type node struct {
	val  interface{}
	prev *node
	next *node
}

func MakeLinked(vals ...any) *LinkedList {
	list := &LinkedList{}
	for _, val := range vals {
		list.Add(val)
	}
	return list
}

func (list *LinkedList) Add(val any) {
	if list == nil {
		panic("list is nil")
	}
	v := &node{val: val}
	if list.last == nil {
		list.first = v
		list.last = v
	} else {
		prev := list.last
		v.prev = prev
		prev.next = v
		list.last = v
	}
	list.size++
}

func (list *LinkedList) find(index int) (n *node) {
	var cur *node
	if index < list.size/2 {
		cur = list.first
		for i := 0; i < index; i++ {
			cur = cur.next
		}
	} else {
		cur = list.last
		for i := list.size - 1; i > index; i-- {
			cur = cur.prev
		}
	}
	return cur
}

func (list *LinkedList) Get(index int) (val any) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index >= list.size {
		panic("index out of bound")
	}
	return list.find(index).val
}
func (list *LinkedList) Set(index int, val any) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index >= list.size {
		panic("index out of bound")
	}
	list.find(index).val = val
}
func (list *LinkedList) Insert(index int, val any) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index > list.size {
		panic("index out of bound")
	}
	if index == list.size {
		list.Add(val)
		return
	}
	pivot := list.find(index)
	cur := &node{
		val:  val,
		prev: pivot.prev,
		next: pivot,
	}
	if cur.prev == nil {
		list.first = cur
	} else {
		cur.prev.next = cur
	}
	pivot.prev = cur
	list.size++
	return
}

func (list *LinkedList) removeNode(n *node) {}

func (list *LinkedList) Remove(index int) (val any)                          {}
func (list *LinkedList) RemoveLast() (val any)                               {}
func (list *LinkedList) RemoveAllByVal(expected Expected) int                {}
func (list *LinkedList) RemoveByVal(expected Expected, count int) int        {}
func (list *LinkedList) ReverseRemoveByVal(expected Expected, count int) int {}
func (list *LinkedList) Len() int                                            {}
func (list *LinkedList) ForEach(consumer Consumer)                           {}
func (list *LinkedList) Contains(expected Expected) bool                     {}
func (list *LinkedList) Range(start int, stop int) []interface{}             {}
