package main

import (
	"flag"
	"fmt"
	"hangdis/redis/client"
	"hangdis/utils"
	"strings"
)

var addr *string

func init() {
	addr = flag.String("addr", "127.0.0.1:8888", "bind addr")
}

func main() {
	c, err := client.MakeClient(*addr)
	if err != nil {
		panic(err)
	}
	utils.Green(fmt.Sprintf("tcp connection establishment addr: %s", *addr))
	c.Start()
	fmt.Println("Please enter the command")
	var cmd string
	for {
		_, err := fmt.Scanln(&cmd)
		if err != nil {
			fmt.Println(err)
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
		c.Send(bys)
	}

}
