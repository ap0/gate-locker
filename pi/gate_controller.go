package main

import (
	"log"
	"time"
)

type GateController struct {
	config   *Config
	gpio     GPIOInterface
	doorbell *DoorbellClient
}

func NewGateController(config *Config, gpio GPIOInterface) *GateController {
	gc := &GateController{
		config:   config,
		gpio:     gpio,
		doorbell: NewDoorbellClient(&config.DoorbellConfig),
	}

	return gc
}

// UnlockGate triggers an unlock via the relay (simulates button press on ESP32)
func (gc *GateController) UnlockGate() {
	log.Println("API unlock requested - triggering unlock relay")
	gc.gpio.TriggerUnlockRelay()
}

func (gc *GateController) CheckESPSignal() {
	if gc.gpio.IsESPSignalActive() {
		log.Println("ESP32 signal detected (button pressed on ESP32)")

		// Trigger doorbell if enabled
		var delay time.Duration
		if gc.config.DoorbellConfig.Enabled {
			log.Println("Calling doorbell API")
			var err error
			delay, err = gc.doorbell.TriggerDoorbell()
			if err != nil {
				log.Printf("Warning: Failed to trigger doorbell: %v", err)
				delay = 0
			}
		}

		// Wait for the calculated delay before signaling ESP32
		if delay > 0 {
			log.Printf("Waiting %v before signaling ESP32 to unlock", delay)
			time.Sleep(delay)
		}

		// Signal ESP32 to proceed with unlock
		// ESP32 will handle its own time-based restrictions and LED feedback
		log.Println("Signaling ESP32 to unlock gate")
		gc.gpio.SetESPResponse(true)
		time.Sleep(100 * time.Millisecond) // Brief signal
		gc.gpio.SetESPResponse(false)
		log.Println("ESP32 unlock signal complete")
	}
}
