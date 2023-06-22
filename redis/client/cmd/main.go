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

var (
	addr          *string
	nullBulkBytes = []byte("$-1\r\n")
	cmdTable      = make(map[string]*Command)
)

type ExecFunc func(*client.Client) error

type Command struct {
	executor ExecFunc
}

func RegisterCMD(name string, executor ExecFunc) {
	name = strings.ToLower(name)
	cmdTable[name] = &Command{
		executor: executor,
	}
}

func matchCMD(c *client.Client, name string) bool {
	name = strings.ToLower(name)
	command := cmdTable[name]
	if command != nil {
		e := command.executor
		if e == nil {
			panic("executor not found")
		}
		_ = e(c)
		return true
	}
	return false
}

func quit(c *client.Client) error {
	return c.Close()
}

func clear(c *client.Client) error {
	command := exec.Command("cmd", "/c", "cls")
	command.Stdout = os.Stdout
	err := command.Run()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func init() {
	addr = flag.String("addr", "127.0.0.1:8888", "bind addr")
	RegisterCMD("quit", quit)
	RegisterCMD("exit", quit)
	RegisterCMD("clear", clear)
	RegisterCMD("cls", clear)
}

func main() {
	c, err := client.MakeClient(*addr)
	if err != nil {
		panic(err)
	}
	fmt.Println(utils.Green(fmt.Sprintf("tcp connection establishment addr: %s", *addr)))
	c.Start()
	fmt.Println(utils.White("Please enter the command"))
	reader := bufio.NewReader(os.Stdin)
	for {
		if c.Status == client.STOP {
			fmt.Println(utils.Yellow("Exit signal "))
			break
		}
		bs, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println(utils.Red(err.Error()))
			continue
		}
		cmd := string(bs[:len(bs)-2])
		cmd = strings.Trim(cmd, " ")
		f := matchCMD(c, cmd)
		if f {
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
