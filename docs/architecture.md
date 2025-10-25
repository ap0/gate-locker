# Architecture

## Overview

The gate locker system consists of two independent but cooperating devices:

- **ESP32 (MicroPython)**: Handles physical gate control, button input, LED status, and time-based restrictions
- **Raspberry Pi (Go)**: Provides API interface, doorbell integration, and intercepts button presses for doorbell delay

## Operating Modes

### Two-Device Mode (Pi Connected)

When the Pi is connected (ESP32 GPIO 32 pulled to ground):

1. User presses physical button
2. ESP32 detects button press
3. ESP32 signals Pi via GPIO 26 (HIGH)
4. Pi receives signal and calls doorbell API
5. Pi calculates delay (1/4 of doorbell voice length)
6. Pi waits for delay period
7. Pi signals ESP32 via GPIO 15→13 (HIGH pulse)
8. ESP32 receives response and unlocks gate
9. ESP32 blinks green LED

**LEDs in two-device mode:** Always green (time window ignored)

### Standalone Mode (No Pi)

When the Pi is NOT connected (ESP32 GPIO 32 pulled HIGH):

1. User presses physical button
2. ESP32 checks time window (6:00am - 8:30pm)
3. If allowed: ESP32 unlocks gate, blinks green LED
4. If denied: ESP32 blinks red LED (no unlock)

**LEDs in standalone mode:** Green during time window, red outside time window

### API Unlock Mode

When unlock is triggered via REST API:

1. API call to Pi `/unlock` endpoint
2. Pi activates relay on GPIO 25 (HIGH)
3. Relay contacts close, pulling ESP32 GPIO 25 to ground
4. ESP32 detects button press (same as physical button)
5. Follows two-device mode flow (no time restrictions)

## Power Management

### ESP32 Light Sleep

- **Active time**: ~0.17% (wakes every 60 seconds for 100ms)
- **Sleep current**: ~0.8mA
- **Wake sources**:
  - Button press (EXT0 on GPIO 25)
  - 60-second timeout for LED updates and time checks

### Daily NTP Sync

- Syncs at boot
- Syncs daily between 2:00-2:10am
- WiFi connects → sync time → WiFi disconnects
- Maintains accurate timekeeping for time windows

## Communication Protocol

### ESP32 → Pi (Button Signal)

```
ESP32 GPIO 26: ─────┐         ┌──────
                     └─────────┘
                     ^button press
```

Pi polls GPIO 26 every 100ms to detect signal.

### Pi → ESP32 (Unlock Response)

```
Pi GPIO 15:    ─────┐   ┌──────
                     └───┘
                     ^100ms pulse
```

ESP32 waits up to 5 seconds for response, then unlocks regardless.

## Time Synchronization

- **NTP sync**: Uses WiFi to sync with NTP servers
- **Timezone**: Configured as offset from UTC (e.g., -7 for PDT)
- **RTC**: Maintains time during light sleep
- **Accuracy**: Re-syncs daily to prevent drift

## Unlock Mechanism

ESP32 controls gate unlock relay:
1. Set GPIO 27 HIGH (500ms)
2. Gate unlock relay activates
3. Set GPIO 27 LOW
4. Blink green LED (3000ms) concurrently via callback

The unlock duration (500ms) completes during the LED blink (3000ms) for efficient, responsive operation.
