package client

import (
	"bytes"
	"fmt"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"strconv"
)

var (
	nullBulkBytes = []byte("$-1\r\n")
)

func ParseInputString(input string) []string {
	bys := []byte(input)
	n := len(bys)
	var list []string
	left, right := 0, 0
	for i := 0; i < n; i++ {
		if bys[i] == '"' {
			i++
			left = i
			for i < n && bys[i] != '"' {
				i++
			}
			right = i
			list = append(list, string(bys[left:right]))
		} else if bys[i] != ' ' {
			left = i
			for i < n && bys[i] != ' ' {
				i++
			}
			right = i
			list = append(list, string(bys[left:right]))
		}
	}
	return list
}

func ParseReplyType(response redis.Reply) {
	bys := response.ToBytes()
	if bytes.Compare(bys, nullBulkBytes) == 0 {
		//reply := response.(*protocol.EmptyMultiBulkReply)
		fmt.Println(utils.Red("null"))
		return
	}
	switch bys[0] {
	case '+':
		reply := response.(*protocol.StandardStatusReply)
		fmt.Println(fmt.Sprintf("+ %s", utils.Blue(reply.Status)))
	case '-':
		reply := response.(*protocol.StandardErrReply)
		fmt.Println(fmt.Sprintf("- %s", utils.Blue(reply.Status)))
	case '$':
		reply := response.(*protocol.BulkReply)
		fmt.Println(fmt.Sprintf("$ %s", utils.Blue(string(reply.Arg))))
	case ':':
		reply := response.(*protocol.IntReply)
		fmt.Println(fmt.Sprintf(": %s", utils.Blue(strconv.FormatInt(reply.Code, 10))))
	case '*':
		reply := response.(*protocol.MultiBulkReply)
		n := len(reply.Args)
		for i := 0; i < n; i++ {
			t := string(reply.Args[i])
			if len(t) == 0 {
				t = "null"
			}
			fmt.Println(fmt.Sprintf("* %d: %s", i, utils.Yellow(t)))
		}
	default:
		fmt.Println(fmt.Sprintf("can not parse %s", utils.Purple(string(response.ToBytes()))))
	}
}
