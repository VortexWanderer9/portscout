BINARY  := portscout
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet check coverage fmt run clean

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	go test -race -cover ./...

vet:
	go vet ./...

check: test vet

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt:
	gofmt -s -w .

run: build
	./bin/$(BINARY) $(ARGS)

clean:
	rm -rf bin
