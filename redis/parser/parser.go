package parser

import (
	"bufio"
	"bytes"
	"errors"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils/logs"
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
		l := len(line)
		if l <= 2 || line[l-2] != '\r' {
			logs.LOG.Debug.Println(string(line))
			continue
		}
		line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
		switch line[0] {
		case '-':
			ch <- &Payload{Data: protocol.MakeErrReply(string(line[1:]))}
		case '*':
			err := parseArray(line, rawReader, ch)
			if err != nil {
				ch <- &Payload{Err: err}
				return
			}
		default:
			args := bytes.Split(line, []byte{' '})
			ch <- &Payload{Data: protocol.MakeMultiBulkReply(args)}
		}
	}
}

func parseArray(line []byte, reader io.Reader, ch chan<- *Payload) error {

}

func protocolError(ch chan<- *Payload, msg string) {
	err := errors.New("protocol error: " + msg)
	ch <- &Payload{Err: err}
}
