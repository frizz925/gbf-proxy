package tunnel

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"

	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	wsHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/websocket"
	"github.com/Frizz925/gbf-proxy/golang/lib/logging"
	"github.com/Frizz925/gbf-proxy/golang/lib/websocket"
	"github.com/Frizz925/gbf-proxy/golang/local"
)

const (
	WritePeriod = time.Second * 30
	PingPeriod  = time.Second * 60
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
	Controller      *websocket.Controller
	Logger          *logging.Logger
	PendingRequests PendingRequestMap
	mutex           *sync.Mutex
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
	logger := logging.New(&logging.LoggerConfig{
		Name: "Tunnel",
	})
	return &TunnelTransport{
		Controller: websocket.NewController(&websocket.Config{
			URL: u,
			ErrorHandler: func(err error) {
				logger.Error(err)
			},
		}),
		Logger:          logger,
		PendingRequests: make(PendingRequestMap),
		mutex:           &sync.Mutex{},
	}
}

func (t *TunnelTransport) Init() error {
	return t.Controller.Connect()
}

func (t *TunnelTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.SendRequest(req)
}

func (t *TunnelTransport) SendRequest(req *http.Request) (*http.Response, error) {
	err := t.Controller.CheckLiveness()
	if err != nil {
		return nil, err
	}

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
	defer t.mutex.Unlock()
	t.mutex.Lock()
	t.PendingRequests[id] = p
}

func (t *TunnelTransport) GetPendingRequest(id string) *PendingRequest {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	return t.PendingRequests[id]
}

func (t *TunnelTransport) RemovePendingRequest(id string) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	delete(t.PendingRequests, id)
}

func (t *TunnelTransport) PopPendingRequest(id string) *PendingRequest {
	p := t.GetPendingRequest(id)
	if p != nil {
		t.RemovePendingRequest(id)
	}
	return p
}

func (t *TunnelTransport) Send(data []byte) error {
	return t.Controller.Write(data)
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

	for s.Running() {
		err := s.serveWebSocket()
		if err != nil {
			if err == websocket.NotConnectedError || err == websocket.UnhealthyError {
				break
			} else if err == websocket.NotInitializedError {
				// This should NOT happen AT ALL. If it somehow gets here
				// then there's invalid initialization logic going on here
				panic(err)
			} else if _, ok := err.(*websocket.CloseError); ok {
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
		if t.Controller.Connected() {
			err := t.Controller.Disconnect()
			if err != nil {
				// Just print the error for now
				t.Logger.Error(err)
			}
		}
		time.Sleep(time.Second)
		err := t.Controller.Connect()
		if err != nil {
			t.Logger.Error(err)
		} else {
			break
		}
	}
	t.Logger.Info("WebSocket connection restored.")
	go s.listenWebSocket()
}

func (s *Server) serveWebSocket() error {
	t := s.transport
	data, err := t.Controller.Read()
	if err != nil {
		return err
	}
	var r *IncomingResponse
	err = msgpack.Unmarshal(data, &r)
	if err != nil {
		return err
	}
	p := t.PopPendingRequest(r.ID)
	if p == nil {
		return fmt.Errorf("Pending request for '%s' not found", r.ID)
	}
	p.Response = r
	p.WaitGroup.Done()
	return nil
}
