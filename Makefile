VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test lint install release-local screenshots

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/kong ./cmd/kong

test:
	go test ./...

lint:
	go vet ./...

install: build
	install -m 0755 bin/kong $(or $(PREFIX),/usr/local)/bin/kong

# Build release tarballs locally (same layout as GitHub releases).
release-local:
	@mkdir -p dist
	@for arch in amd64 arm64; do \
		out="dist/kong_$(VERSION)_linux_$${arch}"; \
		mkdir -p "$$out"; \
		CGO_ENABLED=0 GOOS=linux GOARCH=$$arch go build -trimpath -ldflags="$(LDFLAGS)" -o "$$out/kong" ./cmd/kong; \
		tar -czf "$$out.tar.gz" -C "$$out" kong; \
		sha256sum "$$out.tar.gz" > "$$out.tar.gz.sha256"; \
	done

screenshots:
	bash scripts/generate-screenshots.sh
