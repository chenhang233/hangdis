package client

import (
	"fmt"
	"hangdis/interface/redis"
	"hangdis/redis/parser"
	"hangdis/redis/protocol"
	"hangdis/utils"
	"io"
	"net"
	"sync"
)

type Client struct {
	Conn        net.Conn
	addr        string
	pendingReqs chan *request
	waitingReqs chan *request
	StopStatus  chan int
}

type request struct {
	args  [][]byte
	reply redis.Reply
	err   error
	wait  *sync.WaitGroup
}

const (
	chanSize = 256
	RUN      = 0
	STOP     = 1
)

func MakeClient(addr string) (*Client, error) {
	dial, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		StopStatus:  make(chan int, STOP),
		addr:        addr,
		Conn:        dial,
		pendingReqs: make(chan *request, chanSize),
		waitingReqs: make(chan *request, chanSize),
	}, nil
}

func (c *Client) Start() {
	go c.handleRead()
	go c.handleWrite()
}

func (c *Client) Close() error {
	close(c.pendingReqs)
	close(c.waitingReqs)
	close(c.StopStatus)
	return c.Conn.Close()
}
func (c *Client) Send(args [][]byte) redis.Reply {
	req := &request{args: args, wait: &sync.WaitGroup{}}
	req.wait.Add(1)
	c.pendingReqs <- req
	req.wait.Wait()
	return req.reply
}

func (c *Client) handleWrite() {
	for pending := range c.pendingReqs {
		c.doReq(pending)
	}
}

func (c *Client) doReq(req *request) {
	if req == nil || len(req.args) <= 0 {
		return
	}
	mb := protocol.MakeMultiBulkReply(req.args)
	bytes := mb.ToBytes()
	_, err := c.Conn.Write(bytes)
	if err != nil {
		fmt.Println(err)
		req.wait.Done()
	} else {
		c.waitingReqs <- req
	}
}

func (c *Client) handleRead() {
	ch := parser.ParseStream(c.Conn)
	for payload := range ch {
		fmt.Println(payload, "payload 92")
		if payload.Err != nil {
			if payload.Err == io.EOF {
				fmt.Println(utils.Purple("server closed"))
				c.StopStatus <- STOP
				return
			}
			fmt.Println("client.go handleRead payload.Err:", utils.Red(payload.Err.Error()))
			//c.Close()
			c.StopStatus <- STOP
			return
		}
		c.handlePayload(payload)
	}
	fmt.Println("结束了x")
}

func (c *Client) handlePayload(p *parser.Payload) {
	w := <-c.waitingReqs
	w.reply = p.Data
	w.wait.Done()
}

func (c *Client) WaitMsg() error {
	ch := parser.ParseStream(c.Conn)
	for payload := range ch {
		fmt.Println(payload, "payload")
		if payload.Err != nil {
			if payload.Err == io.EOF {
				fmt.Println(utils.Purple("channel closed"))
				return payload.Err
			}
			fmt.Println("client.go WaitMsg payload.Err:", utils.Red(payload.Err.Error()))
			return payload.Err
		}
		ParseReplyType(payload.Data)
	}
	return nil
}
