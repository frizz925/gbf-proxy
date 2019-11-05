package main

import (
	"gbf-proxy/applications"
	"gbf-proxy/lib/logger"
)

var log = logger.Factory.New()

func main() {
	app := applications.MonolithicApp{}
	err := app.Start()
	if err != nil {
		log.Fatal(err)
	}
}
