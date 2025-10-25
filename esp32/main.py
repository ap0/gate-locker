"""
Gate Locker for ESP32 - MicroPython
Simplified version with WiFi for NTP sync only
"""

import machine
import time
import ntptime
import network
from machine import Pin, RTC
from esp32 import wake_on_ext0
from secrets import WIFI_SSID, WIFI_PASSWORD, TIMEZONE_OFFSET as TZ_OFFSET


# ============================================
# CONFIGURATION
# ============================================

# GPIO Pin Configuration
# Using RTC GPIO pins for outputs that need to stay on during deep sleep
# RTC GPIOs on ESP32: 0, 2, 4, 12, 13, 14, 15, 25, 26, 27, 32, 33, 34, 35, 36, 39
RED_LED_PIN = 33      # RTC GPIO - stays active during sleep
GREEN_LED_PIN = 14    # RTC GPIO - stays active during sleep
BUTTON_PIN = 25       # RTC GPIO - wakes from deep sleep
UNLOCK_PIN = 27       # RTC GPIO - stays active during sleep
ONBOARD_LED_PIN = 2   # Built-in LED on most ESP32 boards

# Two-Device Communication Pins
DEVICE_DETECT_PIN = 32   # RTC GPIO - input to detect if other device is connected
DEVICE_SIGNAL_PIN = 26   # RTC GPIO - output to signal button press to other device
DEVICE_RESPONSE_PIN = 13 # RTC GPIO - input to receive unlock signal from other device
DEVICE_RESPONSE_TIMEOUT_MS = 5000  # Timeout waiting for other device (5 seconds)

# Time Window (24-hour format)
START_HOUR = 6
START_MINUTE = 0
END_HOUR = 20
END_MINUTE = 30

# Durations (milliseconds)
UNLOCK_DURATION_MS = 500
BLINK_INTERVAL_MS = 250
BLINK_TOTAL_DURATION_MS = 3000

# Button debounce time (milliseconds)
DEBOUNCE_MS = 250

# ============================================
# Global State
# ============================================

