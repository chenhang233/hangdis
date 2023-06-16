package database

import (
	"fmt"
	"hangdis/config"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/tcp"
	"os"
	"runtime"
	"strings"
)

func Ping(c redis.Connection, args [][]byte) redis.Reply {
	if len(args) == 1 {
		return protocol.MakePongReply()
	} else if len(args) == 2 {
		return protocol.MakeErrReply(string(args[0]))
	} else {
		return protocol.MakeEmptyMultiBulkReply()
	}
}

func Info(c redis.Connection, args [][]byte) redis.Reply {
	var bys []byte
	if len(args) == 1 {
		items := []string{"server", "client"}
		for _, item := range items {
			bys = append(bys, InfoString(item)...)
		}
		return protocol.MakeBulkReply(bys)
	} else if len(args) == 2 {
		section := strings.ToLower(string(args[1]))
		switch section {
		case "server":
			bys = append(bys, InfoString("server")...)
		case "client":
			bys = append(bys, InfoString("client")...)
		}
		return protocol.MakeBulkReply(bys)
	} else {
		return protocol.MakeErrReply("ERR wrong number of arguments for 'info' command")
	}
}

func InfoString(section string) []byte {
	s := ""
	switch section {
	case "server":
		s = fmt.Sprintf("# Server\r\n"+
			"godis_version:%s\r\n"+
			"go_version:%s\r\n"+
			"process_id:%d\r\n"+
			"run_id:%s\r\n"+
			"tcp_bind:%s\r\n"+
			"config_file:%s\r\n"+
			"operating system: %s\r\n"+
			"instruction set: %s\r\n",
			"0.0.1",
			runtime.Version(),
			os.Getpid(),
			config.Properties.RuntimeID,
			config.Properties.BindAddr,
			config.Properties.AbsPath,
			runtime.GOOS, runtime.GOARCH,
		)
	case "client":
		s = fmt.Sprintf("# Clients\r\n"+
			"connected_clients:%d\r\n",
			tcp.ClientCounter,
		)
	}
	return []byte(s)
}

func Auth(c redis.Connection, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return protocol.MakeErrReply("ERR wrong number of arguments for 'auth' command")
	}
	p := config.Properties.Password
	if p == "" {
		return protocol.MakeStatusReply("No password")
	} else if string(args[1]) != p {
		return protocol.MakeErrReply("ERR invalid password")
	} else {
		c.SetPassword(p)
		return protocol.MakeOkReply()
	}
}
