package main

// GPIOInterface defines the contract for GPIO operations
type GPIOInterface interface {
	Close()

	// ESP32 Communication
	IsESPSignalActive() bool    // Check if ESP32 is signaling button press
	SetESPResponse(active bool) // Signal ESP32 to unlock (true) or clear signal (false)
	TriggerUnlockRelay()        // Trigger relay to simulate button press on ESP32 (for API unlock)
}
