package clients

import (
	"fmt"
	"hangdis/redis/client"
	"hangdis/redis/protocol"
	"testing"
)

func TestClient(t *testing.T) {
	c, err := client.MakeClient("127.0.0.1:8888")
	if err != nil {
		t.Error(err)
	}
	c.Start()
	args := [][]byte{
		[]byte("SET"),
		[]byte("a"),
		[]byte("a"),
	}
	res := c.Send(args)
	if status, ok := res.(*protocol.StandardStatusReply); ok {
		fmt.Println(status.Status)
		if status.Status != "OK" {
			t.Error("`set` failed")
		}
	}
	res = c.Send([][]byte{
		[]byte("GET"),
		[]byte("a"),
	})
	if bulkRet, ok := res.(*protocol.BulkReply); ok {
		fmt.Println(string(bulkRet.Arg))
		if string(bulkRet.Arg) != "a" {
			t.Error("`get` failed")
		}
	}

	res = c.Send([][]byte{
		[]byte("SETNX"),
		[]byte("b"),
		[]byte("hello world"),
	})
	if res, ok := res.(*protocol.IntReply); ok {
		fmt.Println(res.Code)
		if res.Code != 1 {
			t.Error("`setnx` failed ")
		}
	}
	res = c.Send([][]byte{
		[]byte("SETNX"),
		[]byte("b"),
		[]byte("world hello"),
	})
	if res, ok := res.(*protocol.IntReply); ok {
		fmt.Println(res.Code)
		if res.Code != 0 {
			t.Error("`setnx` failed ")
		}
	}
	c.Close()

}
