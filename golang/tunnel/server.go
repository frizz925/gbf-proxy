package tunnel

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"

	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	wsHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/websocket"
	"github.com/Frizz925/gbf-proxy/golang/lib/logging"
	"github.com/Frizz925/gbf-proxy/golang/local"
)

type OutgoingRequest = wsHelpers.Request
type IncomingResponse = wsHelpers.Response

type PendingRequest struct {
	Request   *OutgoingRequest
	Response  *IncomingResponse
	WaitGroup *sync.WaitGroup
}

type PendingRequestMap map[string]*PendingRequest

type TunnelTransport struct {
	URL             *url.URL
	Conn            *websocket.Conn
	Logger          *logging.Logger
	PendingRequests PendingRequestMap
	Mutex           *sync.Mutex
}

type Server struct {
	base      lib.Server
	client    *http.Client
	transport *TunnelTransport
}

type ServerConfig struct {
	TunnelURL *url.URL
}

func NewTunnelTransport(u *url.URL) *TunnelTransport {
	return &TunnelTransport{
		URL: u,
		Logger: logging.New(&logging.LoggerConfig{
			Name: "Tunnel",
		}),
		PendingRequests: make(PendingRequestMap),
		Mutex:           &sync.Mutex{},
	}
}

func (t *TunnelTransport) Init() error {
	conn, _, err := websocket.DefaultDialer.Dial(t.URL.String(), nil)
	if err != nil {
		return err
	}
	t.Conn = conn
	return nil
}

func (t *TunnelTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Conn == nil {
		return nil, errors.New("Tunnel is not initialized yet")
	}
	return t.SendRequest(req)
}

func (t *TunnelTransport) SendRequest(req *http.Request) (*http.Response, error) {
	r, err := httpHelpers.SerializeRequest(req)
	if err != nil {
		return nil, err
	}

	id := uuid.NewV4().String()
	p := &PendingRequest{
		Request: &OutgoingRequest{
			ID:      id,
			Payload: *r,
		},
		WaitGroup: &sync.WaitGroup{},
	}
	t.AddPendingRequest(id, p)

	data, err := msgpack.Marshal(*p.Request)
	if err != nil {
		return nil, err
	}

	p.WaitGroup.Add(1)
	err = t.Send(data)
	if err != nil {
		return nil, err
	}
	p.WaitGroup.Wait()

	if p.Response == nil {
		return nil, errors.New("Failed to get response!")
	}
	res := &p.Response.Payload
	return httpHelpers.UnserializeResponse(res)
}

func (t *TunnelTransport) AddPendingRequest(id string, p *PendingRequest) {
	defer t.Mutex.Unlock()
	t.Mutex.Lock()
	t.PendingRequests[id] = p
}

func (t *TunnelTransport) Send(data []byte) error {
	defer t.Mutex.Unlock()
	t.Mutex.Lock()
	return t.Conn.WriteMessage(websocket.BinaryMessage, data)
}

func (t *TunnelTransport) MarshalRequest(req *http.Request) ([]byte, error) {
	ser, err := httpHelpers.SerializeRequest(req)
	if err != nil {
		return nil, err
	}
	return msgpack.Marshal(*ser)
}

func (t *TunnelTransport) UnmarshalResponse(data []byte) (*http.Response, error) {
	var res *httpHelpers.Response
	err := msgpack.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return httpHelpers.UnserializeResponse(res)
}

func New(config *ServerConfig) lib.Server {
	transport := NewTunnelTransport(config.TunnelURL)
	client := &http.Client{
		Transport: transport,
	}
	base := local.New(&local.ServerConfig{
		HttpClient: client,
	})
	return &Server{
		base:      base,
		client:    client,
		transport: transport,
	}
}

func (s *Server) Name() string {
	return s.base.Name()
}

func (s *Server) Open(addr string) (net.Listener, error) {
	err := s.transport.Init()
	if err != nil {
		return nil, err
	}
	go s.listenWebSocket()
	return s.base.Open(addr)
}

func (s *Server) Close() error {
	return s.base.Close()
}

func (s *Server) WaitGroup() *sync.WaitGroup {
	return s.base.WaitGroup()
}

func (s *Server) Listener() net.Listener {
	return s.base.Listener()
}

func (s *Server) Running() bool {
	return s.base.Running()
}

func (s *Server) listenWebSocket() {
	t := s.transport
	defer t.Conn.Close()
	for s.Running() {
		err := s.serveWebSocket()
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				break
			}
			t.Logger.Error(err)
		}
	}

	if !s.Running() {
		return
	}

	t.Logger.Error("WebSocket connection lost. Restoring...")
	for {
		time.Sleep(time.Second)
		err := t.Init()
		if err != nil {
			t.Logger.Error(err)
		} else {
			break
		}
	}
	t.Logger.Info("WebSocket connection restored.")
}

func (s *Server) serveWebSocket() error {
	t := s.transport
	msgType, data, err := t.Conn.ReadMessage()
	if err != nil {
		return err
	}
	if msgType != websocket.BinaryMessage {
		t.Logger.Info(string(data))
		return nil
	}
	var r *IncomingResponse
	err = msgpack.Unmarshal(data, &r)
	if err != nil {
		return err
	}
	p := t.PendingRequests[r.ID]
	if p == nil {
		return fmt.Errorf("Pending request for '%s' not found", r.ID)
	}
	p.Response = r
	p.WaitGroup.Done()
	return nil
}
