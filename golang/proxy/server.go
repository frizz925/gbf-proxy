package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/Frizz925/gbf-proxy/golang/consts"
	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	"github.com/Frizz925/gbf-proxy/golang/lib/logging"
)

type ServerConfig struct {
	BackendAddr string
}

type Server struct {
	base   *lib.BaseServer
	config *ServerConfig
}

type tunnel struct {
	established bool
	lock        *sync.Mutex
	logger      logging.Logger
}

func New(config *ServerConfig) lib.Server {
	return &Server{
		base:   lib.NewBaseServer("Proxy"),
		config: config,
	}
}

func (s *Server) Name() string {
	return s.base.Name
}

func (s *Server) Open(addr string) (net.Listener, error) {
	s.base.Logger.Infof("Proxy service at %s -> Controller service at %s", addr, s.config.BackendAddr)
	return s.base.Open(addr, s.serve)
}

func (s *Server) Close() error {
	return s.base.Close()
}

func (s *Server) Listener() net.Listener {
	return s.base.Listener
}

func (s *Server) WaitGroup() *sync.WaitGroup {
	return s.base.WaitGroup
}

func (s *Server) Running() bool {
	return s.base.Running()
}

func (s *Server) serve(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			_, ok := err.(*net.OpError)
			if !ok {
				panic(err)
			}
			break
		}
		go s.handle(c)
	}
}

func (s *Server) handle(conn net.Conn) {
	err := s.handleUnsafe(conn)
	if err != nil {
		reqErr, ok := err.(*httpHelpers.RequestError)
		if ok {
			s.respondAndClose(conn, reqErr.StatusCode, reqErr.Message)
		}
		s.base.Logger.Error(err)
	}
}

func (s *Server) handleUnsafe(conn net.Conn) error {
	builder := &strings.Builder{}
	buffer := make([]byte, 65535)
	for s.Running() {
		read, err := conn.Read(buffer)
		if err != nil {
			if !checkNetError(err) {
				return err
			}
			break
		}
		builder.Write(buffer[:read])
		temp := builder.String()
		if strings.Contains(temp, "\r\n\r\n") {
			break
		}
	}

	payload := builder.String()
	sepIdx := strings.Index(payload, "\r\n\r\n")
	if sepIdx <= 0 {
		s.base.Logger.Error("Payload doesn't have header/body delimiter")
		return httpHelpers.NewRequestError(400, "Bad Request", nil)
	}

	header := strings.TrimSpace(payload[:sepIdx])
	lines := strings.Split(header, "\r\n")
	if len(lines) < 2 {
		s.base.Logger.Error("Payload doesn't have headers")
		return httpHelpers.NewRequestError(400, "Bad Request", nil)
	}

	requestLine := lines[0]
	s.base.Logger.Infof("%s %s", conn.RemoteAddr(), requestLine)

	headers := make(map[string]string)
	for _, line := range lines[1:] {
		idx := strings.Index(line, ": ")
		if idx <= 0 {
			s.base.Logger.Error("Payload has deformed headers")
			return httpHelpers.NewRequestError(400, "Bad Request", nil)
		}
		name := line[:idx]
		value := line[idx+2:]
		headers[name] = value
	}

	peer, err := net.Dial("tcp", s.config.BackendAddr)
	if err != nil {
		s.base.Logger.Error("Unable to contact the backend server")
		return httpHelpers.NewRequestError(502, "Bad Gateway", nil)
	}

	method := strings.Split(requestLine, " ")[0]
	if method == "CONNECT" {
		err := s.respond(conn, 200, "Connection Established")
		if err != nil {
			return err
		}
	} else {
		err := writeString(peer, payload)
		if err != nil {
			return err
		}
	}

	t := &tunnel{
		established: true,
		lock:        &sync.Mutex{},
		logger:      s.base.Logger,
	}
	go func() {
		err := t.Pipe(peer, conn, s)
		if err != nil {
			s.base.Logger.Error(err)
		}
	}()
	return t.Pipe(conn, peer, s)
}

func (t *tunnel) Established() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.established
}

func (s *Server) respondAndClose(c net.Conn, code int, reason string) {
	defer c.Close()
	err := s.respond(c, code, reason)
	if err != nil {
		panic(err)
	}
}

func (s *Server) respond(c net.Conn, code int, reason string) error {
	s.base.Logger.Infof("%s %d %s", c.RemoteAddr(), code, reason)
	responseText := strings.Join([]string{
		fmt.Sprintf("HTTP/1.1 %d %s", code, reason),
		"Server: Granblue Proxy " + consts.Version,
		"\r\n",
	}, "\r\n")
	return writeString(c, responseText)
}

func (t *tunnel) Pipe(src net.Conn, dest net.Conn, s *Server) error {
	buffer := make([]byte, 65535)
	for s.Running() && t.Established() {
		read, err := src.Read(buffer)
		if err != nil {
			if !checkNetError(err) {
				return err
			}
			break
		}
		err = write(dest, buffer[:read])
		if err != nil {
			if !checkNetError(err) {
				return err
			}
			break
		}
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.established = false
	return nil
}

func checkNetError(err error) bool {
	_, ok := err.(*net.OpError)
	if err != io.EOF && !ok {
		return false
	}
	return true
}

func writeString(c net.Conn, responseText string) error {
	response := []byte(responseText)
	return write(c, response)
}

func write(c net.Conn, response []byte) error {
	writer := bufio.NewWriter(c)
	length := len(response)
	for written := 0; written < length; {
		n, err := writer.Write(response[written:])
		if err != nil {
			return err
		}
		written += n
	}
	writer.Flush()
	return nil
}
