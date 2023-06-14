package connection

import (
	"hangdis/utils/logs"
	"net"
	"sync"
)

type Connection struct {
	conn        net.Conn
	sendingData sync.WaitGroup
	mu          sync.Mutex
	flags       uint64
	password    string
	queue       [][][]byte
	watching    map[string]uint32
	txErrors    []error
	selectedDB  int
}

var connPool = sync.Pool{
	New: func() interface{} {
		return &Connection{}
	},
}

func (c *Connection) Close() error {
	c.password = ""
	c.queue = nil
	c.watching = nil
	c.txErrors = nil
	c.selectedDB = 0
	connPool.Put(c)
	err := c.conn.Close()
	return err
}

func NewConn(conn net.Conn) *Connection {
	c, ok := connPool.Get().(*Connection)
	if !ok {
		logs.LOG.Error.Println("connection pool make wrong type")
		return &Connection{
			conn: conn,
		}
	}
	c.conn = conn
	return c
}
