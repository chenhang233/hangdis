package clients

import (
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
