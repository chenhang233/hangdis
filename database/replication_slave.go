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
