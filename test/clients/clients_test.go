package clients

import (
	"bufio"
	"bytes"
	"fmt"
	"hangdis/redis/protocol"
	"io"
	"net"
	"testing"
	"time"
)

func TestTcp(t *testing.T) {
	listen, err := net.Listen("tcp", "127.0.0.1:8888")
	if err != nil {
		println(err)
		return
	}
	accept, err := listen.Accept()
	println(accept)
}

func TestClientDial(t *testing.T) {
	//for i := 0; i < 5; i++ {
	//	go func() {
	//
	//	}()
	//}
	dial, _ := net.Dial("tcp", "127.0.0.1:8888")
	mu := &protocol.MultiBulkReply{
		Args: [][]byte{
			[]byte("SET"),
			[]byte("a"),
			[]byte("a"),
		},
	}
	mu = &protocol.MultiBulkReply{
		Args: [][]byte{
			[]byte("PING"),
			[]byte("哈哈哈哈"),
		},
	}
	mu = &protocol.MultiBulkReply{
		Args: [][]byte{
			[]byte("INFO"),
		},
	}
	mu = &protocol.MultiBulkReply{
		Args: [][]byte{
			[]byte("AUTH"),
			[]byte("hang"),
		},
	}
	b := mu.ToBytes()
	go func() {
		reader := bufio.NewReader(dial)
		for {
			readBytes, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Println(err, "---")
				break
			}
			fmt.Println(string(readBytes))
		}
	}()
	for {
		time.Sleep(time.Second * 2)
		dial.Write(b)
	}
}

func TestAuth(t *testing.T) {
	dial, _ := net.Dial("tcp", "127.0.0.1:8888")
	mu := &protocol.MultiBulkReply{
		Args: [][]byte{
			[]byte("INFO"),
		},
	}
	b := mu.ToBytes()
	go func() {
		reader := bufio.NewReader(dial)
		for {
			readBytes, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Println(err, "---")
				break
			}
			fmt.Println(string(readBytes))
		}
	}()
	for {
		time.Sleep(time.Second * 2)
		dial.Write(b)
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

func TestReadFull(t *testing.T) {
	mu := &protocol.MultiBulkReply{
		Args: [][]byte{
			[]byte("SET"),
			[]byte("a"),
			[]byte("a"),
		},
	}
	b := mu.ToBytes()
	reader := bytes.NewReader(b)
	bys := make([]byte, 3)
	io.ReadFull(reader, bys)
	fmt.Println(bys, " bys ", string(bys))
}
