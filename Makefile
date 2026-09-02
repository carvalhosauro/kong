# kong — developer tasks.
# Tools install into ./bin so nothing pollutes go.mod or the global environment.

BIN                   := $(CURDIR)/bin
GOLANGCI_LINT_VERSION := v2.12.2
LEFTHOOK_VERSION      := v2.1.10
GORELEASER_VERSION    := v2.17.0
GOLANGCI_LINT         := $(BIN)/golangci-lint
LEFTHOOK              := $(BIN)/lefthook
GORELEASER            := $(BIN)/goreleaser
export PATH           := $(BIN):$(PATH)

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all setup tools hooks tidy fmt fmt-check lint vet test build install release-local release-check snapshot screenshots clean help

## all: format, lint, test (default local loop)
all: fmt lint test

## setup: one-shot dev environment bootstrap (tools + git hooks + tidy)
setup: tools hooks tidy
	@echo ">> dev environment ready — run 'make all'"

## tools: install pinned dev tools into ./bin
tools: $(GOLANGCI_LINT) $(LEFTHOOK)

$(GOLANGCI_LINT):
	@mkdir -p $(BIN)
	@echo ">> installing golangci-lint $(GOLANGCI_LINT_VERSION) -> $(BIN)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
		| sh -s -- -b $(BIN) $(GOLANGCI_LINT_VERSION)

$(LEFTHOOK):
	@mkdir -p $(BIN)
	@echo ">> installing lefthook $(LEFTHOOK_VERSION) -> $(BIN)"
	@GOBIN=$(BIN) go install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)

## hooks: install git hooks via lefthook
hooks: $(LEFTHOOK)
	@git config --unset core.hooksPath 2>/dev/null || true
	@$(LEFTHOOK) install
	@echo ">> git hooks installed (lefthook)"

## tidy: sync go.mod/go.sum
tidy:
	go mod tidy

## fmt: format code (gofumpt + goimports via golangci-lint)
fmt: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) fmt

## fmt-check: fail if code is not formatted (CI gate)
fmt-check: $(GOLANGCI_LINT)
	@diff="$$($(GOLANGCI_LINT) fmt --diff 2>/dev/null)"; \
	if [ -n "$$diff" ]; then echo "$$diff"; echo ">> not formatted — run 'make fmt'"; exit 1; fi

## lint: run all configured linters
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

## vet: go vet only (subset of lint; kept for quick checks)
vet:
	go vet ./...

## test: race-enabled test run
test:
	go test -race ./...

## build: build the CLI into ./bin (version stamped from git)
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN)/kong ./cmd/kong

## install: install kong into PREFIX (default /usr/local/bin)
install: build
	install -m 0755 $(BIN)/kong $(or $(PREFIX),/usr/local)/bin/kong

$(GORELEASER):
	@mkdir -p $(BIN)
	@echo ">> installing goreleaser $(GORELEASER_VERSION) -> $(BIN)"
	@GOBIN=$(BIN) go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

## release-check: validate .goreleaser.yaml
release-check: $(GORELEASER)
	$(GORELEASER) check

## snapshot: dry-run a release build (no publish, no tag needed)
snapshot: $(GORELEASER)
	$(GORELEASER) build --snapshot --clean

## release-local: build release tarballs locally (same layout as GitHub releases)
release-local:
	@mkdir -p dist
	@for arch in amd64 arm64; do \
		out="dist/kong_$(VERSION)_linux_$${arch}"; \
		mkdir -p "$$out"; \
		CGO_ENABLED=0 GOOS=linux GOARCH=$$arch go build -trimpath -ldflags="$(LDFLAGS)" -o "$$out/kong" ./cmd/kong; \
		tar -czf "$$out.tar.gz" -C "$$out" kong; \
	done
	@cd dist && sha256sum kong_*.tar.gz > checksums.txt

## screenshots: regenerate README terminal screenshots
screenshots:
	bash scripts/generate-screenshots.sh

## clean: remove build + coverage artifacts (tools kept)
clean:
	rm -rf dist coverage.out $(BIN)/kong

## help: list targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
