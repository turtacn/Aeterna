BINARY_NAME=aeterna
VERSION=0.1.0

.PHONY: all build test clean docker

all: build

build:
	@echo "Building Aeterna UPHR-O Engine..."
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/$(BINARY_NAME) cmd/aeterna/main.go

test:
	go test -v ./...

clean:
	go clean
	rm -rf bin/

docker:
	docker build -t aeterna:latest .

# Personal.AI order the ending