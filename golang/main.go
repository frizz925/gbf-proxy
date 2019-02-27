package main

import (
	"github.com/Frizz925/gbf-proxy/golang/cmd"
	"github.com/Frizz925/gbf-proxy/golang/consts"
)

var version = "latest"

func main() {
	consts.Version = version
	cmd.Execute()
}
