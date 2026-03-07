BINARY     := linked
BUILD_DIR  := dist
GO         := go
MODULE     := github.com/russ-blaisdell/linked

.PHONY: all build test lint install clean skill

all: build

## build: compile the binary for the current platform
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) ./cmd/linked

## install: install linked into /usr/local/bin
install: build
	install -m 0755 $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "✓ linked installed to /usr/local/bin/linked"

## test: run all integration tests
test:
	$(GO) test ./tests/integration/... -v -count=1

## test-short: run tests without verbose output
test-short:
	$(GO) test ./tests/integration/... -count=1

## lint: run go vet
lint:
	$(GO) vet ./...

## skill: install the OpenClaw skill
skill:
	@SKILLS_DIR=~/.openclaw/workspace/skills/linkedin; \
	mkdir -p $$SKILLS_DIR; \
	cp skill/linkedin/skill.md $$SKILLS_DIR/skill.md; \
	echo "✓ LinkedIn skill installed to $$SKILLS_DIR"

## release: build binaries for darwin/arm64, darwin/amd64, linux/amd64
release:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin  GOARCH=arm64 $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64  ./cmd/linked
	GOOS=darwin  GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64  ./cmd/linked
	GOOS=linux   GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-linux-amd64   ./cmd/linked
	@echo "✓ Binaries in $(BUILD_DIR)/"

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## help: show this help
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
