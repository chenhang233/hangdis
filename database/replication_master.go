package database

import "sync"

const (
	masterRole = iota
	slaveRole
)

type masterStatus struct {
	mu     sync.RWMutex
	replId string
}
