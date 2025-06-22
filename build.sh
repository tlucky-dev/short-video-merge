#!/bin/bash
# This script builds the Go WebAssembly module.
GOOS=js GOARCH=wasm go build -o main.wasm go/main.go
