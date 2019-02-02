GOCMD=go
GODEP=dep
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run

BUILD_DIR=bin
BINARY_NAME=gbf-proxy

all: deps test build
build: build-darwin build-linux build-windows
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64 -v
test:
	$(GOTEST) ./...
clean:
	$(GOCLEAN)
	rm -f builds/$(BINARY_NAME)-*
run:
	$(GORUN) ./main.go
deps:
	$(GODEP) ensure
docker: deps test build-linux
	docker build . -t frizz925/gbf-proxy
