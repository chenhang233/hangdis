package connection

import (
	"fmt"
	"hangdis/utils"
	"hangdis/utils/logs"
	"io"
	"sync"
)

type FakeConn struct {
	Connection
	buf    []byte
	offset int
	waitOn chan struct{}
	closed bool
	mu     sync.Mutex
}

func NewFakeConn() *FakeConn {
	c := &FakeConn{}
	return c
}

func (c *FakeConn) Write(b []byte) (int, error) {
	if c.closed {
		return 0, io.EOF
	}
	c.mu.Lock()
	c.buf = append(c.buf, b...)
	c.mu.Unlock()
	c.notify()
	return len(b), nil
}

func (c *FakeConn) notify() {
	if c.waitOn != nil {
		c.mu.Lock()
		if c.waitOn != nil {
			logs.LOG.Debug.Println(utils.Yellow(fmt.Sprintf("notify %p", c.waitOn)))
			close(c.waitOn)
			c.waitOn = nil
		}
		c.mu.Unlock()
	}
}

func (c *FakeConn) wait(offset int) {
	c.mu.Lock()
	if c.offset != offset {
		return
	}
	if c.waitOn == nil {
		c.waitOn = make(chan struct{})
	}
	waitOn := c.waitOn
	logs.LOG.Debug.Println(utils.Yellow(fmt.Sprintf("wait on %p", waitOn)))
	c.mu.Unlock()
	<-waitOn
	logs.LOG.Debug.Println(utils.Blue(fmt.Sprintf("wait on %p finish", waitOn)))
}

func (c *FakeConn) Read(p []byte) (int, error) {
	c.mu.Lock()
	n := copy(p, c.buf[c.offset:])
	c.offset += n
	offset := c.offset
	c.mu.Unlock()
	if n == 0 {
		if c.closed {
			return n, io.EOF
		}
		c.wait(offset)
		if c.closed {
			return n, io.EOF
		}
		n = copy(p, c.buf[c.offset:])
		c.offset += n
		return n, nil
	}
	if c.closed {
		return n, io.EOF
	}
	return n, nil
}

func (c *FakeConn) Clean() {
	c.waitOn = make(chan struct{})
	c.buf = nil
	c.offset = 0
}

func (c *FakeConn) Bytes() []byte {
	return c.buf
}

func (c *FakeConn) Close() error {
	c.closed = true
	c.notify()
	return nil
}
