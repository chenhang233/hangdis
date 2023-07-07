package main

import (
	"bufio"
	"flag"
	"fmt"
	"hangdis/redis/client"
	"hangdis/utils"
	"os"
	"os/exec"
	"strings"
)

var (
	addr     *string
	cmdTable = make(map[string]*Command)
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
	err := c.Close()
	os.Exit(0)
	return err
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
	fmt.Println(utils.Green(fmt.Sprintf("tcp connection establishment addr: %s client: %s", *addr, c.Conn.LocalAddr())))
	c.Start()
	fmt.Println(utils.White("Please enter the command"))
	reader := bufio.NewReader(os.Stdin)
	for {
		bs, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println(utils.Red(err.Error()))
			continue
		}
		cmd := string(bs[:len(bs)-2])
		cmd = strings.Trim(cmd, " ")
		cmd = strings.ToLower(cmd)
		f := matchCMD(c, cmd)
		if f {
			continue
		}
		list := client.ParseInputString(cmd)
		bys := utils.ToCmdLine(list...)
		client.ParseReplyType(c.Send(bys))
	}
}
