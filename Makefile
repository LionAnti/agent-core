.PHONY: all build test vet race bench lint clean

GO ?= go
GOFLAGS ?=

all: build vet test

build:
	$(GO) build $(GOFLAGS) ./...

test:
	$(GO) test $(GOFLAGS) -count=1 ./...

vet:
	$(GO) vet $(GOFLAGS) ./...

race:
	$(GO) test -race -count=1 ./...

bench:
	$(GO) test -bench=. -benchmem -count=3 ./...

lint:
	@golangci-lint run ./... || echo "golangci-lint not installed"

coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean:
	$(GO) clean ./...
	rm -f coverage.out coverage.html

.PHONY: example-basic example-advanced

example-basic:
	$(GO) run ./examples/basic/

example-advanced:
	$(GO) run ./examples/advanced/
