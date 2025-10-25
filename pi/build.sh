#!/bin/bash

echo "Building gate-locker for Raspberry Pi Zero 2 W..."

# Set environment variables for cross-compilation to ARM64 (Pi Zero 2 W)
export GOOS=linux
export GOARCH=arm64

# Build the binary
# CGO_ENABLED=1 go build -o gate-locker-arm64 .
docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.24 env CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o gate-locker-arm64

if [ $? -eq 0 ]; then
    echo "Build successful! Binary: gate-locker-arm64"
    echo "Copy this binary to your Raspberry Pi Zero 2 W"
else
    echo "Build failed!"
    exit 1
fi