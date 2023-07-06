package pubsub

import (
	List "hangdis/datastruct/list"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"hangdis/utils/logs"
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

func UnsubscribeAll(hub *Hub, c redis.Connection) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for _, channel := range c.GetChannels() {
		unsubscribe0(hub, channel, c)
	}
}
func Subscribe(hub *Hub, c redis.Connection, args [][]byte) redis.Reply {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	channels := make([]string, len(args))
	for i, b := range args {
		channels[i] = string(b)
	}
	for _, channel := range channels {
		if subscribe0(hub, channel, c) {
			_, err := c.Write(makeMsg(_subscribe, channel, int64(c.SubsCount())))
			if err != nil {
				logs.LOG.Warn.Println(err)
			}
		}
	}
	return protocol.MakeEmptyMultiBulkReply()
}

func UnSubscribe(hub *Hub, c redis.Connection, args [][]byte) redis.Reply {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	var channels []string
	if len(args) > 0 {
		channels = make([]string, len(args))
		for i, b := range args {
			channels[i] = string(b)
		}
	} else {
		channels = c.GetChannels()
	}
	if len(channels) == 0 {
		_, _ = c.Write(unSubscribeNothing)
		return protocol.MakeEmptyMultiBulkReply()
	}
	for _, channel := range channels {
		if unsubscribe0(hub, channel, c) {
			_, err := c.Write(makeMsg(_unsubscribe, channel, int64(c.SubsCount())))
			if err != nil {
				logs.LOG.Warn.Println(err)
			}
		}
	}
	return protocol.MakeEmptyMultiBulkReply()
}

func Publish(hub *Hub, args [][]byte) redis.Reply {
	if len(args) != 2 {
		return protocol.MakeErrReply("publish args")
	}
	channel := string(args[0])
	message := args[1]
	hub.mu.Lock()
	defer hub.mu.Unlock()
	raw, exists := hub.subs.Get(channel)
	if !exists {
		return protocol.MakeIntReply(0)
	}
	var subscribers List.List
	subscribers = raw.(*List.LinkedList)
	subscribers.ForEach(func(i int, v any) bool {
		c := v.(redis.Connection)
		replyArgs := make([][]byte, 3)
		replyArgs[0] = messageBytes
		replyArgs[1] = []byte(channel)
		replyArgs[2] = message
		_, err := c.Write(protocol.MakeMultiBulkReply(replyArgs).ToBytes())
		if err != nil {
			logs.LOG.Warn.Println(err)
		}
		return true
	})
	return protocol.MakeIntReply(int64(subscribers.Len()))
}
