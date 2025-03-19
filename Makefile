.PHONY: all build clean serve typegen

# Detect operating system
UNAME := $(shell uname)

# Default Go location for wasm_exec.js
ifeq ($(UNAME), Linux)
GOROOT := $(shell go env GOROOT)
WASM_EXEC_JS := $(GOROOT)/misc/wasm/wasm_exec.js
else
GOROOT := $(shell go env GOROOT)
WASM_EXEC_JS := $(GOROOT)/misc/wasm/wasm_exec.js
endif

# If not found in standard location, try to download it
ifeq (,$(wildcard $(WASM_EXEC_JS)))
WASM_EXEC_JS := https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js
endif

# Build settings
SERVER_MAIN := ./cmd/server/main.go
WASM_MAIN := ./pkg/ui/wasm/main.go
TYPE_GEN := ./cmd/typegen/main.go
OUTPUT_DIR := ./static
WASM_OUT := $(OUTPUT_DIR)/main.wasm
SERVER_OUT := server
TYPE_GEN_OUT := typegen
TS_DEFS_OUT := $(OUTPUT_DIR)/go-wasm-types.d.ts

all: build

# Build everything including TypeScript definitions
build: $(OUTPUT_DIR)/wasm_exec.js $(WASM_OUT) $(SERVER_OUT) generate-types

# Generate TypeScript definitions (separate target to avoid circular deps)
generate-types: $(TYPE_GEN_OUT)
	@echo "Generating TypeScript definitions..."
	@mkdir -p $(OUTPUT_DIR)
	@./$(TYPE_GEN_OUT) ./pkg/ui/wasm $(TS_DEFS_OUT)
	@echo "TypeScript definitions generated at $(TS_DEFS_OUT)"

# Build the type generator
$(TYPE_GEN_OUT): $(TYPE_GEN)
	@echo "Building type generator..."
	@go build -o $(TYPE_GEN_OUT) $(TYPE_GEN)

# Copy or download wasm_exec.js
$(OUTPUT_DIR)/wasm_exec.js:
	@echo "Copying wasm_exec.js..."
	@mkdir -p $(OUTPUT_DIR)
	@if [ -f "$(WASM_EXEC_JS)" ]; then \
		cp "$(WASM_EXEC_JS)" "$(OUTPUT_DIR)/"; \
	else \
		echo "Downloading wasm_exec.js..."; \
		curl -o "$(OUTPUT_DIR)/wasm_exec.js" "$(WASM_EXEC_JS)"; \
	fi

# Build WebAssembly client
$(WASM_OUT): $(WASM_MAIN)
	@echo "Building WebAssembly client..."
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=js GOARCH=wasm go build -o $(WASM_OUT) $(WASM_MAIN)

# Build server
$(SERVER_OUT): $(SERVER_MAIN)
	@echo "Building server..."
	@go build -o $(SERVER_OUT) $(SERVER_MAIN)

# Run server
serve: build
	@echo "Starting server..."
	@./$(SERVER_OUT)

# Clean built files
clean:
	@echo "Cleaning..."
	@rm -f $(WASM_OUT) $(SERVER_OUT) $(OUTPUT_DIR)/wasm_exec.js $(TS_DEFS_OUT) $(TYPE_GEN_OUT)