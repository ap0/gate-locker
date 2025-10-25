# Pin Mapping

## Two-Device Communication

| Connection | Raspberry Pi | ESP32 | Description |
|------------|--------------|-------|-------------|
| Button signal | GPIO 26 (input) | GPIO 26 (output) | ESP32 signals Pi when button pressed |
| Unlock response | GPIO 15 (output) | GPIO 13 (input) | Pi signals ESP32 to unlock after doorbell delay |
| Device detection | GND | GPIO 32 (input, pull-up) | ESP32 detects if Pi is connected |

## API Unlock (Relay)

| Connection | Raspberry Pi | ESP32 | Description |
|------------|--------------|-------|-------------|
| Relay control | GPIO 25 (output) | - | Pi drives relay coil |
| Relay contacts (C/NO) | - | GPIO 25 | Relay contacts simulate button press |

**Note:** The relay provides electrical isolation between Pi and ESP32. When Pi GPIO 25 goes HIGH, the relay closes contacts that pull ESP32 GPIO 25 to ground (same as physical button).

## ESP32 Local Pins (No Pi Connection)

| Pin | Function | Type | Description |
|-----|----------|------|-------------|
| GPIO 25 | Button | Input (pull-up, active LOW) | Physical button and relay input |
| GPIO 27 | Unlock relay | Output (active HIGH) | Controls gate unlock mechanism |
| GPIO 33 | Red LED | Output (active HIGH) | Status indicator (outside time window) |
| GPIO 14 | Green LED | Output (active HIGH) | Status indicator (inside time window) |
| GPIO 2 | Onboard LED | Output (active HIGH) | Built-in LED for activity indication |

## Signal Logic

- **ESP32 GPIO 26 (signal)**: HIGH when button pressed
- **Pi GPIO 15 → ESP32 GPIO 13 (response)**: HIGH to signal unlock
- **ESP32 GPIO 32 (detect)**: LOW when Pi connected, HIGH when standalone
- **Pi GPIO 25 (relay)**: HIGH to activate relay
- **Button/Relay → ESP32 GPIO 25**: Pulled to ground when pressed/activated
