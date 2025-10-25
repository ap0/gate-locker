package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type MockGPIOManager struct {
	espSignalPinNum   int
	espResponsePinNum int
	unlockRelayPinNum int

	espSignalState   bool
	espResponseState bool
	unlockRelayState bool
	lastSignalTime   time.Time
	debounceTime     time.Duration

	mutex       sync.RWMutex
	stdinReader *bufio.Scanner
	signalChan  chan bool
	shutdown    chan bool
}

func NewMockGPIOManager(espSignalPinNum, espResponsePinNum, unlockRelayPinNum int) (*MockGPIOManager, error) {
	log.Println("=== MOCK GPIO MODE ===")
	log.Printf("Simulating GPIO pins - ESP Signal: %d, ESP Response: %d, Unlock Relay: %d", espSignalPinNum, espResponsePinNum, unlockRelayPinNum)
	log.Println("Type 'signal' or 's' to simulate ESP32 button signal")
	log.Println("Type 'status' to see current state")
	log.Println("Type 'quit' or 'q' to exit")
	log.Println("=====================")

	mgm := &MockGPIOManager{
		espSignalPinNum:   espSignalPinNum,
		espResponsePinNum: espResponsePinNum,
		unlockRelayPinNum: unlockRelayPinNum,
		debounceTime:      250 * time.Millisecond,
		stdinReader:       bufio.NewScanner(os.Stdin),
		signalChan:        make(chan bool, 10),
		shutdown:          make(chan bool),
	}

	// Start stdin reader goroutine
	go mgm.readStdin()

	return mgm, nil
}

func (mgm *MockGPIOManager) readStdin() {
	for {
		select {
		case <-mgm.shutdown:
			return
		default:
			if mgm.stdinReader.Scan() {
				input := strings.TrimSpace(strings.ToLower(mgm.stdinReader.Text()))
				switch input {
				case "signal", "s":
					mgm.signalChan <- true
					log.Println("[MOCK] ESP32 signal simulated (button press from ESP32)")
				case "quit", "q", "exit":
					log.Println("[MOCK] Quit command received")
					os.Exit(0)
				case "status":
					mgm.printStatus()
				default:
					if input != "" {
						log.Printf("[MOCK] Unknown command: %s (try 'signal', 'status', or 'quit')", input)
					}
				}
			}
		}
	}
}

func (mgm *MockGPIOManager) printStatus() {
	mgm.mutex.RLock()
	defer mgm.mutex.RUnlock()

	signalStatus := "LOW"
	if mgm.espSignalState {
		signalStatus = "HIGH"
	}

	responseStatus := "LOW"
	if mgm.espResponseState {
		responseStatus = "HIGH"
	}

	unlockRelayStatus := "LOW"
	if mgm.unlockRelayState {
		unlockRelayStatus = "HIGH"
	}

	log.Printf("[MOCK STATUS] ESP Signal: %s, ESP Response: %s, Unlock Relay: %s", signalStatus, responseStatus, unlockRelayStatus)
}

func (mgm *MockGPIOManager) Close() {
	log.Println("[MOCK] Closing GPIO connections")
	close(mgm.shutdown)
	mgm.SetESPResponse(false)
}

func (mgm *MockGPIOManager) IsESPSignalActive() bool {
	now := time.Now()
	if now.Sub(mgm.lastSignalTime) < mgm.debounceTime {
		return false
	}

	select {
	case <-mgm.signalChan:
		mgm.lastSignalTime = now
		mgm.mutex.Lock()
		mgm.espSignalState = true
		mgm.mutex.Unlock()
		return true
	default:
		return false
	}
}

func (mgm *MockGPIOManager) SetESPResponse(active bool) {
	mgm.mutex.Lock()
	defer mgm.mutex.Unlock()

	if mgm.espResponseState != active {
		mgm.espResponseState = active
		if active {
			log.Printf("[MOCK] ESP Response Pin %d: HIGH - Signaling ESP32 to unlock", mgm.espResponsePinNum)
		} else {
			log.Printf("[MOCK] ESP Response Pin %d: LOW - Clearing signal", mgm.espResponsePinNum)
		}
	}
}

func (mgm *MockGPIOManager) TriggerUnlockRelay() {
	log.Printf("[MOCK] Unlock Relay Pin %d: HIGH (activating relay, active HIGH)", mgm.unlockRelayPinNum)
	mgm.mutex.Lock()
	mgm.unlockRelayState = true
	mgm.mutex.Unlock()

	time.Sleep(500 * time.Millisecond)

	mgm.mutex.Lock()
	mgm.unlockRelayState = false
	mgm.mutex.Unlock()
	log.Printf("[MOCK] Unlock Relay Pin %d: LOW (relay off)", mgm.unlockRelayPinNum)
}
