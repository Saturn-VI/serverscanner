.DEFAULT_GOAL := help

# Binary output directory and names
BIN_DIR := bin
SCANNER_BIN := $(BIN_DIR)/scanner

# Source files
SCANNER_SOURCES := $(shell find . -maxdepth 1 -name '*.go')

.PHONY: all scanner clean deps help run-scanner

# Build all binaries
all: $(SCANNER_BIN)

# Build the scanner
scanner: $(SCANNER_BIN)

# Scanner binary target
$(SCANNER_BIN): $(SCANNER_SOURCES) | $(BIN_DIR)
	@echo "Building scanner..."
	go build -o $@ .

# Create bin directory
$(BIN_DIR):
	@echo "Creating bin directory..."
	mkdir -p $(BIN_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(BIN_DIR)

# Install/update dependencies
deps:
	@echo "Tidying dependencies..."
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Build both scanner and viewer"
	@echo "  scanner      - Build the main scanner"
	@echo "  clean        - Remove build artifacts"
	@echo "  deps         - Install/update dependencies"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make all"
	@echo "  make run-scanner ARGS=5000"
