BINARY_NAME = burnshot
BUILD_TARGET = ./cmd/burnshot
DIST_DIR = dist
VERSION ?= dev
LDFLAGS = -s -w -X main.version=$(VERSION)

.PHONY: build cross test clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(BUILD_TARGET)

cross:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(BUILD_TARGET)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(BUILD_TARGET)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(BUILD_TARGET)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(BUILD_TARGET)

test:
	go test ./...

clean:
	rm -rf $(BINARY_NAME) $(DIST_DIR)
