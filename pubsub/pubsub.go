package pubsub

import (
	"hangdis/datastruct/list"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"strconv"
)

var (
	_subscribe         = "subscribe"
	_unsubscribe       = "unsubscribe"
	messageBytes       = []byte("message")
	unSubscribeNothing = []byte("*3\r\n$11\r\nunsubscribe\r\n$-1\n:0\r\n")
)

func makeMsg(t string, channel string, code int64) []byte {
	return []byte("*3\r\n$" + strconv.FormatInt(int64(len(t)), 10) + protocol.CRLF + t + protocol.CRLF +
		"$" + strconv.FormatInt(int64(len(channel)), 10) + protocol.CRLF + channel + protocol.CRLF +
		":" + strconv.FormatInt(code, 10) + protocol.CRLF)
}

func subscribe0(hub *Hub, channel string, client redis.Connection) bool {
	client.Subscribe(channel)
	val, exists := hub.subs.Get(channel)
	var subscribers *list.LinkedList
	if exists {

	} else {

	}
}

func unsubscribe0(hub *Hub, channel string, client redis.Connection) bool {}

func UnsubscribeAll(hub *Hub, c redis.Connection) {}

func UnSubscribe(db *Hub, c redis.Connection, args [][]byte) redis.Reply {}

func Publish(hub *Hub, args [][]byte) redis.Reply {}

func Subscribe(hub *Hub, c redis.Connection, args [][]byte) redis.Reply {}
