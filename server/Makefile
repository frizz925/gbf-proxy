# Makefile
VERSION=$(shell git describe --always --long --dirty)
BUILDTIME=$(shell date +%s)

GOENTRY=cli/monolithic/main.go
GOBUILD=go build -ldflags="-X main.version=$(VERSION) -X main.buildTime=$(BUILDTIME)"

default: build

build:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/gbf-proxy $(GOENTRY)
