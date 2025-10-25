#!/bin/bash

ssh adam@gate-unlocker.local "sudo systemctl stop gate-locker.service"
scp gate-locker-arm64 adam@gate-unlocker.local:/opt/gate-locker/gate-locker
ssh adam@gate-unlocker.local "sudo systemctl start gate-locker.service"