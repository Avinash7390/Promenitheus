.PHONY: build test clean run-target run-scraper help

# Build variables
BINARY_DIR=bin
PROMENITHEUS_BINARY=$(BINARY_DIR)/promenitheus
EXAMPLE_TARGET_BINARY=$(BINARY_DIR)/example-target

# Default target
all: build

# Build both binaries
build: $(PROMENITHEUS_BINARY) $(EXAMPLE_TARGET_BINARY)

$(PROMENITHEUS_BINARY):
	@echo "Building Promenitheus..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(PROMENITHEUS_BINARY) ./cmd/promenitheus

$(EXAMPLE_TARGET_BINARY):
	@echo "Building example target..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(EXAMPLE_TARGET_BINARY) ./cmd/example-target

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)

# Run example target on port 8080
run-target: $(EXAMPLE_TARGET_BINARY)
	@echo "Starting example target service on :8080..."
	@$(EXAMPLE_TARGET_BINARY) --port 8080

# Run Promenitheus on port 9090
run-scraper: $(PROMENITHEUS_BINARY)
	@echo "Starting Promenitheus scraper on :9090..."
	@$(PROMENITHEUS_BINARY) --config config.yaml --port 9090

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  make build        - Build both binaries"
	@echo "  make test         - Run all tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make run-target   - Run example target service"
	@echo "  make run-scraper  - Run Promenitheus scraper"
	@echo "  make deps         - Install dependencies"
	@echo "  make help         - Show this help message"
