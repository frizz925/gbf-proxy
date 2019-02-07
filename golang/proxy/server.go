package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/Frizz925/gbf-proxy/golang/lib"
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
}

func New(config *ServerConfig) lib.Server {
	return &Server{
		base:   lib.NewBaseServer("Proxy"),
		config: config,
	}
}

func (s *Server) Open(addr string) (net.Listener, error) {
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
		go s.handleSafe(c)
	}
}

func (s *Server) handleSafe(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	s.handle(conn)
}

func (s *Server) handle(conn net.Conn) {
	builder := &strings.Builder{}
	reader := bufio.NewReader(conn)
	for s.Running() {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}
		builder.WriteString(line)
		temp := builder.String()
		if strings.Contains(temp, "\r\n\r\n") {
			break
		}
	}

	payload := builder.String()
	sepIdx := strings.Index(payload, "\r\n\r\n")
	if sepIdx <= 0 {
		respondAndClose(conn, 400, "Bad Request")
		return
	}

	header := strings.TrimSpace(payload[:sepIdx])
	lines := strings.Split(header, "\r\n")
	if len(lines) < 2 {
		respondAndClose(conn, 400, "Bad Request")
		return
	}

	requestLine := lines[0]
	log.Printf("[Proxy] %s %s", conn.RemoteAddr(), requestLine)

	headers := make(map[string]string)
	for _, line := range lines[1:] {
		idx := strings.Index(line, ": ")
		if idx <= 0 {
			respondAndClose(conn, 400, "Bad Request")
			return
		}
		name := line[:idx]
		value := line[idx+2:]
		headers[name] = value
	}

	peer, err := net.Dial("tcp", s.config.BackendAddr)
	if err != nil {
		respondAndClose(conn, 502, "Bad Gateway")
		return
	}

	method := strings.Split(requestLine, " ")[0]
	if method == "CONNECT" {
		err := respond(conn, 200, "Connection Established")
		if err != nil {
			panic(err)
		}
	} else {
		err := writeString(peer, payload)
		if err != nil {
			panic(err)
		}
	}

	t := &tunnel{
		established: true,
		lock:        &sync.Mutex{},
	}
	go t.Pipe(peer, conn)
	go t.Pipe(conn, peer)
}

func (t *tunnel) Pipe(src net.Conn, dest net.Conn) {
	defer func() {
		src.Close()
		dest.Close()
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	reader := bufio.NewReader(src)
	buffer := make([]byte, 65535)
	for t.Established() {
		n, err := reader.Read(buffer)
		if err != nil {
			_, ok := err.(*net.OpError)
			if err != io.EOF && !ok {
				panic(err)
			}
			break
		}
		err = write(dest, buffer[:n])
		if err != nil {
			_, ok := err.(*net.OpError)
			if err != io.EOF && !ok {
				panic(err)
			}
			break
		}
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.established = false
}

func (t *tunnel) Established() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.established
}

func respondAndClose(c net.Conn, code int, reason string) {
	defer c.Close()
	err := respond(c, code, reason)
	if err != nil {
		panic(err)
	}
}

func respond(c net.Conn, code int, reason string) error {
	log.Printf("%s %d %s", c.RemoteAddr(), code, reason)
	responseText := strings.Join([]string{
		fmt.Sprintf("HTTP/1.1 %d %s", code, reason),
		"Server: Granblue Proxy 0.1-alpha",
		"\r\n",
	}, "\r\n")
	return writeString(c, responseText)
}

func writeString(c net.Conn, responseText string) error {
	response := []byte(responseText)
	return write(c, response)
}

func write(c net.Conn, response []byte) error {
	writer := bufio.NewWriter(c)
	_, err := writer.Write(response)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}
