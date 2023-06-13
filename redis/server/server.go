package server

import (
	"context"
	"hangdis/interface/database"
	"net"
	"sync"
)

type Handler struct {
	activeConn sync.Map // *client -> placeholder
	db         database.DB
	closing    bool // refusing new client and new request
}

func MakeHandler() *Handler {
	var db database.DB

	h := &Handler{
		db: db,
	}
	return h
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn) {}

func (h *Handler) Close() error {
	return nil
}
