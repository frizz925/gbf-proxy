package websocket

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WritePeriod = time.Second * 30
	ReadPeriod  = time.Second * 30
	PingPeriod  = (ReadPeriod * 9) / 10
)

var NotInitializedError = errors.New("Not yet initialized")
var NotConnectedError = errors.New("WebSocket isn't connected")
var AlreadyConnectedError = errors.New("Already connected")
var UnhealthyError = errors.New("WebSocket connection isn't healthy")

type ErrorHandler func(error)
type CloseError = websocket.CloseError
type Upgrader = websocket.Upgrader
type Conn = websocket.Conn

type Context struct {
	Conn      *websocket.Conn
	mutex     *sync.Mutex
	rMutex    *sync.Mutex
	wMutex    *sync.Mutex
	connected bool
	healthy   bool
}

type Controller struct {
	Dialer       *websocket.Dialer
	URL          *url.URL
	Context      *Context
	errorHandler ErrorHandler
}

type Config struct {
	URL          *url.URL
	Dialer       *websocket.Dialer
	ErrorHandler ErrorHandler
}

func NewController(config *Config) *Controller {
	dialer := config.Dialer
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}

	return &Controller{
		Dialer:       dialer,
		URL:          config.URL,
		errorHandler: config.ErrorHandler,
	}
}

func (c *Controller) Connect() error {
	if c.Connected() {
		return AlreadyConnectedError
	}
	if c.Context != nil {
		defer c.Context.Unlock()
		c.Context.Lock()
		c.Context.Conn.Close()
	}

	conn, _, err := c.Dialer.Dial(c.URL.String(), nil)
	if err != nil {
		return err
	}

	c.Context = NewContext(conn)
	c.Context.Init(c.errorHandler)
	return nil
}

func (c *Controller) Disconnect() error {
	if !c.Connected() {
		return NotConnectedError
	}
	defer c.Context.Unlock()
	c.Context.Lock()

	c.Context.connected = false
	c.Context.healthy = false
	return c.Context.Conn.Close()
}

func (c *Controller) Connected() bool {
	if c.Context == nil {
		return false
	}
	return c.Context.Connected()
}

func (c *Controller) Healthy() bool {
	if c.Context == nil {
		return false
	}
	return c.Context.Healthy()
}

func (c *Controller) Read() ([]byte, error) {
	err := c.CheckLiveness()
	if err != nil {
		return nil, err
	}
	return c.Context.Read()
}

func (c *Controller) Write(data []byte) error {
	err := c.CheckLiveness()
	if err != nil {
		return err
	}

	return c.Context.Write(data)
}

func (c *Controller) CheckLiveness() error {
	if c.Context == nil {
		return NotInitializedError
	}
	return c.Context.CheckLiveness()
}

func NewContext(ws *websocket.Conn) *Context {
	return &Context{
		Conn:      ws,
		connected: true,
		healthy:   true,
		mutex:     &sync.Mutex{},
		rMutex:    &sync.Mutex{},
		wMutex:    &sync.Mutex{},
	}
}

func (c *Context) Init(errorHandler ErrorHandler) {
	c.Conn.SetPingHandler(c.PingHandler)
	c.Conn.SetPongHandler(c.PongHandler)
	c.Conn.SetCloseHandler(c.CloseHandler)
	go c.Tick(errorHandler)
}

func (c *Context) Lock() {
	c.mutex.Lock()
	c.rMutex.Lock()
	c.wMutex.Lock()
}

func (c *Context) Unlock() {
	c.mutex.Unlock()
	c.rMutex.Unlock()
	c.wMutex.Unlock()
}

func (c *Context) Connected() bool {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.connected
}

func (c *Context) Healthy() bool {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.healthy
}

func (c *Context) CheckLiveness() error {
	if !c.Connected() {
		return NotConnectedError
	}
	if !c.Healthy() {
		return UnhealthyError
	}
	return nil
}

func (c *Context) Read() ([]byte, error) {
	err := c.CheckLiveness()
	if err != nil {
		return nil, err
	}

	defer c.rMutex.Unlock()
	c.rMutex.Lock()
	_, data, err := c.Conn.ReadMessage()
	return data, err
}

func (c *Context) Write(data []byte) error {
	err := c.CheckLiveness()
	if err != nil {
		return err
	}

	defer c.wMutex.Unlock()
	c.wMutex.Lock()
	return c.Conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Context) PingHandler(string) error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	err := c.Conn.SetWriteDeadline(time.Now().Add(WritePeriod))
	if err != nil {
		return err
	}
	defer c.wMutex.Unlock()
	c.wMutex.Lock()
	return c.Conn.WriteMessage(websocket.PongMessage, nil)
}

func (c *Context) PongHandler(string) error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	return c.Conn.SetReadDeadline(time.Now().Add(ReadPeriod))
}

func (c *Context) CloseHandler(int, string) error {
	defer c.Unlock()
	c.Lock()
	c.connected = false
	c.healthy = false
	return nil
}

func (c *Context) Tick(errorHandler func(error)) {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()

	for c.Connected() {
		select {
		case <-ticker.C:
			err := c.sendPing()
			c.mutex.Lock()
			if err != nil {
				errorHandler(err)
				c.healthy = false
			} else {
				c.healthy = true
			}
			c.mutex.Unlock()
		}
	}
}

func (c *Context) sendPing() error {
	defer c.wMutex.Unlock()
	c.wMutex.Lock()
	return c.Conn.WriteMessage(websocket.PingMessage, nil)
}

func (c *Context) lockWithMutex(mutex *sync.Mutex) {
	c.mutex.Lock()
	mutex.Lock()
}

func (c *Context) unlockWithMutex(mutex *sync.Mutex) {
	c.mutex.Unlock()
	mutex.Unlock()
}
