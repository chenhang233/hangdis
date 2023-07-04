package test

import "testing"

func TestMap(t *testing.T) {
	m := map[string]bool{}
	delete(m, "ff")
	delete(m, "ff1")
	delete(m, "ff2")
	delete(m, "ff3")
}
