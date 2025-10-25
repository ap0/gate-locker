from machine import Pin
import time

# Configure GPIO 33
red_pin = Pin(33, Pin.OUT)
green_pin = Pin(14, Pin.OUT)
unlock_pin = Pin(27, Pin.OUT)

print("Toggling GPIO 33 every 2 seconds")
print("Measure voltage at SSR (-) relative to GND")
print("Press Ctrl+C to stop\n")

try:
    while True:
        print("GPIO 33 HIGH (3.3V)")
        red_pin.value(1)
        green_pin.value(0)
        unlock_pin.value(0)
        time.sleep(0.5)
        print("GPIO 33 LOW (0V)")
        red_pin.value(0)
        green_pin.value(1)
        unlock_pin.value(1)
        time.sleep(0.5)
except KeyboardInterrupt:
    print("\nStopped. Setting GPIO 33 LOW")
    red_pin.value(0)