.PHONY: all debug release clean help deps gen \
	build-all build-all-debug build-all-release \
	linux-amd64-debug linux-amd64-release \
	linux-arm64-debug linux-arm64-release \
	windows-x86_64-debug windows-x86_64-release \
	darwin-amd64-debug darwin-amd64-release \
	darwin-arm64-debug darwin-arm64-release \
	build-image

# Check if ANSI colors are supported
ifeq ($(shell tput colors 2>/dev/null),)
    # ANSI colors not supported
    BLUE :=
    GREEN :=
    RED :=
    YELLOW :=
    RESET :=
    BOLD :=
else
    # ANSI colors supported
    BLUE := \033[34m
    GREEN := \033[32m
    RED := \033[31m
    YELLOW := \033[33m
    RESET := \033[0m
    BOLD := \033[1m
endif

# Project metadata
BINARY_NAME := go-cert-provider
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build directories
BUILD_DIR := build

# Go build flags
GO_BUILD := go build
GO_FLAGS := -trimpath
DEBUG_LDFLAGS := -X 'github.com/dh-kam/go-cert-provider/config.Version=$(VERSION)' \
                 -X 'github.com/dh-kam/go-cert-provider/config.BuildTime=$(BUILD_TIME)' \
                 -X 'github.com/dh-kam/go-cert-provider/config.GitCommit=$(GIT_COMMIT)'
RELEASE_LDFLAGS := $(DEBUG_LDFLAGS) -s -w -extldflags '-static'
DEBUG_GCFLAGS := all=-N -l

# Platform configurations
PLATFORMS := linux-amd64 linux-arm64 windows-x86_64 darwin-amd64 darwin-arm64

# Detect current architecture
HOST_ARCH := $(shell uname -m)
ifeq ($(HOST_ARCH),x86_64)
    GOARCH_CURRENT := amd64
else ifeq ($(HOST_ARCH),aarch64)
    GOARCH_CURRENT := arm64
else ifeq ($(HOST_ARCH),arm64)
    GOARCH_CURRENT := arm64
else
    GOARCH_CURRENT := amd64
endif

LINUX_PLATFORM := linux-$(GOARCH_CURRENT)

# Helper function to get platform directory
platform_dir = $(BUILD_DIR)/$(1)/$(2)

# Make does not offer a recursive wildcard function, so here's one:
rwildcard=$(wildcard $1$2) $(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2))

# How to recursively find all files with the same name in a given folder
GO_SOURCES := $(call rwildcard,./,*.go)

# Default target - build for current platform
all: debug release

# Build all platforms and modes
build-all: build-all-debug build-all-release

build-all-debug: linux-amd64-debug linux-arm64-debug windows-x86_64-debug darwin-amd64-debug darwin-arm64-debug

build-all-release: linux-amd64-release linux-arm64-release windows-x86_64-release darwin-amd64-release darwin-arm64-release

# Current platform debug/release (no platform suffix)
debug: $(BUILD_DIR)/current/debug/$(BINARY_NAME)

$(BUILD_DIR)/current/debug/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building debug version for current platform...$(RESET)"
	@mkdir -p $(BUILD_DIR)/current/debug
	@CGO_ENABLED=0 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(DEBUG_LDFLAGS)" \
		-gcflags="$(DEBUG_GCFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Debug build completed: $@$(RESET)"

release: $(BUILD_DIR)/current/release/$(BINARY_NAME)

$(BUILD_DIR)/current/release/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building release version for current platform...$(RESET)"
	@mkdir -p $(BUILD_DIR)/current/release
	@CGO_ENABLED=0 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(RELEASE_LDFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Release build completed: $@$(RESET)"

# Linux AMD64
linux-amd64-debug: $(BUILD_DIR)/linux-amd64/debug/$(BINARY_NAME)

$(BUILD_DIR)/linux-amd64/debug/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building debug version for linux-amd64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux-amd64/debug
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(DEBUG_LDFLAGS)" \
		-gcflags="$(DEBUG_GCFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Debug build completed: $@$(RESET)"

linux-amd64-release: $(BUILD_DIR)/linux-amd64/release/$(BINARY_NAME)

$(BUILD_DIR)/linux-amd64/release/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building release version for linux-amd64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux-amd64/release
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(RELEASE_LDFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Release build completed: $@$(RESET)"

# Linux ARM64
linux-arm64-debug: $(BUILD_DIR)/linux-arm64/debug/$(BINARY_NAME)

$(BUILD_DIR)/linux-arm64/debug/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building debug version for linux-arm64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux-arm64/debug
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(DEBUG_LDFLAGS)" \
		-gcflags="$(DEBUG_GCFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Debug build completed: $@$(RESET)"

linux-arm64-release: $(BUILD_DIR)/linux-arm64/release/$(BINARY_NAME)

$(BUILD_DIR)/linux-arm64/release/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building release version for linux-arm64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux-arm64/release
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(RELEASE_LDFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Release build completed: $@$(RESET)"

# Windows x86_64 (amd64)
windows-x86_64-debug: $(BUILD_DIR)/windows-x86_64/debug/$(BINARY_NAME).exe

$(BUILD_DIR)/windows-x86_64/debug/$(BINARY_NAME).exe: $(GO_SOURCES)
	@echo "$(BLUE)Building debug version for windows-x86_64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows-x86_64/debug
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(DEBUG_LDFLAGS)" \
		-gcflags="$(DEBUG_GCFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Debug build completed: $@$(RESET)"

windows-x86_64-release: $(BUILD_DIR)/windows-x86_64/release/$(BINARY_NAME).exe

$(BUILD_DIR)/windows-x86_64/release/$(BINARY_NAME).exe: $(GO_SOURCES)
	@echo "$(BLUE)Building release version for windows-x86_64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows-x86_64/release
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(RELEASE_LDFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Release build completed: $@$(RESET)"

# macOS AMD64 (Intel)
darwin-amd64-debug: $(BUILD_DIR)/darwin-amd64/debug/$(BINARY_NAME)

$(BUILD_DIR)/darwin-amd64/debug/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building debug version for darwin-amd64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/darwin-amd64/debug
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(DEBUG_LDFLAGS)" \
		-gcflags="$(DEBUG_GCFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Debug build completed: $@$(RESET)"

darwin-amd64-release: $(BUILD_DIR)/darwin-amd64/release/$(BINARY_NAME)

$(BUILD_DIR)/darwin-amd64/release/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building release version for darwin-amd64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/darwin-amd64/release
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(RELEASE_LDFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Release build completed: $@$(RESET)"

# macOS ARM64 (Apple Silicon)
darwin-arm64-debug: $(BUILD_DIR)/darwin-arm64/debug/$(BINARY_NAME)

$(BUILD_DIR)/darwin-arm64/debug/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building debug version for darwin-arm64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/darwin-arm64/debug
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(DEBUG_LDFLAGS)" \
		-gcflags="$(DEBUG_GCFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Debug build completed: $@$(RESET)"

darwin-arm64-release: $(BUILD_DIR)/darwin-arm64/release/$(BINARY_NAME)

$(BUILD_DIR)/darwin-arm64/release/$(BINARY_NAME): $(GO_SOURCES)
	@echo "$(BLUE)Building release version for darwin-arm64...$(RESET)"
	@mkdir -p $(BUILD_DIR)/darwin-arm64/release
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO_BUILD) $(GO_FLAGS) \
		-ldflags="$(RELEASE_LDFLAGS)" \
		-o $@ .
	@echo "$(GREEN)Release build completed: $@$(RESET)"

# Clean build artifacts
clean:
	@echo "$(RED)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Clean completed$(RESET)"

# Update dependencies
deps:
	@echo "$(BLUE)Installing/Updating dependencies...$(RESET)"
	@go install github.com/99designs/gqlgen@latest
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(RESET)"

# Generate GraphQL code
gen:
	@echo "$(BLUE)Generating GraphQL code...$(RESET)"
	@gqlgen generate
	@echo "$(GREEN)GraphQL code generated$(RESET)"

# Docker image build
IMAGE_NAME ?= go-cert-provider
IMAGE_TAG ?= $(VERSION)

build-image: $(LINUX_PLATFORM)-release
	@echo "$(BLUE)Building Docker image for $(GOARCH_CURRENT) architecture...$(RESET)"
	@echo "$(BLUE)Building Docker image $(IMAGE_NAME):$(IMAGE_TAG)...$(RESET)"
	@docker build \
		--build-arg ARCH=$(GOARCH_CURRENT) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) .
	@docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest
	@echo "$(GREEN)Docker image built successfully:$(RESET)"
	@echo "  $(IMAGE_NAME):$(IMAGE_TAG)"
	@echo "  $(IMAGE_NAME):latest"
	@echo "  Architecture: $(GOARCH_CURRENT)"

# Show help
help:
	@echo "$(BOLD)Available targets:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Default builds (current platform):$(RESET)"
	@echo "  $(BLUE)all$(RESET)              - Build both debug and release for current platform (default)"
	@echo "  $(BLUE)debug$(RESET)            - Build debug version with debug symbols"
	@echo "  $(BLUE)release$(RESET)          - Build release version (statically linked, stripped)"
	@echo ""
	@echo "$(YELLOW)Multi-platform builds:$(RESET)"
	@echo "  $(BLUE)build-all$(RESET)        - Build all platforms in both debug and release modes"
	@echo "  $(BLUE)build-all-debug$(RESET)  - Build all platforms in debug mode"
	@echo "  $(BLUE)build-all-release$(RESET) - Build all platforms in release mode"
	@echo ""
	@echo "$(YELLOW)Platform-specific builds:$(RESET)"
	@echo "  $(BLUE)linux-amd64-debug$(RESET)    - Linux AMD64 debug build"
	@echo "  $(BLUE)linux-amd64-release$(RESET)  - Linux AMD64 release build (static)"
	@echo "  $(BLUE)linux-arm64-debug$(RESET)    - Linux ARM64 debug build"
	@echo "  $(BLUE)linux-arm64-release$(RESET)  - Linux ARM64 release build (static)"
	@echo "  $(BLUE)windows-x86_64-debug$(RESET)  - Windows x86_64 debug build"
	@echo "  $(BLUE)windows-x86_64-release$(RESET) - Windows x86_64 release build (static)"
	@echo "  $(BLUE)darwin-amd64-debug$(RESET)   - macOS Intel debug build"
	@echo "  $(BLUE)darwin-amd64-release$(RESET) - macOS Intel release build (static)"
	@echo "  $(BLUE)darwin-arm64-debug$(RESET)   - macOS Apple Silicon debug build"
	@echo "  $(BLUE)darwin-arm64-release$(RESET) - macOS Apple Silicon release build (static)"
	@echo ""
	@echo "$(YELLOW)Utility targets:$(RESET)"
	@echo "  $(BLUE)clean$(RESET)            - Remove all build artifacts"
	@echo "  $(BLUE)deps$(RESET)             - Update Go dependencies"
	@echo "  $(BLUE)gen$(RESET)              - Generate GraphQL code"
	@echo "  $(BLUE)build-image$(RESET)      - Build Docker image (alpine-based) from linux-amd64-release"
	@echo "  $(BLUE)help$(RESET)             - Show this help message"
	@echo ""
	@echo "$(YELLOW)Build configurations:$(RESET)"
	@echo "  Debug builds:   CGO disabled, debug symbols included, optimizations disabled"
	@echo "  Release builds: CGO disabled, stripped, statically linked, optimized"
	@echo ""
	@echo "$(YELLOW)Output structure:$(RESET)"
	@echo "  build/current/{debug,release}/$(BINARY_NAME)           - Current platform builds"
	@echo "  build/linux-amd64/{debug,release}/$(BINARY_NAME)       - Linux AMD64 builds"
	@echo "  build/linux-arm64/{debug,release}/$(BINARY_NAME)       - Linux ARM64 builds"
	@echo "  build/windows-x86_64/{debug,release}/$(BINARY_NAME).exe - Windows x86_64 builds"
	@echo "  build/darwin-amd64/{debug,release}/$(BINARY_NAME)      - macOS Intel builds"
	@echo "  build/darwin-arm64/{debug,release}/$(BINARY_NAME)      - macOS Apple Silicon builds"
	@echo ""
	@echo "$(YELLOW)Environment variables:$(RESET)"
	@echo "  VERSION        - Override version (default: git describe or 'dev')"
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  make                        # Build debug and release for current platform"
	@echo "  make linux-amd64-release    # Build static Linux AMD64 binary"
	@echo "  make build-all-release      # Build release binaries for all platforms"
	@echo "  VERSION=v1.0.0 make release # Build with custom version" 