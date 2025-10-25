package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	GPIOConfig `json:"gpio"`

	APIConfig `json:"api"`

	DoorbellConfig `json:"doorbell"`
}

type GPIOConfig struct {
	// ESP32 Communication Pins
	ESPSignalPin    int `json:"esp_signal_pin"`    // Input: receives button signal from ESP32
	ESPResponsePin  int `json:"esp_response_pin"`  // Output: sends unlock response to ESP32
	UnlockRelayPin  int `json:"unlock_relay_pin"`  // Output: triggers relay to simulate button press (for API unlock)
}

type APIConfig struct {
	Port string `json:"port"`
}

type DoorbellConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
}

func loadConfig(filename string) (*Config, error) {
	config := &Config{
		GPIOConfig: GPIOConfig{
			ESPSignalPin:   26, // Receives signal from ESP32's DEVICE_SIGNAL_PIN
			ESPResponsePin: 15, // Sends response to ESP32's DEVICE_RESPONSE_PIN
			UnlockRelayPin: 25, // Triggers relay to simulate button press (wired parallel to physical button)
		},
		APIConfig: APIConfig{
			Port: ":8080",
		},
		DoorbellConfig: DoorbellConfig{
			Enabled: false,
			Host:    "lightning-control.local:8080",
		},
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return config, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}
