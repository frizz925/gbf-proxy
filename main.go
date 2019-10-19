package main

import "github.com/Frizz925/gbf-proxy/applications"

func main() {
	app := applications.NewHelloWorldApp()
	app.Start()
}
