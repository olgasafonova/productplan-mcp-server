.PHONY: build build-all clean test

VERSION := 4.2.0
BINARY := productplan
BUILD_DIR := build

# Build for current platform
build:
	go build -ldflags="-s -w" -o $(BINARY) .

# Build for all platforms
build-all: clean
	mkdir -p $(BUILD_DIR)
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 .
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe .

clean:
	rm -rf $(BUILD_DIR) $(BINARY)

# Install locally
install: build
	cp $(BINARY) /usr/local/bin/

# Create release archives
release: build-all
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY)-$(VERSION)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64 && \
	tar -czf $(BINARY)-$(VERSION)-darwin-arm64.tar.gz $(BINARY)-darwin-arm64 && \
	tar -czf $(BINARY)-$(VERSION)-linux-amd64.tar.gz $(BINARY)-linux-amd64 && \
	tar -czf $(BINARY)-$(VERSION)-linux-arm64.tar.gz $(BINARY)-linux-arm64 && \
	zip $(BINARY)-$(VERSION)-windows-amd64.zip $(BINARY)-windows-amd64.exe
