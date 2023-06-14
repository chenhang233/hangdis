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

// -----------------------------------------------------------------------------------------------------------

func (c *Connection) Subscribe(channel string) {

}
func (c *Connection) UnSubscribe(channel string) {

}
func (c *Connection) SubsCount() int {
	return 0
}
func (c *Connection) GetChannels() []string {
	return nil
}
func (c *Connection) InMultiState() bool {
	return false
}
func (c *Connection) SetMultiState(bool) {
}
func (c *Connection) GetQueuedCmdLine() [][][]byte {
	return nil
}
func (c *Connection) EnqueueCmd([][]byte) {
}
func (c *Connection) ClearQueuedCmds() {

}
func (c *Connection) GetWatching() map[string]uint32 {
	return nil
}

func (c *Connection) SetSlave() {

}
func (c *Connection) IsSlave() bool {
	return false
}
func (c *Connection) SetMaster() {

}
func (c *Connection) IsMaster() bool {
	return false
}
