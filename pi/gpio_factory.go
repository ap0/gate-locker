package main

import (
	"runtime"
)

// NewGPIO creates the appropriate GPIO implementation based on the platform
func NewGPIO(espSignalPinNum, espResponsePinNum, doorUnlockPinNum int) (GPIOInterface, error) {
	if runtime.GOOS == "linux" && runtime.GOARCH == "arm64" {
		// Running on Raspberry Pi (ARM64 Linux)
		return NewGPIOManager(espSignalPinNum, espResponsePinNum, doorUnlockPinNum)
	}

	// Running on Mac/other platforms - use mock
	return NewMockGPIOManager(espSignalPinNum, espResponsePinNum, doorUnlockPinNum)
}
