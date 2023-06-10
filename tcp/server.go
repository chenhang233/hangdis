package tcp

import (
	"fmt"
	"hangdis/config"
	"hangdis/utils/logs"
	"io"
	"net"
)

type Server struct {
	Name      string
	config    *config.ServerConfig
	Log       *logs.LogConf
	RunServer net.Listener
}

func New() (*Server, error) {
	log, err := logs.LoadLog(logs.ServerLogPath)
	if err != nil {
		panic(err)
	}
	server := &Server{
		Log:  log,
		Name: "hangdis",
	}
	sc := config.SetupConfig("hangdis.conf")
	server.config = sc
	sf := fmt.Sprintf("ip: %s port: %d", sc.Bind, sc.Port)
	sf2 := fmt.Sprintf("RuntimeID: %s MaxClients: %d AbsPath: %s", sc.RuntimeID, sc.MaxClients, sc.AbsPath)
	server.Log.Debug.Println(sf)
	server.Log.Debug.Println(sf2)
	return server, server.new()
}

func (s *Server) new() error {
	server, err := net.Listen("tcp", s.config.BindAddr)
	if err != nil {
		s.Log.Error.Println(err)
		return err
	}
	s.RunServer = server
	s.handleClientRequest()
	return nil
}

func (s *Server) handleClientRequest() {
	for {
		accept, err := s.RunServer.Accept()
		if err != nil {
			s.Log.Warn.Println(err)
			continue
		}
		go s.Read(&accept)
	}
}

func (s *Server) Read(c *net.Conn) {
	defer (*c).Close()
	all, err := io.ReadAll(*c)
	if err != nil {
		s.Log.Error.Println(err)
		return
	}
	fmt.Println(all)
	fmt.Println(string(all))
}

func (s *Server) Write(c *net.Conn) {

}
