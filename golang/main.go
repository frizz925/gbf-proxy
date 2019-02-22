package main

import "github.com/Frizz925/gbf-proxy/golang/cmd"

var version = "latest"

func main() {
	cmd.Version = version
	cmd.Execute()
}
