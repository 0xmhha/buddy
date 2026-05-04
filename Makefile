.PHONY: build test fmt vet tidy clean release-binaries install-plugin uninstall-plugin print-%

BIN     := bin/buddy
BIN_MCP := bin/buddy-mcp
PKG     := ./...

# Version metadata injected into cmd/buddy at link time. GIT_SHA falls back to
# "unknown" so a non-git tree (release tarball, sandbox) still builds; an
# unflagged `go build` separately falls back to the in-source defaults
# ("dev" / "unknown"). BUILD_DATE is RFC3339 UTC. Both are evaluated at
# `make` invocation, which is the moment we want to capture.
# Roadmap §3 M6 T3.
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.gitSHA=$(GIT_SHA) -X main.buildDate=$(BUILD_DATE)

# RELEASE_VERSION is embedded into release artifact filenames. Keep this in
# sync with cmd/buddy/main.go's `var version` (line ~38) — when bumping the
# version for a release, update both. v0.2 may move to a single-source-of-truth
# VERSION file or build-time embed if release cadence increases.
# Roadmap §3 M6 T1.
RELEASE_VERSION ?= 0.1.0
DIST := dist
RELEASE_BINS := \
	$(DIST)/buddy_$(RELEASE_VERSION)_linux_amd64 \
	$(DIST)/buddy_$(RELEASE_VERSION)_linux_arm64 \
	$(DIST)/buddy_$(RELEASE_VERSION)_darwin_amd64 \
	$(DIST)/buddy_$(RELEASE_VERSION)_darwin_arm64 \
	$(DIST)/buddy-mcp_$(RELEASE_VERSION)_linux_amd64 \
	$(DIST)/buddy-mcp_$(RELEASE_VERSION)_linux_arm64 \
	$(DIST)/buddy-mcp_$(RELEASE_VERSION)_darwin_amd64 \
	$(DIST)/buddy-mcp_$(RELEASE_VERSION)_darwin_arm64

build:
	@mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN)     ./cmd/buddy
	go build -trimpath                        -o $(BIN_MCP) ./cmd/buddy-mcp

test:
	go test -race -count=1 $(PKG)

fmt:
	gofmt -s -w .

vet:
	go vet $(PKG)

tidy:
	go mod tidy

clean:
	rm -rf bin/ $(DIST) *.out coverage.txt

GITHUB_REPO := 0xmhha/buddy

install-plugin:
	claude plugin marketplace add $(GITHUB_REPO)
	claude plugin install buddy@buddy

install-plugin-local:
	claude plugin marketplace add $(shell pwd)
	claude plugin install buddy@buddy

uninstall-plugin:
	claude plugin uninstall buddy@buddy || true
	claude plugin marketplace remove buddy || true

# print-VAR: utility for CI to read a Makefile variable without parsing.
# Example: `make -s print-RELEASE_VERSION` -> `0.1.0`.
# Used by .github/workflows/release.yml to verify the pushed tag matches
# RELEASE_VERSION before publishing artifacts. Roadmap §3 M6 T2.
print-%:
	@echo $($*)

# release-binaries cross-compiles buddy for the v0.1 support matrix:
# linux/amd64, linux/arm64, darwin/amd64, darwin/arm64. CGO_ENABLED=0 is safe
# because modernc.org/sqlite is pure Go — no cgo toolchain needed, and the
# resulting binaries are statically linked on Linux. Each binary carries the
# same -trimpath + ldflags as `make build` so `--version` reports the right
# gitSHA and buildDate. SHA256SUMS uses `shasum -a 256` output convention so
# downstream verification works with both `shasum -a 256 -c` (BSD) and
# `sha256sum -c` (GNU coreutils).
# Roadmap §3 M6 T1.
release-binaries: $(DIST)/SHA256SUMS

$(DIST)/SHA256SUMS: $(RELEASE_BINS)
	@cd $(DIST) && shasum -a 256 buddy_$(RELEASE_VERSION)_* buddy-mcp_$(RELEASE_VERSION)_* > SHA256SUMS
	@cat $(DIST)/SHA256SUMS

$(DIST)/buddy_$(RELEASE_VERSION)_linux_amd64:
	@mkdir -p $(DIST)
	GOOS=linux  GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $@ ./cmd/buddy

$(DIST)/buddy_$(RELEASE_VERSION)_linux_arm64:
	@mkdir -p $(DIST)
	GOOS=linux  GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $@ ./cmd/buddy

$(DIST)/buddy_$(RELEASE_VERSION)_darwin_amd64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $@ ./cmd/buddy

$(DIST)/buddy_$(RELEASE_VERSION)_darwin_arm64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $@ ./cmd/buddy

$(DIST)/buddy-mcp_$(RELEASE_VERSION)_linux_amd64:
	@mkdir -p $(DIST)
	GOOS=linux  GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o $@ ./cmd/buddy-mcp

$(DIST)/buddy-mcp_$(RELEASE_VERSION)_linux_arm64:
	@mkdir -p $(DIST)
	GOOS=linux  GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -o $@ ./cmd/buddy-mcp

$(DIST)/buddy-mcp_$(RELEASE_VERSION)_darwin_amd64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o $@ ./cmd/buddy-mcp

$(DIST)/buddy-mcp_$(RELEASE_VERSION)_darwin_arm64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -o $@ ./cmd/buddy-mcp
