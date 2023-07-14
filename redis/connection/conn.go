package connection

import (
	"hangdis/utils/logs"
	"net"
	"sync"
)

const (
	flagSlave = uint64(1 << iota)
	flagMaster
	// within a transaction
	flagMulti
)

type Connection struct {
	conn        net.Conn
	sendingData sync.WaitGroup
	mu          sync.Mutex
	subs        map[string]bool
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

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() error {
	c.password = ""
	c.queue = nil
	c.watching = nil
	c.txErrors = nil
	c.subs = nil
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

func (c *Connection) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	c.sendingData.Add(1)
	defer func() {
		c.sendingData.Done()
	}()
	return c.conn.Write(b)
}

func (c *Connection) Name() string {
	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return ""
}
func (c *Connection) SetPassword(password string) {
	c.password = password
}
func (c *Connection) GetPassword() string {
	return c.password
}

func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}

func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

func (c *Connection) AddTxError(err error) {
	c.txErrors = append(c.txErrors, err)
}

func (c *Connection) GetTxErrors() []error {
	return c.txErrors
}

func (c *Connection) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.subs == nil {
		c.subs = make(map[string]bool)
	}
	c.subs[channel] = true
}
func (c *Connection) UnSubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subs, channel)
}

func (c *Connection) SubsCount() int {
	return len(c.subs)
}

func (c *Connection) GetChannels() []string {
	n := len(c.subs)
	if n == 0 {
		return []string{}
	}
	channels := make([]string, n)
	i := 0
	for channel, _ := range c.subs {
		channels[i] = channel
		i++
	}
	return channels
}

func (c *Connection) SetSlave() {
	c.flags |= flagSlave
}
func (c *Connection) IsSlave() bool {
	return c.flags&flagSlave > 0
}
func (c *Connection) SetMaster() {
	c.flags |= flagMaster
}
func (c *Connection) IsMaster() bool {
	return c.flags&flagMaster > 0
}

func (c *Connection) InMultiState() bool {
	return c.flags&flagMulti > 0

}

func (c *Connection) SetMultiState(state bool) {
	if !state {
		c.watching = nil
		c.queue = nil
		c.flags &= ^flagMulti
		return
	}
	c.flags |= flagMulti
}
func (c *Connection) GetQueuedCmdLine() [][][]byte {
	return c.queue
}
func (c *Connection) EnqueueCmd(cmdLine [][]byte) {
	c.queue = append(c.queue, cmdLine)
}
func (c *Connection) ClearQueuedCmds() {
	c.queue = nil
}
func (c *Connection) GetWatching() map[string]uint32 {
	if c.watching == nil {
		c.watching = make(map[string]uint32)
	}
	return c.watching
}
