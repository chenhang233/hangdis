package test

import (
	"os"
	"testing"
)

func TestMap(t *testing.T) {
	m := map[string]bool{}
	delete(m, "ff")
	delete(m, "ff1")
	delete(m, "ff2")
	delete(m, "ff3")
}

func TestTempFile(t *testing.T) {
	temp, err := os.CreateTemp("tmp", "*.aof")
	if err != nil {
		panic(err)
	}
	temp.WriteString("hello world")
	temp.Close()
}
