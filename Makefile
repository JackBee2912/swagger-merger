.PHONY: build clean test demo example help build-cli build-docker run-docker ci-cd install release

# Module name
MODULE_NAME=swagger-merger
CLI_NAME=swagger-merger
VERSION=1.0.0

# Build the library
build:
	@echo "🔨 Building $(MODULE_NAME) library..."
	go mod tidy
	go build ./pkg/merger
	@echo "✅ Build completed!"

# Build CLI tool
build-cli:
	@echo "🔨 Building $(CLI_NAME) CLI tool..."
	go mod tidy
	go build -o bin/$(CLI_NAME) ./cmd/$(CLI_NAME)
	@echo "✅ CLI tool built: bin/$(CLI_NAME)"

# Build CLI tool for multiple platforms
build-cli-all:
	@echo "🔨 Building $(CLI_NAME) CLI tool for multiple platforms..."
	go mod tidy
	mkdir -p bin
	
	# Linux
	GOOS=linux GOARCH=amd64 go build -o bin/$(CLI_NAME)-linux-amd64 ./cmd/$(CLI_NAME)
	GOOS=linux GOARCH=arm64 go build -o bin/$(CLI_NAME)-linux-arm64 ./cmd/$(CLI_NAME)
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build -o bin/$(CLI_NAME)-darwin-amd64 ./cmd/$(CLI_NAME)
	GOOS=darwin GOARCH=arm64 go build -o bin/$(CLI_NAME)-darwin-arm64 ./cmd/$(CLI_NAME)
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -o bin/$(CLI_NAME)-windows-amd64.exe ./cmd/$(CLI_NAME)
	GOOS=windows GOARCH=arm64 go build -o bin/$(CLI_NAME)-windows-arm64.exe ./cmd/$(CLI_NAME)
	
	@echo "✅ CLI tools built for all platforms in bin/"

# Build Docker image
build-docker:
	@echo "🐳 Building Docker image..."
	docker build -t $(MODULE_NAME):$(VERSION) .
	docker tag $(MODULE_NAME):$(VERSION) $(MODULE_NAME):latest
	@echo "✅ Docker image built: $(MODULE_NAME):$(VERSION)"

# Run Docker container
run-docker:
	@echo "🐳 Running Docker container..."
	docker run --rm -v $(PWD):/workspace $(MODULE_NAME):latest --help

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	go clean
	rm -rf bin/
	rm -f $(CLI_NAME)
	@echo "✅ Clean completed!"

# Install CLI tool locally
install: build-cli
	@echo "📦 Installing CLI tool..."
	cp bin/$(CLI_NAME) /usr/local/bin/$(CLI_NAME)
	@echo "✅ CLI tool installed to /usr/local/bin/$(CLI_NAME)"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	@echo "✅ Dependencies installed!"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./pkg/merger -v
	@echo "✅ Tests completed!"

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test ./pkg/merger -v -cover
	@echo "✅ Tests with coverage completed!"

# Run all tests with coverage report
test-coverage-report:
	@echo "🧪 Running tests with coverage report..."
	go test ./pkg/merger -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run demo
demo: build
	@echo "🚀 Running demo..."
	cd demo && go run main.go

# Run example
example: build
	@echo "📚 Running example..."
	cd example && go run main.go

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./pkg/merger/... ./cmd/$(CLI_NAME)/...
	@echo "✅ Code formatted!"

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run ./pkg/merger/... ./cmd/$(CLI_NAME)/...
	@echo "✅ Linting completed!"

# Generate documentation
docs:
	@echo "📖 Generating documentation..."
	godoc -http=:6060 &
	@echo "✅ Documentation server started at http://localhost:6060"

# CI/CD pipeline
ci-cd: deps fmt lint test test-coverage build-cli
	@echo "🚀 CI/CD pipeline completed successfully!"

# Release preparation
release: clean build-cli-all test-coverage-report
	@echo "📦 Preparing release v$(VERSION)..."
	@echo "✅ Release artifacts ready in bin/ and coverage.html"

# Quick test of CLI tool
test-cli: build-cli
	@echo "🧪 Testing CLI tool..."
	./bin/$(CLI_NAME) --version
	./bin/$(CLI_NAME) --help
	@echo "✅ CLI tool test completed!"

# Show help
help:
	@echo "Swagger Merger Library - Makefile"
	@echo "================================="
	@echo ""
	@echo "Available commands:"
	@echo "  build              - Build the library"
	@echo "  build-cli          - Build CLI tool"
	@echo "  build-cli-all      - Build CLI tool for all platforms"
	@echo "  build-docker       - Build Docker image"
	@echo "  run-docker         - Run Docker container"
	@echo "  clean              - Clean build artifacts"
	@echo "  install            - Install CLI tool locally"
	@echo "  deps               - Install dependencies"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  test-coverage-report - Run tests with coverage report"
	@echo "  test-cli           - Test CLI tool"
	@echo "  demo               - Run demo application"
	@echo "  example            - Run example application"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code"
	@echo "  docs               - Generate documentation"
	@echo "  ci-cd              - Run full CI/CD pipeline"
	@echo "  release            - Prepare release artifacts"
	@echo "  help               - Show this help"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build-cli"
	@echo "  make test"
	@echo "  make ci-cd"
	@echo "  make release"
	@echo ""
	@echo "CLI usage:"
	@echo "  ./bin/swagger-merger --input file1.yaml,file2.yaml --output merged.yaml"
	@echo "  ./bin/swagger-merger --input ./docs --output merged.yaml --verbose --stats"
	@echo "  ./bin/swagger-merger --input ./docs --output merged.yaml --pattern '*.yaml,*.yml'" 