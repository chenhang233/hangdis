package client

import (
	"fmt"
	"hangdis/interface/redis"
	"hangdis/redis/parser"
	"hangdis/redis/protocol"
	"net"
	"sync"
)

type Client struct {
	conn        net.Conn
	addr        string
	pendingReqs chan *request
	waitingReqs chan *request
	status      int
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
		status:      STOP,
		addr:        addr,
		conn:        dial,
		pendingReqs: make(chan *request, chanSize),
		waitingReqs: make(chan *request, chanSize),
	}, nil
}

func (c *Client) Start() {
	go c.handleRead()
	go c.handleWrite()
	c.status = RUN
}

func (c *Client) Close() {
	c.status = STOP
	close(c.pendingReqs)
	close(c.waitingReqs)
	_ = c.conn.Close()
}
func (c *Client) Send(args [][]byte) redis.Reply {
	if c.status == STOP {
		return protocol.MakeErrReply("client closed")
	}
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
	_, err := c.conn.Write(bytes)
	if err != nil {
		fmt.Println(err)
		req.wait.Done()
	} else {
		c.waitingReqs <- req
	}
}

func (c *Client) handleRead() {
	ch := parser.ParseStream(c.conn)
	for payload := range ch {
		if payload.Err != nil {
			fmt.Println("handleRead payload.Err:", payload.Err)
			//c.Close()
			return
		}
		c.handlePayload(payload)
	}
}

func (c *Client) handlePayload(p *parser.Payload) {
	w := <-c.waitingReqs
	w.reply = p.Data
	w.wait.Done()
}
