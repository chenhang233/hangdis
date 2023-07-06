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

func Make(vals ...any) *LinkedList {}

func (list *LinkedList) find(index int) (n *node) {}

func (list *LinkedList) Add(val any) {}

func (list *LinkedList) Get(index int) (val any)   {}
func (list *LinkedList) Set(index int, val any)    {}
func (list *LinkedList) Insert(index int, val any) {}

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
