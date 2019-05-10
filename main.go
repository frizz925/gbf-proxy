package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var appName = "Granblue Proxy"
var appVersion = "latest"

type ShutdownRequest struct{}

type ResponseError struct {
	Code    int
	Status  string
	Message string
}

func (e *ResponseError) Error() string {
	return e.Message
}

func main() {
	err := mainUnsafe()
	if err != nil {
		log.Fatal(err)
	}
}

func mainUnsafe() error {
	wg := &sync.WaitGroup{}
	l, err := net.Listen("tcp4", "127.0.0.1:8088")
	if err != nil {
		return err
	}
	log.Printf("Listening at %s", l.Addr().String())

	c := make(chan ShutdownRequest, 1)
	// Handle the listener itself
	wg.Add(1)
	go handleListener(c, l, wg)

	// Handle the exit signals
	return handleSignal(c, l, wg)
}

func handleSignal(c chan ShutdownRequest, l net.Listener, wg *sync.WaitGroup) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	select {
	case <-signalChan:
		c <- ShutdownRequest{}
		err := l.Close()
		wg.Wait()
		if err != nil {
			return err
		}
	}
	return nil
}

func handleListener(c chan ShutdownRequest, l net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-c:
			return
		default:
		}

		conn, err := l.Accept()
		if err != nil {
			continue
		}

		wg.Add(1)
		go handleConn(conn, wg)
	}
}

func handleConn(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()
	defer log.Printf("%s Connection closed", conn.RemoteAddr().String())
	log.Printf("%s Connection opened", conn.RemoteAddr().String())
	err := handleConnRequest(conn)
	if err != nil {
		handleConnError(conn, err)
	}
}

func handleConnRequest(conn net.Conn) error {
	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return err
	}
	defer req.Body.Close()

	err = handleRequest(conn, req)
	if err != nil {
		if v, ok := err.(*ResponseError); ok {
			errMsg := fmt.Sprintf("%d %s: %s", v.Code, v.Status, v.Message)
			handleConnError(conn, errors.New(errMsg))
			res := createResponse(req, v.Code, v.Status, stringToBody(v.Message))
			defer res.Body.Close()
			return dumpResponse(conn, res)
		} else {
			handleConnError(conn, err)
			body := stringToBody(err.Error())
			res := createResponse(req, http.StatusInternalServerError, "Internal Server Error", body)
			defer res.Body.Close()
			return dumpResponse(conn, res)
		}
	}
	return nil
}

func handleRequest(in net.Conn, req *http.Request) error {
	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		host = req.URL.Hostname()
		port = req.URL.Port()
	}

	if host == "" {
		return &ResponseError{
			Code:    http.StatusBadRequest,
			Status:  "Bad Request",
			Message: "Malformed request: Host is empty",
		}
	}
	if !strings.HasPrefix(host, "game") || !strings.HasSuffix(host, ".granbluefantasy.jp") {
		return &ResponseError{
			Code:    http.StatusForbidden,
			Status:  "Forbidden",
			Message: fmt.Sprintf("Host %s is not allowed", host),
		}
	}

	if req.Method == "CONNECT" {
		log.Printf("%s %s %s", in.RemoteAddr().String(), req.Method, req.Host)
		err := dumpResponse(in, createResponse(req, http.StatusOK, "Connection Established", nil))
		if err != nil {
			return err
		}
		newReq, err := http.ReadRequest(bufio.NewReader(in))
		if err != nil {
			return err
		}
		defer newReq.Body.Close()
		req = newReq
	}

	if port == "" {
		port = "80"
	}
	if req.Host == "" {
		req.Host = net.JoinHostPort(host, port)
	}
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}
	if req.URL.Path == "" {
		req.URL.Path = "/"
	}
	log.Printf("%s %s %s", in.RemoteAddr().String(), req.Method, req.URL.String())

	addr := net.JoinHostPort(host, port)
	out, err := net.Dial("tcp4", addr)
	if err != nil {
		return err
	}
	defer out.Close()
	go pipe(in, out)

	p, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
	}
	_, err = out.Write(p)
	if err != nil {
		return err
	}
	pipe(out, in)

	return nil
}

func pipe(dst io.WriteCloser, src io.ReadCloser) {
	defer src.Close()
	defer dst.Close()
	for {
		_, err := io.CopyN(dst, src, 4096)
		if err != nil {
			break
		}
	}
}

func createResponse(req *http.Request, code int, status string, body io.ReadCloser) *http.Response {
	header := make(http.Header)
	header.Set("X-Proxy-Server", fmt.Sprintf("%s %s", appName, appVersion))

	return &http.Response{
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		StatusCode: code,
		Status:     status,
		Body:       body,
		Header:     header,
	}
}

func dumpResponse(w io.Writer, res *http.Response) error {
	p, err := httputil.DumpResponse(res, res.Body != nil)
	if err != nil {
		return err
	}
	_, err = w.Write(p)
	return err
}

func stringToBody(payload string) io.ReadCloser {
	return bytesToBody([]byte(payload))
}

func bytesToBody(payload []byte) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(payload))
}

func handleConnError(conn net.Conn, err error) {
	log.Printf("%s %s", conn.RemoteAddr().String(), err)
}

func handleError(err error) {
	log.Println(err)
}
