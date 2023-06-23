package bitmap

type BitMap []byte

func New() *BitMap {
	b := BitMap(make([]byte, 0))
	return &b
}

func FromBytes(bytes []byte) *BitMap {
	bm := BitMap(bytes)
	return &bm
}

func (b *BitMap) ToBytes() []byte {
	return *b
}

func (b *BitMap) toByteSize(bitSize int64) int64 {
	if bitSize%8 == 0 {
		return bitSize / 8
	}
	return bitSize/8 + 1
}

func (b *BitMap) grow(bitSize int64) {
	size := b.toByteSize(bitSize)
	gap := size - int64(len(*b))
	if gap <= 0 {
		return
	}
	*b = append(*b, make([]byte, gap)...)
}

func (b *BitMap) Get(offset int64) byte {
	bitIndex := offset / 8
	bitOffset := offset % 8
	if bitIndex >= int64(len(*b)) {
		return 0
	}
	return ((*b)[bitIndex] >> bitOffset) & 0x01
}

func (b *BitMap) Set(offset int64, val byte) {
	bitIndex := offset / 8
	bitOffset := offset % 8
	mask := byte(1 << bitOffset)
	b.grow(offset + 1)
	if val > 0 {
		(*b)[bitIndex] |= mask
	} else {
		(*b)[bitIndex] &^= mask
	}
}
