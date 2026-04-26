.PHONY: build test fmt vet tidy clean

BIN := bin/buddy
PKG := ./...

# Version metadata injected into cmd/buddy at link time. GIT_SHA falls back to
# "unknown" so a non-git tree (release tarball, sandbox) still builds; an
# unflagged `go build` separately falls back to the in-source defaults
# ("dev" / "unknown"). BUILD_DATE is RFC3339 UTC. Both are evaluated at
# `make` invocation, which is the moment we want to capture.
# Roadmap §3 M6 T3.
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.gitSHA=$(GIT_SHA) -X main.buildDate=$(BUILD_DATE)

build:
	@mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/buddy

test:
	go test -race -count=1 $(PKG)

fmt:
	gofmt -s -w .

vet:
	go vet $(PKG)

tidy:
	go mod tidy

clean:
	rm -rf bin/ *.out coverage.txt
