.PHONY: build test lint ci clean install-tools fmt vet

BINARY_NAME := gaze
MAIN_PATH := ./cmd/gaze
BUILD_DIR := ./bin
COVERAGE_FILE := coverage.out

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

test:
	@echo "Running tests..."
	go test -race -coverprofile=$(COVERAGE_FILE) ./...

lint:
	@echo "Running linters..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	gofmt -w .
	goimports -w .

vet:
	@echo "Running go vet..."
	go vet ./...

ci: lint test build
	@echo "All CI checks passed."

clean:
	rm -rf $(BUILD_DIR) $(COVERAGE_FILE)

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
