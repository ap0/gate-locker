#!/bin/bash
set -e

echo "Stopping gate-locker service..."
sudo systemctl stop gate-locker

echo "Backing up current binary..."
sudo cp /opt/gate-locker/gate-locker /opt/gate-locker/gate-locker.backup

echo "Building new version..."
cd /opt/gate-locker
go build -o gate-locker

echo "Starting gate-locker service..."
sudo systemctl start gate-locker

echo "Checking service status..."
sleep 2
sudo systemctl status gate-locker --no-pager

echo "Update complete!"