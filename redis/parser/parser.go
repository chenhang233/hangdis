package parser

import (
	"bufio"
	"bytes"
	"errors"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"io"
)

type Payload struct {
	Data redis.Reply
	Err  error
}

func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse(reader, ch)
	return ch
}

func parse(rawReader io.Reader, ch chan<- *Payload) {
	reader := bufio.NewReader(rawReader)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			ch <- &Payload{Err: err}
			close(ch)
			return
		}
		if len(line) <= 2 {
			continue
		}
		line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
		switch line[0] {
		default:
			args := bytes.Split(line, []byte{' '})
			ch <- &Payload{Data: protocol.MakeMultiBulkReply(args)}
		}
	}
}

func protocolError(ch chan<- *Payload, msg string) {
	err := errors.New("protocol error: " + msg)
	ch <- &Payload{Err: err}
}
