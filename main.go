package main

import (
	"fmt"
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
	log, err := logs.LoadLog(logs.ServerLogPath)
	if err != nil {
		panic(err)
	}
	config.SetupConfig("hangdis.conf")
	p := config.Properties
	sf := fmt.Sprintf("ip: %s port: %d", p.Bind, p.Port)
	sf2 := fmt.Sprintf("RuntimeID: %s MaxClients: %d AbsPath: %s", p.RuntimeID, p.MaxClients, p.AbsPath)
	log.Debug.Println(sf)
	log.Debug.Println(sf2)
	config := &tcp.Config{
		Log:  log,
		Name: "hangdis",
	}
	tcp.ListenAndServeWithSignal(config, server.MakeHandler())

	s, err := tcp.New()
	if err != nil {
		panic(err)
	}
	s.Log.Info.Println("server close...")
}
