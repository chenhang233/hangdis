package main

import (
	"hangdis/config"
	"hangdis/redis/server"
	"hangdis/tcp"
	"hangdis/utils/logs"
)

const banner = `
##################################################
			hangDis	       
##################################################
`

func main() {
	print(banner)
	config.SetupConfig("hangdis.conf")
	c := &tcp.Config{
		Name:    "hangdis",
		Address: config.Properties.BindAddr,
	}
	err := tcp.ListenAndServeWithSignal(c, server.MakeHandler())
	if err != nil {
		logs.LOG.Error.Println(err)
	}
}
