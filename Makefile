.PHONY: all build test install clean rust-lib

# Detect OS for library extension
UNAME := $(shell uname -s)
ifeq ($(UNAME),Darwin)
	LIB_EXT := dylib
	LIB_INSTALL_DIR := /usr/local/lib
else ifeq ($(UNAME),Linux)
	LIB_EXT := so
	LIB_INSTALL_DIR := /usr/local/lib
else
	LIB_EXT := dll
	LIB_INSTALL_DIR := /usr/local/lib
endif

RUST_LIB := beefdown-rs/target/release/libbeefdown_clock.$(LIB_EXT)
RUST_LIB_NAME := libbeefdown_clock.$(LIB_EXT)

# Default target: build everything
all: build

# Build both Rust library and Go binary
build: rust-lib
	@echo "Building Go binary..."
	go build

# Build the Rust library
rust-lib:
	@echo "Building Rust library..."
	cd beefdown-rs && cargo build --release

# Run tests for both Go and Rust
test: rust-lib
	@echo "Running Go tests..."
	go test ./sequence
	go test ./sequence/parsers/metadata
	go test ./sequence/parsers/part
	@echo "Running Rust tests..."
	cd beefdown-rs && cargo test --release

# Install the Go binary (library stays in project directory)
install: rust-lib
	@echo "Installing Go binary..."
	go install .
	@echo ""
	@echo "✅ Installation complete!"
	@echo "   - Go binary: $(shell go env GOPATH)/bin/beefdown"
	@echo "   - Rust library: $(RUST_LIB) (local)"
	@echo ""
	@echo "Note: The Rust library is not installed system-wide."
	@echo "      Keep the project directory to maintain functionality."

# Clean build artifacts
clean:
	@echo "Cleaning Go build cache..."
	go clean
	@echo "Cleaning Rust build artifacts..."
	cd beefdown-rs && cargo clean
	@echo "✅ Clean complete!"

# Uninstall the Go binary (library is local, so just clean it)
uninstall:
	@echo "Uninstalling Go binary..."
	go clean -i
	@echo "✅ Uninstall complete!"
	@echo ""
	@echo "To remove the Rust library, run: make clean"

# Development: build and run with example file
dev: build
	./beefdown sequences/example.md

# Show help
help:
	@echo "Beefdown Build System"
	@echo ""
	@echo "Targets:"
	@echo "  make              - Build everything (Rust library + Go binary)"
	@echo "  make build        - Same as default"
	@echo "  make rust-lib     - Build only the Rust library"
	@echo "  make test         - Run all tests (Go + Rust)"
	@echo "  make install      - Install Go binary to GOPATH (library stays local)"
	@echo "  make uninstall    - Remove installed binary"
	@echo "  make clean        - Clean all build artifacts"
	@echo "  make dev          - Build and run with example file"
	@echo "  make help         - Show this help"
	@echo ""
	@echo "System info:"
	@echo "  OS: $(UNAME)"
	@echo "  Library extension: .$(LIB_EXT)"
	@echo "  Rust library: $(RUST_LIB)"
	@echo "  Go binary path: $(shell go env GOPATH)/bin/"
