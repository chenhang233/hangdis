package server

import (
	"context"
	database2 "hangdis/database"
	"hangdis/interface/database"
	"hangdis/redis/connection"
	"hangdis/redis/parser"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/logs"
	"io"
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
	stream := parser.ParseStream(conn)
	var err error
	for payload := range stream {
		if payload.Err != nil {
			if payload.Err == io.EOF || payload.Err == io.ErrUnexpectedEOF {
				h.closeClient(client)
				logs.LOG.Warn.Println("client closed " + utils.Purple(client.RemoteAddr().String()))
				return
			}
			reply := protocol.MakeErrReply(payload.Err.Error())
			_, err = client.Write(reply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logs.LOG.Warn.Println("client closed" + client.RemoteAddr().String())
				return
			}
			continue
		}
		if payload.Data == nil {
			logs.LOG.Warn.Println("empty payload data")
			continue
		}
		mu, ok := payload.Data.(*protocol.MultiBulkReply)
		if !ok {
			logs.LOG.Error.Println("require MultiBulkReply protocol")
		}
		exec := h.db.Exec(client, mu.Args)
		if exec != nil {
			_, err = client.Write(exec.ToBytes())
		} else {
			_, err = client.Write(protocol.UnknownErrReplyBytes)
		}

		if err != nil {
			logs.LOG.Error.Println(err)
		}
	}
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
