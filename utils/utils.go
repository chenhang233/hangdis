package utils

import (
	"bytes"
)

func ConvertRange(start int64, end int64, size int64) (int, int) {
	if start < -size {
		return -1, -1
	} else if start < 0 {
		start = size + start
	} else if start >= size {
		return -1, -1
	}
	if end < -size {
		return -1, -1
	} else if end < 0 {
		end = size + end + 1
	} else if end < size {
		end = end + 1
	} else {
		end = size
	}
	if start > end {
		return -1, -1
	}
	return int(start), int(end)
}

func Equals(a any, b any) bool {
	b1, ok1 := a.([]byte)
	b2, ok2 := b.([]byte)
	if ok1 && ok2 {
		return bytes.Equal(b1, b2)
	}
	return a == b
}
