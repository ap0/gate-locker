#!/bin/bash

esptool.py --port /dev/cu.usbserial-0001 erase_flash
esptool.py --baud 460800 write_flash -z 0x1000 ~/Downloads/ESP32_GENERIC-20250911-v1.26.1.bin
#mpremote mip install github:peterhinch/micropython-mqtt