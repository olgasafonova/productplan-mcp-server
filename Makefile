.PHONY: build build-all clean test lint vet check-api

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY := productplan-mcp-server
BUILD_DIR := build
LDFLAGS := -s -w -X main.version=$(VERSION)

# Build for current platform
build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/productplan

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run linter
lint:
	golangci-lint run --timeout=5m

# Run go vet
vet:
	go vet ./...

# Build for all platforms
build-all: clean
	mkdir -p $(BUILD_DIR)
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/productplan
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/productplan
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/productplan
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/productplan
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/productplan

clean:
	rm -rf $(BUILD_DIR) $(BINARY)

# Install locally
install: build
	cp $(BINARY) /usr/local/bin/

# Create release archives
# Run integration tests against live ProductPlan API (requires PRODUCTPLAN_API_TOKEN)
check-api:
	go test -tags integration -run TestAPIEndpoints -v ./internal/api/

release: build-all
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY)-$(VERSION)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64 && \
	tar -czf $(BINARY)-$(VERSION)-darwin-arm64.tar.gz $(BINARY)-darwin-arm64 && \
	tar -czf $(BINARY)-$(VERSION)-linux-amd64.tar.gz $(BINARY)-linux-amd64 && \
	tar -czf $(BINARY)-$(VERSION)-linux-arm64.tar.gz $(BINARY)-linux-arm64 && \
	zip $(BINARY)-$(VERSION)-windows-amd64.zip $(BINARY)-windows-amd64.exe
