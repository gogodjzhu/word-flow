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
	# Special handling for Windows
	@if [ "$(OS)" = "windows" ]; then OUTPUT_NAME="$(OUTPUT_NAME).exe"; fi
	@echo "Building for $(OS)/$(ARCH)..."
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTPUT_DIR)/$(OUTPUT_NAME) $(SRC)
	@if [ $$? -ne 0 ]; then echo "Failed to build for $(OS)/$(ARCH)"; exit 1; fi
	@echo "Compressing $(OUTPUT_NAME) into tar.gz..."
	@tar -czvf $(OUTPUT_DIR)/$(OUTPUT_NAME).tar.gz -C $(OUTPUT_DIR) $(OUTPUT_NAME)

# Clean build files
clean:
	rm -rf $(OUTPUT_DIR)

.PHONY: all clean
