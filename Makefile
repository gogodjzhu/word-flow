VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")
LDFLAGS = -ldflags "-X github.com/gogodjzhu/word-flow/pkg/cmd/root.version=$(VERSION)"
# Output directory
OUTPUT_DIR = build
# Target platforms
PLATFORMS = darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64
# Go source file path
SRC = cmd/wordflow/main.go
# Application name
APP_NAME = wordflow

# Default target: build for all platforms
all: $(PLATFORMS)

# Build and compress for each platform
$(PLATFORMS):
	@$(eval OS := $(word 1,$(subst /, ,$@)))
	@$(eval ARCH := $(word 2,$(subst /, ,$@)))
	@$(eval OUTPUT_NAME := $(APP_NAME)-$(OS)-$(ARCH))
	@if [ "$(OS)" = "windows" ]; then OUTPUT_NAME="$(OUTPUT_NAME).exe"; fi
	@echo "Building for $(OS)/$(ARCH)..."
	GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(OUTPUT_NAME) $(SRC)
	@if [ $$? -ne 0 ]; then echo "Failed to build for $(OS)/$(ARCH)"; exit 1; fi
	@echo "Compressing $(OUTPUT_NAME) into tar.gz/zip..."
	@if [ "$(OS)" = "windows" ]; then \
		cd $(OUTPUT_DIR) && zip -j $(OUTPUT_NAME).zip $(OUTPUT_NAME); \
	else \
		tar -czvf $(OUTPUT_DIR)/$(OUTPUT_NAME).tar.gz -C $(OUTPUT_DIR) $(OUTPUT_NAME); \
	fi

# Build for current platform
build:
	go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) $(SRC)

# Generate SHA256 checksums
checksums:
	@cd $(OUTPUT_DIR) && sha256sum *.tar.gz *.zip > SHA256SUMS

# Clean build files
clean:
	rm -rf $(OUTPUT_DIR)

.PHONY: all build checksums clean