.PHONY: build test test-routing test-skill-form verify-go-version verify-versions fmt vet tidy clean release-binaries install-plugin uninstall-plugin print-%

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
RELEASE_VERSION ?= 0.6.4
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

# test-routing verifies the buddy plugin router wire-up invariants
# (SKILL.md/PROCEDURE.md counts, plugin.json command coverage, single-mode
# target existence, description drift). Run before merging changes that
# touch plugin/commands/buddy/, plugin/skills/, or plugin/.claude-plugin/.
test-routing:
	@bash scripts/test-router-wireup.sh

# test-skill-form reports PROCEDURE.md section-form deviations across 148 skills
# (B6 — ADR-004 §2.2 condition 3). Report-only by default; --strict for CI gate
# (currently 43 deviations are *known free-form / externally-inherited* skills,
# so strict mode is opt-in until B6 follow-up unifies them).
test-skill-form:
	@bash scripts/lint-skill-procedure.sh

# verify-go-version guards against the drift that broke the first v0.6.1
# tag: go.mod's `go` directive was bumped to 1.25.0 in commit becbcf1
# (W3-3 scheduler) but .github/workflows/release.yml stayed at
# `go-version: '1.22'` for four releases. setup-go@v5 hid this with
# GOTOOLCHAIN=auto auto-download fallback; @v6 exports GOTOOLCHAIN=local
# by default and removed the fallback, so the drift fails CI fast.
#
# This target reads go.mod's go directive and release.yml's go-version
# input, compares major.minor (a `1.25` workflow value matches a
# `1.25.0` go.mod directive), and exits non-zero with a clear message
# on mismatch. Wired into the release workflow before the cross-compile
# step.
verify-go-version:
	@mod_go=$$(awk '/^go [0-9]/ {print $$2; exit}' go.mod); \
	wf_go=$$(grep -E "^[[:space:]]*go-version:" .github/workflows/release.yml | head -1 | sed -E "s/.*go-version:[[:space:]]*['\"]?([^'\"[:space:]]+).*/\1/"); \
	mod_short=$$(echo "$$mod_go" | awk -F. '{print $$1"."$$2}'); \
	wf_short=$$(echo "$$wf_go" | awk -F. '{print $$1"."$$2}'); \
	if [ -z "$$mod_go" ] || [ -z "$$wf_go" ]; then \
	  echo "verify-go-version: could not parse go.mod ($$mod_go) or release.yml ($$wf_go)"; \
	  exit 2; \
	fi; \
	if [ "$$mod_short" != "$$wf_short" ]; then \
	  echo "::error::go.mod go directive ($$mod_go) and release.yml go-version ($$wf_go) disagree on major.minor"; \
	  echo "::error::setup-go@v6 exports GOTOOLCHAIN=local — workflow cannot auto-download a newer toolchain"; \
	  echo "fix: align release.yml go-version with go.mod's $$mod_short or update go.mod via 'go mod edit -go=$$wf_short'"; \
	  exit 1; \
	fi; \
	echo "verify-go-version: go.mod ($$mod_go) and release.yml ($$wf_go) agree on major.minor=$$mod_short"

# verify-versions guards the second drift axis identified in v0.6.1's
# post-mortem: the five version sources that must stay synchronized for
# every release. The existing release.yml step covers tag↔Makefile;
# verify-go-version covers go.mod↔workflow; this target covers the
# Makefile↔plugin.json↔marketplace.json↔server.go↔main.go axis.
#
# All five values must equal RELEASE_VERSION exactly. Any mismatch is
# almost certainly a half-finished release bump that would publish
# inconsistent binaries (e.g. plugin.json says 0.6.1 but the cli binary
# self-reports 0.6.0). Run as part of the release workflow before
# cross-compile.
verify-versions:
	@want="$(RELEASE_VERSION)"; \
	mk="$$want"; \
	pj=$$(awk -F'"' '/^  "version":/ {print $$4; exit}' plugin/.claude-plugin/plugin.json); \
	mp=$$(awk -F'"' '/"version":[[:space:]]*"/ {print $$4; exit}' .claude-plugin/marketplace.json); \
	sv=$$(awk -F'"' '/Version:[[:space:]]*"/ {print $$2; exit}' internal/mcp/server.go); \
	mn=$$(awk -F'"' '/version[[:space:]]*=[[:space:]]*"/ {print $$2; exit}' cmd/buddy/main.go); \
	miss=""; \
	for pair in "Makefile:$$mk" "plugin.json:$$pj" "marketplace.json:$$mp" "server.go:$$sv" "main.go:$$mn"; do \
	  name=$${pair%%:*}; val=$${pair#*:}; \
	  if [ "$$val" != "$$want" ]; then miss="$$miss $$name=$$val"; fi; \
	done; \
	if [ -n "$$miss" ]; then \
	  echo "::error::version sources disagree with Makefile RELEASE_VERSION ($$want):$$miss"; \
	  echo "fix: bump every disagreeing source to $$want, or update Makefile RELEASE_VERSION to match the intended release"; \
	  exit 1; \
	fi; \
	echo "verify-versions: all 5 sources agree on $$want"

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
