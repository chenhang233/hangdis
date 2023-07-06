package pubsub

import (
	List "hangdis/datastruct/list"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
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
	raw, exists := hub.subs.Get(channel)
	var subscribers List.List
	if exists {
		subscribers = raw.(*List.LinkedList)
	} else {
		subscribers = List.MakeLinked()
	}
	if subscribers.Contains(func(a any) bool {
		return utils.Equals(a, client)
	}) {
		return false
	}
	subscribers.Add(client)
	return true
}

func unsubscribe0(hub *Hub, channel string, client redis.Connection) bool {
	client.UnSubscribe(channel)
	raw, exists := hub.subs.Get(channel)
	if !exists {
		return false
	}
	var subscribers List.List
	subscribers = raw.(*List.LinkedList)
	subscribers.RemoveAllByVal(func(a any) bool {
		return utils.Equals(a, client)
	})
	if subscribers.Len() == 0 {
		hub.subs.Remove(channel)
	}
	return true
}

func UnsubscribeAll(hub *Hub, c redis.Connection) {}

func UnSubscribe(db *Hub, c redis.Connection, args [][]byte) redis.Reply {}

func Publish(hub *Hub, args [][]byte) redis.Reply {}

func Subscribe(hub *Hub, c redis.Connection, args [][]byte) redis.Reply {}
