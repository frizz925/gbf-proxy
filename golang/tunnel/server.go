package tunnel

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack"

	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	"github.com/Frizz925/gbf-proxy/golang/lib/logging"
	"github.com/Frizz925/gbf-proxy/golang/local"
)

type TunnelTransport struct {
	URL    *url.URL
	Conn   *websocket.Conn
	Logger *logging.Logger
}

type Server struct {
	base      lib.Server
	client    *http.Client
	transport *TunnelTransport
}

type ServerConfig struct {
	TunnelURL *url.URL
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
	data, err := t.MarshalRequest(req)
	if err != nil {
		return nil, err
	}
	err = t.Conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return nil, err
	}
	_, data, err = t.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return t.UnmarshalResponse(data)
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
	transport := &TunnelTransport{
		URL: config.TunnelURL,
		Logger: logging.New(&logging.LoggerConfig{
			Name: "Tunnel",
		}),
	}
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
