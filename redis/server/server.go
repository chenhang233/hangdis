package server

import (
	"context"
	database2 "hangdis/database"
	"hangdis/interface/database"
	"hangdis/redis/connection"
	"hangdis/utils/logs"
	"net"
	"sync"
)

type Handler struct {
	activeConn sync.Map
	db         database.DB
	closing    bool
}

func MakeHandler() *Handler {
	var db database.DB
	db = database2.NewStandaloneServer()
	h := &Handler{
		db: db,
	}
	return h
}

func (h *Handler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing {
		_ = conn.Close()
		return
	}

	client := connection.NewConn(conn)
	h.activeConn.Store(client, struct{}{})
}

func (h *Handler) Close() error {
	logs.LOG.Debug.Println("handler close ...")
	h.closing = true
	h.activeConn.Range(func(key, value any) bool {
		c := key.(*connection.Connection)
		c.Close()
		return true
	})
	h.db.Close()
	return nil
}
