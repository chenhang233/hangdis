package tcp

import (
	"fmt"
	"hangdis/interface/tcp"
	"hangdis/utils/logs"
	"io"
	"net"
	"time"
)

type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
	Log        *logs.LogConf
	Name       string
}

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) {

}

func New() (*Server, error) {

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
