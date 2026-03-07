.PHONY: all build test install clean rust-lib uninstall help

# Default target: build everything
all: build

# Build both Rust library and Go binary
build: rust-lib
	@echo "Building Go binary..."
	go build

# Build the Rust library
rust-lib:
	@echo "Building Rust library..."
	cd rust && cargo build --release

# Run tests for both Go and Rust
test: rust-lib
	@echo "Running Go tests..."
	go test ./...
	@echo "Running Rust tests..."
	cd rust && cargo test --release

# Install the Go binary (statically linked, no external library needed)
install: rust-lib
	@echo "Installing Go binary..."
	go install .
	@echo ""
	@echo "✅ Installation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning Go build cache..."
	go clean
	@echo "Cleaning Rust build artifacts..."
	cd rust && cargo clean
	@echo "✅ Clean complete!"

# Uninstall the Go binary
uninstall:
	@echo "Uninstalling Go binary..."
	go clean -i
	@echo "✅ Uninstall complete!"

# Show help
help:
	@echo "Beefdown Build System"
	@echo ""
	@echo "Targets:"
	@echo "  make              - Build everything (Rust static library + Go binary)"
	@echo "  make build        - Same as default"
	@echo "  make rust-lib     - Build only the Rust static library"
	@echo "  make test         - Run all tests (Go + Rust)"
	@echo "  make install      - Install self-contained Go binary to GOPATH"
	@echo "  make uninstall    - Remove installed binary"
	@echo "  make clean        - Clean all build artifacts"
	@echo "  make help         - Show this help"
