.PHONY: build test install clean lint fmt

# Build for current platform
build:
	@echo "Building AIT..."
	@go build -o bin/ait ./cmd/ait
	@echo "✓ Built bin/ait"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@GOOS=darwin GOARCH=amd64 go build -o dist/ait-darwin-amd64 ./cmd/ait
	@GOOS=darwin GOARCH=arm64 go build -o dist/ait-darwin-arm64 ./cmd/ait
	@GOOS=linux GOARCH=amd64 go build -o dist/ait-linux-amd64 ./cmd/ait
	@GOOS=linux GOARCH=arm64 go build -o dist/ait-linux-arm64 ./cmd/ait
	@GOOS=windows GOARCH=amd64 go build -o dist/ait-windows-amd64.exe ./cmd/ait
	@echo "✓ Built all platforms to dist/"

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

# Install locally
install: build
	@echo "Installing AIT..."
	@go install ./cmd/ait
	@echo "✓ Installed ait to $(shell go env GOPATH)/bin/ait"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ dist/ coverage.out
	@echo "✓ Cleaned build artifacts"

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Formatted"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies ready"
