package database

import (
	"context"
	"sync"
)

type slaveStatus struct {
	mutex  sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func (repl *slaveStatus) stopSlaveWithMutex() {
}

func (server *Server) stopSlaveStatus() {

}

func (server *Server) slaveCron() {

}