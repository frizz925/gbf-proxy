package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Frizz925/gbf-proxy/golang/proxy"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: gbf-proxy PORT [HOST]")
	}
	host := "0.0.0.0"
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Port must be a number!")
	}
	if len(os.Args) >= 3 {
		host = os.Args[1]
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	s := proxy.NewServer()
	l, err := s.Open(addr)
	if err != nil {
		panic(err)
	}
	log.Println("Proxy server listening at " + l.Addr().String())
	s.WaitGroup.Wait()
}
