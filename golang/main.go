package main

import (
	"log"

	"github.com/Frizz925/gbf-proxy/golang/proxy"
)

func main() {
	s := proxy.NewServer()
	l, err := s.Start("localhost:8088")
	if err != nil {
		panic(err)
	}
	log.Println("Proxy server listening at " + l.Addr().String())
	s.WaitGroup.Wait()
}
