package parser

import (
	"bufio"
	"bytes"
	"errors"
	"hangdis/interface/redis"
	"hangdis/redis/protocol"
	"hangdis/utils/logs"
	"io"
	"strconv"
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
		case '+': // representative ping/pong reply
			content := string(line[1:])
			ch <- &Payload{
				Data: protocol.MakeStatusReply(content),
			}
		case '-': // representative error reply
			ch <- &Payload{Data: protocol.MakeErrReply(string(line[1:]))}
		case '$': // representative get reply/null reply
			err = parseBulkString(line, reader, ch)
			if err != nil {
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
		case '*': // representative MultiBulkReply
			err = parseArray(line, reader, ch)
			if err != nil {
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
		default:
			args := bytes.Split(line, []byte{' '})
			ch <- &Payload{Data: protocol.MakeMultiBulkReply(args)}
		}
	}
}

func parseArray(line []byte, reader *bufio.Reader, ch chan<- *Payload) error {
	num, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		return err
	} else if num < 0 || num == 0 {
		return errors.New("illegal array header " + string(line[1:]))
	}
	lines := make([][]byte, 0, num)
	for i := int64(0); i < num; i++ {
		line, err := reader.ReadBytes('\n')
		l := len(line)
		if err != nil {
			return err
		}
		if l < 4 || line[0] != '$' || line[l-2] != '\r' {
			return errors.New("illegal bulk string header" + string(line))
		}
		strLen, err := strconv.ParseInt(string(line[1:l-2]), 10, 64)
		if err != nil {
			return err
		}
		if strLen < -1 {
			return errors.New("illegal bulk string header" + string(line))
		} else if strLen == 0 || strLen == -1 {
			lines = append(lines, []byte{})
		} else {
			body := make([]byte, strLen+2)
			_, err = io.ReadFull(reader, body)
			if err != nil {
				return err
			}
			lines = append(lines, body[:len(body)-2])
		}
	}
	ch <- &Payload{Data: protocol.MakeMultiBulkReply(lines)}
	return nil
}

func parseBulkString(header []byte, reader *bufio.Reader, ch chan<- *Payload) error {
	num, err := strconv.ParseInt(string(header[1:]), 10, 64)
	if err != nil {
		return err
	} else if num < 0 || num == 0 {
		return errors.New("illegal array header " + string(header[1:]))
	}
	body := make([]byte, num+2)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	ch <- &Payload{Data: protocol.MakeBulkReply(body[:len(body)-2])}
	return nil
}
