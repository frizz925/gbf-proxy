package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Frizz925/gbf-proxy/golang/controller"
	"github.com/Frizz925/gbf-proxy/golang/lib"
	"github.com/Frizz925/gbf-proxy/golang/proxy"
)

func main() {
	if len(os.Args) < 3 {
		fatal("Usage: gbf-proxy <service> <address> [args ...]")
	}
	name := os.Args[1]
	addr := os.Args[2]
	if !strings.Contains(addr, ":") {
		port, err := strconv.Atoi(addr)
		if err != nil {
			fatal("Address must be either a port or host:port format")
		}
		addr = fmt.Sprintf("0.0.0.0:%d", port)
	}
	switch name {
	case "controller":
		controllerService(addr)
	case "proxy":
		proxyService(addr)
	default:
		fatalf("Unknown service '%s'", name)
	}
}

func controllerService(addr string) {
	runService("Controller", controller.NewServer(), addr)
}

func proxyService(addr string) {
	if len(os.Args) < 4 {
		fatal("Usage: gbf-proxy proxy <address> <backend-address> [args ...]")
	}
	backendAddr := os.Args[3]
	log.Printf("Using %s as backend", backendAddr)
	runService("Proxy", proxy.NewServer(backendAddr), addr)
}

func runService(name string, s lib.Server, addr string) {
	l, err := s.Open(addr)
	if err != nil {
		panic(err)
	}
	log.Printf("%s service listening at %s\n", name, l.Addr().String())
	s.WaitGroup().Wait()
}

func fatalf(format string, args ...interface{}) {
	fatal(fmt.Sprintf(format, args...))
}

func fatal(message string) {
	_, err := os.Stderr.WriteString(message)
	if err != nil {
		panic(err)
	}
	os.Exit(1)
}