class GateController:
    def __init__(self):
        # Initialize GPIO pins
        self.button = Pin(BUTTON_PIN, Pin.IN, Pin.PULL_UP)
        self.red_led = Pin(RED_LED_PIN, Pin.OUT)
        self.green_led = Pin(GREEN_LED_PIN, Pin.OUT)
        self.onboard_led = Pin(ONBOARD_LED_PIN, Pin.OUT)
        self.unlock_pin = Pin(UNLOCK_PIN, Pin.OUT)

        # Two-device communication pins
        self.device_detect = Pin(DEVICE_DETECT_PIN, Pin.IN, Pin.PULL_UP)
        self.device_signal = Pin(DEVICE_SIGNAL_PIN, Pin.OUT)
        self.device_response = Pin(DEVICE_RESPONSE_PIN, Pin.IN, Pin.PULL_UP)

        # Set initial states (active HIGH for LEDs and relays)
        self.red_led.value(0)    # OFF (active HIGH)
        self.green_led.value(0)  # OFF (active HIGH)
        self.onboard_led.value(0) # OFF (active HIGH for onboard LED)
        self.unlock_pin.value(0) # Locked (active HIGH)
        self.device_signal.value(0)  # Signal OFF initially

        # State tracking
        self.last_button_time = 0
        self.in_button_handler = False
        self.is_blinking = False
        self.unlocking = False

        # RTC for timekeeping
        self.rtc = RTC()

        # Setup button interrupt with wake capability for light sleep
        self.button.irq(trigger=Pin.IRQ_FALLING, handler=self.button_handler, wake=machine.SLEEP)

        # Track last NTP sync day for daily 2am sync
        self.last_sync_day = None

        print("Gate Controller initialized")
        print(f"Button: GPIO{BUTTON_PIN}, Red LED: GPIO{RED_LED_PIN}, Green LED: GPIO{GREEN_LED_PIN}")
        print(f"Unlock: GPIO{UNLOCK_PIN}")
        print(f"Device Detect: GPIO{DEVICE_DETECT_PIN}, Signal: GPIO{DEVICE_SIGNAL_PIN}, Response: GPIO{DEVICE_RESPONSE_PIN}")
        print(f"Time window: {START_HOUR:02d}:{START_MINUTE:02d} - {END_HOUR:02d}:{END_MINUTE:02d}")
        print("Failsafe: Wired directly to ESP32 power supply")

    def is_device_connected(self):
        """Check if another device is connected (detect pin pulled low)"""
        return self.device_detect.value() == 0  # Active LOW

    def is_in_button_handler(self):
        return self.in_button_handler

    def is_unlock_allowed(self):
        """Check if unlock is allowed based on time window"""
        # Get current local time
        year, month, day, weekday, hour, minute, second, subsecond = self.rtc.datetime()

        # Convert to minutes since midnight for easier comparison
        current_minutes = hour * 60 + minute
        start_minutes = START_HOUR * 60 + START_MINUTE
        end_minutes = END_HOUR * 60 + END_MINUTE

        return start_minutes <= current_minutes < end_minutes

    def wait_for_device_response(self):
        """Wait for other device to signal unlock (with timeout)
        Returns True if response received, False if timeout"""
        print("Waiting for device response...")
        start_time = time.ticks_ms()

        while time.ticks_diff(time.ticks_ms(), start_time) < DEVICE_RESPONSE_TIMEOUT_MS:
            # Check if response pin is HIGH (signal to unlock)
            if self.device_response.value() == 1:
                elapsed = time.ticks_diff(time.ticks_ms(), start_time)
                print(f"Device response received after {elapsed}ms")
                return True

            # Small sleep to avoid busy-waiting
            time.sleep_ms(10)

        print(f"Device response timeout after {DEVICE_RESPONSE_TIMEOUT_MS}ms")
        return False

    def set_red_led(self, on):
        """Set red LED state (active HIGH) - ensures green is off when red is on"""
        self.red_led.value(1 if on else 0)
        if on:
            self.green_led.value(0)  # Turn off green

    def set_green_led(self, on):
        """Set green LED state (active HIGH) - ensures red is off when green is on"""
        self.green_led.value(1 if on else 0)
        if on:
            self.red_led.value(0)  # Turn off red

    def blink_led(self, led_func, interval_ms, total_duration_ms, callback=None):
        """Blink an LED for a specified duration with optional callback during sleep"""
        self.is_blinking = True
        start_time = time.ticks_ms()
        half_interval = interval_ms // 2

        while time.ticks_diff(time.ticks_ms(), start_time) < total_duration_ms:
            led_func(True)
            self.onboard_led.value(1)  # Onboard LED ON (active HIGH)

            # Break into smaller sleeps to feed watchdog more frequently
            remaining = half_interval
            while remaining > 0:
                sleep_time = min(100, remaining)
                time.sleep(sleep_time / 1000.0)
                if callback:
                    callback()
                remaining -= sleep_time

            led_func(False)
            self.onboard_led.value(0)  # Onboard LED OFF

            # Break into smaller sleeps to feed watchdog more frequently
            remaining = half_interval
            while remaining > 0:
                sleep_time = min(100, remaining)
                time.sleep(sleep_time / 1000.0)
                if callback:
                    callback()
                remaining -= sleep_time

        self.is_blinking = False

    def blink_green(self, callback=None):
        """Blink green LED"""
        self.set_red_led(False)
        self.blink_led(self.set_green_led, BLINK_INTERVAL_MS, BLINK_TOTAL_DURATION_MS, callback)

    def blink_red(self, callback=None):
        """Blink red LED"""
        self.set_green_led(False)
        self.blink_led(self.set_red_led, BLINK_INTERVAL_MS, BLINK_TOTAL_DURATION_MS, callback)

    def unlock_gate(self):
        """Unlock the gate for specified duration"""
        if self.unlocking:
            print("Unlock already in progress")
            return

        self.unlocking = True
        print("Unlocking gate")

        # Trigger unlock relay
        self.unlock_pin.value(1)  # Active HIGH
        unlock_start_time = time.ticks_ms()
        unlock_completed = [False]  # Use list to allow modification in nested function

        def check_unlock_duration():
            """Callback to turn off unlock pin after duration"""
            if not unlock_completed[0] and time.ticks_diff(time.ticks_ms(), unlock_start_time) >= UNLOCK_DURATION_MS:
                self.unlock_pin.value(0)  # Lock again (active HIGH - off)
                unlock_completed[0] = True

        # Blink green LED while checking unlock duration in callback
        self.blink_green(callback=check_unlock_duration)

        # Ensure unlock pin is off (in case blink was shorter than unlock duration)
        if not unlock_completed[0]:
            self.unlock_pin.value(0)

        print("Gate unlock complete")
        self.unlocking = False

    def button_handler(self, pin):
        """Button interrupt handler - uses debouncing without disabling interrupt"""
        # Restore CPU frequency after waking from light sleep
        # Light sleep may reduce CPU freq, restore to full speed
        machine.freq(240000000)  # 240 MHz

        # Critical: Small delay after freq change to let system stabilize
        # Without this, GPIO reads and timing can be unreliable
        time.sleep_ms(5)

        print("IRQ fired")

        # Skip if already processing - this catches any queued interrupts
        if self.in_button_handler:
            print("Already processing")
            return

        # Debounce check - ignore if too soon after last press
        current_time = time.ticks_ms()
        time_since_last = time.ticks_diff(current_time, self.last_button_time)
        if time_since_last < DEBOUNCE_MS:
            print(f"Debounced: {time_since_last}ms")
            return

        # Wait a bit and check if button is still pressed (debounce)
        time.sleep_ms(50)
        if self.button.value() == 1:  # Button released (active LOW)
            print("Button not held, ignoring")
            return

        # Mark as processing and update last button time
        self.last_button_time = current_time
        self.in_button_handler = True

        try:
            print("Button pressed")

            # Check if another device is connected
            device_connected = self.is_device_connected()
            print(f"Device connected: {device_connected}")

            if device_connected:
                # Two-device mode: signal the other device and wait for response
                print("Two-device mode: signaling other device")

                # Set signal pin HIGH to indicate button press
                self.device_signal.value(1)

                # Wait for other device to respond
                response_received = self.wait_for_device_response()
                print(f"Response received: {response_received}")

                # Clear signal pin
                self.device_signal.value(0)

                # Unlock the gate (either on response or timeout)
                print("Unlocking gate (two-device mode)")
                self.unlock_gate()
            else:
                # Single-device mode: use time-based rules
                # Check if unlock is allowed
                if not self.is_unlock_allowed():
                    print("Button unlock denied - outside time window")
                    self.blink_red()
                else:
                    # Unlock the gate
                    self.unlock_gate()
        finally:
            # Small delay to let button settle
            time.sleep(0.3)

            # Mark as done processing
            self.in_button_handler = False
            # print("Handler complete, ready for next press")

    def check_and_sync_time(self):
        """Check if it's time for daily NTP sync (around 2am) and sync if needed"""
        year, month, day, weekday, hour, minute, second, subsecond = self.rtc.datetime()

        # Check if it's around 2am (2:00-2:10) and we haven't synced today
        if hour == 2 and minute < 10:
            # Use day as sync marker
            if self.last_sync_day != day:
                print(f"Daily NTP sync triggered at {hour:02d}:{minute:02d}")
                sync_time()
                self.last_sync_day = day

    def get_status(self):
        """Get current system status"""
        year, month, day, weekday, hour, minute, second, subsecond = self.rtc.datetime()
        unlock_state = "ALLOWED" if self.is_unlock_allowed() else "DENIED"

        print(f"\n{'='*40}")
        print(f"Unlock Status: {unlock_state}")
        print(f"Time: {hour:02d}:{minute:02d}:{second:02d}")
        print(f"Date: {year}-{month:02d}-{day:02d}")
        print(f"{'='*40}\n")



