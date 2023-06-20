package protocol

import (
	"bytes"
	"strconv"
)

var (
	CRLF                 = "\r\n"
	UnknownErrReplyBytes = []byte("-ERR unknown\r\n")
	nullBulkBytes        = []byte("$-1\r\n")
)

type BulkReply struct {
	Arg []byte
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}
func (r *BulkReply) ToBytes() []byte {
	if r.Arg == nil {
		return nullBulkBytes
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

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
	return buf.Bytes()
}

type EmptyMultiBulkReply struct{}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}
func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return nullBulkBytes
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

type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-Err syntax error\r\n")

func MakeSyntaxErrReply() *SyntaxErrReply {
	return &SyntaxErrReply{}
}

// ToBytes marshals redis.Reply
func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

type StandardStatusReply struct {
	Status string
}

func MakeStatusReply(status string) *StandardStatusReply {
	return &StandardStatusReply{
		Status: status,
	}
}
func (r *StandardStatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

type OkReply struct{}

var okBytes = []byte("+OK\r\n")

// ToBytes marshal redis.Reply
func (r *OkReply) ToBytes() []byte {
	return okBytes
}
func MakeOkReply() *OkReply {
	return &OkReply{}
}

type IntReply struct {
	Code int64
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}
