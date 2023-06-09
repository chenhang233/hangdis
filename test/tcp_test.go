package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
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

func TestContext(t *testing.T) {
	ctx := context.Background()
	go func() {
		time.Sleep(time.Second)
		//done := ctx.Done()
	}()
	select {
	case <-ctx.Done():
		fmt.Println("done")
	}
}

func TestByte(t *testing.T) {
	const a = 66
	bys := []byte{a}
	fmt.Println(bys, string(bys))
}
