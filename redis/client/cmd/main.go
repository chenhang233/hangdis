package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"hangdis/interface/redis"
	"hangdis/redis/client"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var addr *string
var nullBulkBytes = []byte("$-1\r\n")

func init() {
	addr = flag.String("addr", "127.0.0.1:8888", "bind addr")
}

func main() {
	c, err := client.MakeClient(*addr)
	if err != nil {
		panic(err)
	}
	fmt.Println(utils.Green(fmt.Sprintf("tcp connection establishment addr: %s", *addr)))
	c.Start()
	fmt.Println(utils.White("Please enter the command"))
	for {
		reader := bufio.NewReader(os.Stdin)
		bs, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println(utils.Red(err.Error()))
			continue
		}
		cmd := string(bs[:len(bs)-2])
		cmd = strings.Trim(cmd, " ")
		if cmd == "quit" || cmd == "exit" {
			break
		}
		if cmd == "clear" || cmd == "cls" {
			command := exec.Command("cmd", "/c", "cls")
			command.Stdout = os.Stdout
			err := command.Run()
			if err != nil {
				fmt.Println(err)
			}
			continue
		}
		list := strings.Split(cmd, " ")
		n := len(list)
		bys := make([][]byte, n)
		for i := 0; i < n; i++ {
			str := list[i]
			bys[i] = make([]byte, len(str))
			bys[i] = []byte(str)
		}
		parseReplyType(c.Send(bys))
	}
	c.Close()
}

func parseReplyType(response redis.Reply) {
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
	default:
		fmt.Println(fmt.Sprintf("can not parse %s", utils.Purple(string(response.ToBytes()))))
	}
}
