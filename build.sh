#!/bin/bash
# This script builds the Go WebAssembly module and copies wasm_exec.js.

# Check if GOROOT is set
if [ -z "$GOROOT" ]; then
  echo "Error: GOROOT environment variable is not set."
  echo "Please ensure Go is installed correctly and GOROOT is defined."
  exit 1
fi

# Check if wasm_exec.js exists in GOROOT
WASM_EXEC_PATH="$GOROOT/misc/wasm/wasm_exec.js"
if [ ! -f "$WASM_EXEC_PATH" ]; then
  echo "Error: wasm_exec.js not found at $WASM_EXEC_PATH"
  echo "Please ensure your Go installation is complete."
  exit 1
fi

echo "Copying wasm_exec.js from $WASM_EXEC_PATH to the current directory..."
cp "$WASM_EXEC_PATH" .

echo "Building WebAssembly module..."
GOOS=js GOARCH=wasm go build -o main.wasm go/main.go

echo "Build complete. main.wasm and wasm_exec.js should be in the current directory."
