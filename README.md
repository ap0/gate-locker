# Gate Locker

A dual-device gate control system with doorbell integration, time-based access control, and remote API unlocking.

## System Components

### ESP32 (MicroPython)
- **Location**: `esp32/`
- **Language**: MicroPython
- **Role**: Physical gate control, button handling, LED status indicators, time-based restrictions
- **Power**: Light sleep mode (~0.8mA) with instant button wake

### Raspberry Pi (Go)
- **Location**: `pi/`
- **Language**: Go
- **Role**: REST API server, doorbell integration, button press interception for doorbell delay
- **API Port**: 8080

## Features

### Two-Device Mode
- Physical button press triggers doorbell
- Configurable delay before gate unlocks (1/4 of doorbell voice length)
- ESP32 communicates with Pi via GPIO
- Always shows green LED (ignores time restrictions)

### Standalone Mode
- Works without Raspberry Pi
- Time-based access control (6:00am - 8:30pm)
- Red/green LED indicates allowed/denied status
- No doorbell integration

### API Unlock
- REST API endpoint: `POST http://gate-unlocker.local:8080/unlock`
- Pi triggers relay that simulates button press
- Bypasses time restrictions
- Follows doorbell delay flow in two-device mode

## Documentation

- **[Pin Mapping](docs/pin-mapping.md)**: Complete GPIO pin connections between devices
- **[Architecture](docs/architecture.md)**: Detailed system design and operation modes

## Quick Start

### ESP32 Setup
```bash
cd esp32

# 1. Create secrets file from template
cp secrets.py.example secrets.py

# 2. Edit secrets.py with your WiFi credentials and timezone
# WIFI_SSID, WIFI_PASSWORD, TIMEZONE_OFFSET

# 3. Flash files to ESP32
ampy --port /dev/cu.SLAB_USBtoUART put boot.py
ampy --port /dev/cu.SLAB_USBtoUART put secrets.py
ampy --port /dev/cu.SLAB_USBtoUART put main.py
```

### Raspberry Pi Setup
```bash
cd pi
go build -o gate-locker
./gate-locker
# Or run as systemd service
```

### API Usage
```bash
# Check status
curl http://gate-unlocker.local:8080/status

# Unlock gate
curl -X POST http://gate-unlocker.local:8080/unlock
```

## Configuration

### ESP32 (esp32/secrets.py)
**Create from template:** `cp secrets.py.example secrets.py`

```python
WIFI_SSID = "your_ssid_here"
WIFI_PASSWORD = "your_password_here"
TIMEZONE_OFFSET = -7  # Hours from UTC (e.g., -8 for PST, -7 for PDT)
```

**Note:** `secrets.py` is gitignored and will not be committed. Never commit WiFi credentials!

### ESP32 (esp32/main.py)
- Time window: `START_HOUR`, `START_MINUTE`, `END_HOUR`, `END_MINUTE`
- GPIO pins: See pin configuration section
- Other timing constants: `UNLOCK_DURATION_MS`, `BLINK_INTERVAL_MS`, etc.

### Raspberry Pi (pi/config.json)
```json
{
  "gpio": {
    "esp_signal_pin": 26,
    "esp_response_pin": 15,
    "unlock_relay_pin": 25
  },
  "api": {
    "port": ":8080"
  },
  "doorbell": {
    "enabled": false,
    "host": "lightning-control.local:8080"
  }
}
```

## Hardware Requirements

- ESP32 development board
- Raspberry Pi (any model with GPIO)
- 2x LEDs (red/green) with current-limiting resistors
- 2x relays (gate unlock + API unlock simulation)
- MOSFETs for relay control
- Physical button (normally open, active LOW)
- Wiring between Pi and ESP32 GPIO pins

## Power Consumption

- **ESP32 Light Sleep**: ~0.8mA
- **ESP32 Active**: ~80mA (during unlock/LED blink)
- **Wake Frequency**: Every 60 seconds (100ms wake)
- **NTP Sync**: Once at boot, then daily at ~2am
- **Expected Battery Life**: Months to years (depending on battery capacity and usage)

## Project Structure

```
gate-locker/
├── pi/                    # Raspberry Pi Go code
│   ├── main.go
│   ├── config.go
│   ├── gpio.go
│   ├── api.go
│   └── config.json
├── esp32/                 # ESP32 MicroPython code
│   ├── main.py
│   ├── boot.py
│   ├── secrets.py.example # Template for WiFi credentials
│   └── secrets.py         # (gitignored - create from example)
├── docs/
│   ├── pin-mapping.md
│   └── architecture.md
└── README.md
```

## License

[Your License Here]
