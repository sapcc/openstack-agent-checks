GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

all: test build

build:
	go build -ldflags="-X github.com/prometheus/common/version.Version=v$(date -u +%Y%m%d%H%M%S) -X github.com/prometheus/common/version.Branch=$(git branch --show-current) -X github.com/prometheus/common/version.Revision=$(git rev-parse HEAD)" -v

test: 
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
