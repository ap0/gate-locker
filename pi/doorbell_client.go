package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DoorbellClient struct {
	config *DoorbellConfig
	client *http.Client
}

type DoorbellResponse struct {
	Status        string `json:"status"`
	VoiceLengthMs int    `json:"voice_length_ms"`
}

func NewDoorbellClient(config *DoorbellConfig) *DoorbellClient {
	return &DoorbellClient{
		config: config,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (d *DoorbellClient) TriggerDoorbell() (time.Duration, error) {
	if !d.config.Enabled {
		return 0, nil
	}

	url := fmt.Sprintf("http://%s/doorbell", d.config.Host)
	resp, err := d.client.Post(url, "application/json", bytes.NewReader([]byte{}))
	if err != nil {
		return 0, fmt.Errorf("failed to trigger doorbell: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("doorbell API returned status: %d", resp.StatusCode)
	}

	var data DoorbellResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to decode doorbell response: %w", err)
	}

	// Calculate delay as 1/4 of voice length
	delayMs := data.VoiceLengthMs / 4
	delay := time.Duration(delayMs) * time.Millisecond

	fmt.Printf("Doorbell triggered: status=%s, voice_length=%dms, delay=%dms\n",
		data.Status, data.VoiceLengthMs, delayMs)

	return delay, nil
}
