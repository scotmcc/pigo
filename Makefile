MODULE   := github.com/scotmcc/pigo
BINARY   := pigo
CMD      := ./cmd/pigo

# Build metadata — injected into the binary via ldflags.
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS  := -X $(MODULE)/internal/keys.Version=$(VERSION) \
            -X $(MODULE)/internal/keys.Commit=$(COMMIT) \
            -X $(MODULE)/internal/keys.Date=$(DATE)

.PHONY: build clean vet test install

# Build the binary with version info baked in.
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

# Remove build artifacts.
clean:
	rm -f $(BINARY)

# Run static analysis.
vet:
	go vet ./...

# Run tests.
test:
	go test ./...

# Build + run pigo install.
install: build
	./$(BINARY) install
