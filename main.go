package main

import (
	"hangdis/config"
	"hangdis/redis/server"
	"hangdis/tcp"
	"hangdis/utils"
	"hangdis/utils/logs"
)

const banner = `
##################################################
			hangDis	       
##################################################
`

func main() {
	print(utils.Red(banner))
	config.SetupConfig("hangdis.conf")
	c := &tcp.Config{
		Name:    "hangdis",
		Address: config.Properties.BindAddr,
	}
	err := tcp.ListenAndServeWithSignal(c, server.MakeHandler())
	if err != nil {
		logs.LOG.Error.Println("main: tcp.ListenAndServeWithSignal")
		logs.LOG.Error.Println(err)
	}
}
