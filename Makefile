BINARY  := portscout
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet fmt run clean

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	go test -race -cover ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

run: build
	./bin/$(BINARY) $(ARGS)

clean:
	rm -rf bin
