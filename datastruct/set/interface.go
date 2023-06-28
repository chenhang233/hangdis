package set

type Consumer func(member string) bool
type Set interface {
	Add(val string) int
	Remove(val string) int
	Has(val string) bool
	Len() int
	ToSlice() []string
	ForEach(consumer Consumer)
}
