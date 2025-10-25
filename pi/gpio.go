package main

import (
	"log"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type GPIOManager struct {
	espSignalPin   rpio.Pin
	espResponsePin rpio.Pin
	unlockRelayPin rpio.Pin

	espSignalPinNum   int
	espResponsePinNum int
	unlockRelayPinNum int

	lastSignalTime time.Time
	debounceTime   time.Duration
}

func NewGPIOManager(espSignalPinNum, espResponsePinNum, unlockRelayPinNum int) (*GPIOManager, error) {
	err := rpio.Open()
	if err != nil {
		return nil, err
	}

	gm := &GPIOManager{
		espSignalPin:      rpio.Pin(espSignalPinNum),
		espResponsePin:    rpio.Pin(espResponsePinNum),
		unlockRelayPin:    rpio.Pin(unlockRelayPinNum),
		espSignalPinNum:   espSignalPinNum,
		espResponsePinNum: espResponsePinNum,
		unlockRelayPinNum: unlockRelayPinNum,
		debounceTime:      250 * time.Millisecond,
	}

	// ESP32 signal pin - input from ESP32 (active HIGH when button pressed)
	gm.espSignalPin.Input()
	gm.espSignalPin.PullDown()

	// ESP32 response pin - output to ESP32 (set HIGH to signal unlock)
	gm.espResponsePin.Output()
	gm.espResponsePin.Low() // Start with no signal

	// Unlock relay pin - output to relay (active HIGH to trigger relay)
	gm.unlockRelayPin.Output()
	gm.unlockRelayPin.Low() // Start LOW (relay off)

	return gm, nil
}

func (gm *GPIOManager) Close() {
	gm.espResponsePin.Low()  // Clear response signal
	gm.unlockRelayPin.Low()  // Turn off relay (LOW = off)
	rpio.Close()
}

// IsESPSignalActive checks if ESP32 is signaling a button press
func (gm *GPIOManager) IsESPSignalActive() bool {
	now := time.Now()
	if now.Sub(gm.lastSignalTime) < gm.debounceTime {
		return false
	}

	if gm.espSignalPin.Read() == rpio.High {
		gm.lastSignalTime = now
		return true
	}
	return false
}

// SetESPResponse sets the response pin to signal ESP32
// true = signal to unlock, false = clear signal
func (gm *GPIOManager) SetESPResponse(active bool) {
	if active {
		log.Println("Signaling ESP32 to unlock gate")
		gm.espResponsePin.High()
	} else {
		log.Println("Clearing ESP32 unlock signal")
		gm.espResponsePin.Low()
	}
}

// TriggerUnlockRelay triggers the relay to simulate a button press on ESP32
func (gm *GPIOManager) TriggerUnlockRelay() {
	log.Printf("Triggering unlock relay on GPIO %d HIGH (activating relay, active HIGH)", gm.unlockRelayPinNum)
	gm.unlockRelayPin.High() // Pull HIGH to activate relay (active HIGH)
	log.Printf("Unlock relay GPIO %d is now HIGH, holding for 500ms", gm.unlockRelayPinNum)
	time.Sleep(500 * time.Millisecond) // Hold long enough to trigger button press
	gm.unlockRelayPin.Low() // Turn off relay (LOW = off)
	log.Printf("Unlock relay GPIO %d released to LOW", gm.unlockRelayPinNum)
}
