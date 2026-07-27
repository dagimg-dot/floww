GO ?= go
GO_BUILD_OUT ?= build/floww

.PHONY: build test vet clean fmt check-fmt lint

build:
	$(GO) build -o $(GO_BUILD_OUT) ./cmd/floww/

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

clean:
	$(GO) clean
	rm -rf build/

fmt:
	go fmt ./...

check-fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

lint:
	golangci-lint run ./...
