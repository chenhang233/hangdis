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

func TestA(t *testing.T) {
	arr := [][]int{
		{1, 1, 0},
		{1, 1, 0},
		{0, 0, 1},
	}

	for k, v := range arr[0] {
		fmt.Println(k, v, "--")
	}
	findCircleNum(arr)
}

func findCircleNum(isConnected [][]int) (ans int) {
	vis := make([]bool, len(isConnected))
	var dfs func(int)
	dfs = func(from int) {
		vis[from] = true
		for to, conn := range isConnected[from] {
			fmt.Println(to, conn, isConnected[from])
			if conn == 1 && !vis[to] {
				dfs(to)
			}
		}
	}
	for i, v := range vis {
		if !v {
			ans++
			dfs(i)
		}
	}
	return

}
