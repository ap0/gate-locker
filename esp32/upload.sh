#!/bin/bash

SERIAL_PORT="/dev/cu.usbserial-001"

if [ -z "$1" ]; then

    export ARG="main.py"
    echo "No argument provided!"
    exit 1
else
    export ARG=$1
fi

mpremote connect /dev/cu.usbserial-0001 cp $ARG : && mpremote connect /dev/cu.usbserial-0001