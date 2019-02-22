GOCMD=go
GODEP=dep
GOBUILD=$(GOCMD) build -v -ldflags="-X main.version=$(VERSION)"
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run

BUILD_DIR=bin
BINARY_NAME=gbf-proxy
VERSION=$(shell git describe --always --long --dirty)

all: deps test build
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)
build-darwin:
	GOOS=darwin GOARCH=386 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-386
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
build-linux:
	GOOS=linux GOARCH=386 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
build-alpine-linux:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-alpine-linux-386
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-alpine-linux-amd64
build-windows:
	GOOS=windows GOARCH=386 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-386
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64
test:
	$(GOTEST) -race ./...
coverage:
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...
clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-*
run:
	$(GORUN) ./main.go
deps:
	$(GODEP) ensure
docker-image: deps test build-alpine-linux
	docker build -t gbf-proxy:$(VERSION) .
version:
	$(info $(VERSION))