# ============================================
# WiFi and NTP Functions
# ============================================

def connect_wifi():
    """Connect to WiFi for NTP sync"""
    wlan = network.WLAN(network.STA_IF)
    wlan.active(True)

    if wlan.isconnected():
        print("Already connected to WiFi")
        return wlan

    print(f"Connecting to WiFi: {WIFI_SSID}")
    wlan.connect(WIFI_SSID, WIFI_PASSWORD)

    # Wait for connection (timeout after 10 seconds)
    timeout = 10
    while not wlan.isconnected() and timeout > 0:
        time.sleep(1)
        timeout -= 1
        print(".", end="")

    if wlan.isconnected():
        print(f"Connected! IP: {wlan.ifconfig()[0]}")
        return wlan
    else:
        print("Failed to connect to WiFi")
        return None

def sync_time():
    """Sync time via NTP - connects WiFi, syncs, then disconnects"""
    print("\n--- Starting NTP Time Sync ---")

    # Connect WiFi
    wlan = connect_wifi()
    if not wlan:
        print("Cannot sync time - WiFi connection failed")
        return False

    # Sync with NTP server
    try:
        print("Syncing with NTP server...")
        ntptime.settime()

        # Get current time and apply timezone offset
        rtc = RTC()
        year, month, day, weekday, hour, minute, second, subsecond = rtc.datetime()

        # Apply timezone offset
        hour = (hour + TZ_OFFSET) % 24

        # Set RTC with corrected time (year, month, day, weekday, hour, minute, second, subsecond)
        rtc.datetime((year, month, day, weekday, hour, minute, second, subsecond))

        print(f"Time synced: {year}-{month:02d}-{day:02d} {hour:02d}:{minute:02d}:{second:02d}")
        print("--- NTP Sync Complete ---\n")

        return True

    except Exception as e:
        print(f"NTP sync failed: {e}")
        return False

    finally:
        # Always disconnect WiFi after sync attempt
        print("Disconnecting WiFi...")
        if wlan.isconnected():
            wlan.disconnect()
        wlan.active(False)
        print("WiFi disconnected")


