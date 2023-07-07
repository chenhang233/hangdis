package tcp

import (
	"context"
	"errors"
	"fmt"
	"hangdis/config"
	"hangdis/interface/tcp"
	"hangdis/utils"
	"hangdis/utils/logs"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Address    string
	MaxConnect int
	Timeout    time.Duration
	Name       string
}

var ClientCounter int

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
	A:
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			fmt.Println(sig.String(), "sig", syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
			closeChan <- struct{}{}
		}
		goto A
	}()
	cfg.MaxConnect = utils.GetConnNum(config.Properties.MaxClients)
	listen, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logs.LOG.Info.Println(utils.Green(fmt.Sprintf("bind: %s, start listening...", cfg.Address)))
	ListenAndServe(listen, handler, closeChan, cfg)
	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}, cfg *Config) {
	errCh := make(chan error, 1)
	defer close(errCh)
	go func() {
		select {
		case <-closeChan:
			logs.LOG.Debug.Print(utils.Purple("get exit signal"))
			_ = listener.Close()
			_ = handler.Close()
		case er := <-errCh:
			logs.LOG.Error.Println(fmt.Sprintf("accept error: %s", utils.Red(er.Error())))
		}
	}()
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			break
		}
		if ClientCounter >= cfg.MaxConnect {
			errCh <- errors.New("Over maximum connection ")
			continue
		}
		logs.LOG.Info.Println(fmt.Sprintf("client link: %s", utils.Blue(conn.RemoteAddr().String())))
		ClientCounter++
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
				ClientCounter--
				logs.LOG.Info.Println(fmt.Sprintf("client leave: %s", utils.Purple(conn.RemoteAddr().String())))
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}

//func New() (*Server, error) {
//
//}
//
//func (s *Server) new() error {
//	server, err := net.Listen("tcp", s.config.BindAddr)
//	if err != nil {
//		s.Log.Error.Println(err)
//		return err
//	}
//	s.RunServer = server
//	s.handleClientRequest()
//	return nil
//}
//
//func (s *Server) handleClientRequest() {
//	for {
//		accept, err := s.RunServer.Accept()
//		if err != nil {
//			s.Log.Warn.Println(err)
//			continue
//		}
//		go s.Read(&accept)
//	}
//}
//
//func (s *Server) Read(c *net.Conn) {
//	defer (*c).Close()
//	all, err := io.ReadAll(*c)
//	if err != nil {
//		s.Log.Error.Println(err)
//		return
//	}
//	fmt.Println(all)
//	fmt.Println(string(all))
//}
//
//func (s *Server) Write(c *net.Conn) {
//
//}
