package tcp

import (
	"fmt"
	"hangdis/config"
	"net"
)

type Server struct {
	name string
	ip   net.IP
	port uint16
}

func (s *Server) New() {
	setupConfig := config.SetupConfig("hangdis.conf")
	sf := fmt.Sprintf("ip: %s port: %d", setupConfig.Bind, setupConfig.Port)

}
