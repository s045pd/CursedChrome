.PHONY: all build test test-race vet lint fmt clean run smoke tidy

BIN_DIR ?= ./bin
BIN     ?= $(BIN_DIR)/cursed-server

all: vet test build

build:
	@mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags="-s -w" -o $(BIN) ./cmd/cursed-server

run: build
	$(BIN)

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed; skipping"; \
	fi

tidy:
	go mod tidy

smoke:
	@bash ./scripts/smoke.sh

clean:
	rm -rf $(BIN_DIR)