# ============================================
# Main Program
# ============================================

def main():
    print("\n" + "="*50)
    print("Gate Locker ESP32 - MicroPython")
    print("="*50 + "\n")

    # Print current RTC time (before sync)
    rtc = RTC()
    year, month, day, weekday, hour, minute, second, subsecond = rtc.datetime()
    print(f"Current time (before sync): {year}-{month:02d}-{day:02d} {hour:02d}:{minute:02d}:{second:02d}")
    print()

    # Always sync NTP on boot (using lightsleep only, so any boot is a real boot)
    sync_time()

    # Initialize gate controller
    controller = GateController()

    # Show initial status
    controller.get_status()

    print("\nGate controller running...")
    print("Press CTRL+C to stop\n")

    try:
        # Use light sleep for power savings
        print("Entering light sleep mode...")
        print("Configuring GPIO wake from light sleep...")

        # Set initial LED state based on device connection or time window
        if controller.is_device_connected():
            # Device connected: always green
            controller.set_green_led(True)
            controller.set_red_led(False)
        else:
            # No device: use time window
            allowed = controller.is_unlock_allowed()
            controller.set_green_led(allowed)
            controller.set_red_led(not allowed)

        # Configure wake sources for light sleep
        # Button pin is configured via IRQ with wake=machine.SLEEP
        # EXT0 provides additional wake capability for button
        wake_on_ext0(pin=controller.button, level=0)  # Wake when button goes LOW
        print(f"Wake sources: Button (GPIO{BUTTON_PIN})")

        time.sleep(0.1)
        while True:
            # Enter light sleep - wake on button press or timeout
            machine.lightsleep(60000)  # Sleep for 1 minute or until GPIO interrupt
            # After wake, restore CPU frequency immediately
            machine.freq(240000000)

            # Check if it's time for daily NTP sync (around 2am)
            controller.check_and_sync_time()

            # Update LED state after wake to reflect device connection or time window
            if not controller.is_blinking and not controller.unlocking:
                if controller.is_device_connected():
                    # Device connected: always green
                    controller.set_green_led(True)
                    controller.set_red_led(False)
                else:
                    # No device: use time window
                    allowed = controller.is_unlock_allowed()
                    controller.set_green_led(allowed)
                    controller.set_red_led(not allowed)

    except KeyboardInterrupt:
        print("\n\nShutting down...")
        controller.set_red_led(False)
        controller.set_green_led(False)
        controller.unlock_pin.value(0)  # Ensure locked on shutdown
        print("Cleanup complete")


# ============================================
# Entry Point
# ============================================

if __name__ == "__main__":
    main()
