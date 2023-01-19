GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

all: test build

build:
	go build -v

test: 
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
