package protocol

import (
	"bytes"
	"fmt"
	"strconv"
)

var (
	CRLF                 = "\r\n"
	UnknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type MultiBulkReply struct {
	Args [][]byte
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

/*
ToBytes
SET a a 传输格式: *3 $3 SET $1 a $1 a
*/
func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	fmt.Println(buf.String(), "   38")
	return buf.Bytes()
}

type EmptyMultiBulkReply struct{}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}

func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return []byte("empty")
}

type PongReply struct{}

var pongBytes = []byte("+PONG\r\n")

func MakePongReply() *PongReply {
	return &PongReply{}
}

func (r *PongReply) ToBytes() []byte {
	return pongBytes
}

type StandardErrReply struct {
	Status string
}

func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}
