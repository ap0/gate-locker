package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const configFile = "config.json"

func main() {
	log.Println("Starting gate locker service...")

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config: %+v", config)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Println("Config file not found, creating default config")
		if err := config.save(configFile); err != nil {
			log.Printf("Warning: Failed to save default config: %v", err)
		}
	}

	gpio, err := NewGPIO(
		config.GPIOConfig.ESPSignalPin,
		config.GPIOConfig.ESPResponsePin,
		config.GPIOConfig.UnlockRelayPin,
	)
	if err != nil {
		log.Fatalf("Failed to initialize GPIO: %v", err)
	}
	defer gpio.Close()

	controller := NewGateController(config, gpio)

	apiServer := NewAPIServer(controller, config)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				controller.CheckESPSignal()
			case <-sigChan:
				return
			}
		}
	}()

	log.Println("Gate locker service started successfully")
	log.Printf("ESP32 Communication - Signal pin: %d (input), Response pin: %d (output), Unlock Relay pin: %d (output)",
		config.GPIOConfig.ESPSignalPin, config.GPIOConfig.ESPResponsePin, config.GPIOConfig.UnlockRelayPin)
	log.Printf("API server listening on %s", config.APIConfig.Port)

	<-sigChan
	log.Println("Shutting down gate locker service...")
}
