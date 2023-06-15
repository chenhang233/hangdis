package clients

import (
	"bufio"
	"bytes"
	"fmt"
	"hangdis/redis/protocol"
	"net"
	"testing"
)

func TestTcp(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		println(err)
		return
	}
	accept, err := listen.Accept()
	println(accept)
}

func TestClientDial(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func() {
			dial, err := net.Dial("tcp", "127.0.0.1:8888")
			if err != nil {
				println(err)
				return
			}
			str := "hello world"
			dial.Write([]byte(str))
		}()
	}
	for {

	}
}

func TestMultiBulkReply(t *testing.T) {

	mu := &protocol.MultiBulkReply{
		Args: [][]byte{
			[]byte("INFO"),
		},
	}
	b := mu.ToBytes()
	reader := bytes.NewReader(b)
	r := bufio.NewReader(reader)
	line, _ := r.ReadBytes('\n')
	fmt.Println(line, "line", string(line), line[0] == '*')
	args := bytes.Split(line, []byte{' '})
	fmt.Println(args)
	protocol.MakeMultiBulkReply(args)
}
