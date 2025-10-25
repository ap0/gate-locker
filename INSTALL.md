# Gate Locker Installation Guide

This guide will help you set up the Gate Locker service on a Raspberry Pi Zero 2 W.

## Prerequisites

- Raspberry Pi Zero 2 W with SD card (8GB+)
- Network connection (WiFi or Ethernet)
- SSH access to the Pi

## 1. Prepare Raspberry Pi

### Flash Raspberry Pi OS
1. Download and flash **Raspberry Pi OS Lite** to your SD card using Raspberry Pi Imager
2. Before ejecting, enable SSH by creating an empty file named `ssh` in the boot partition
3. Configure WiFi by creating `wpa_supplicant.conf` in the boot partition:
   ```
   country=US
   ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
   update_config=1
   
   network={
       ssid="Your-WiFi-Name"
       psk="Your-WiFi-Password"
   }
   ```

### Initial Setup
1. Insert SD card and boot the Pi
2. Find the Pi's IP address (check your router or use `nmap`)
3. SSH into the Pi: `ssh adam@<pi-ip-address>`
4. Change default password: `passwd`

## 2. Update System

```bash
# Update system
sudo apt update && sudo apt upgrade -y
```

## 3. Deploy Gate Locker

### Create Installation Directory
```bash
sudo mkdir -p /opt/gate-locker
sudo chown adam:adam /opt/gate-locker
```

### Cross-Compile and Deploy
From your development machine:
```bash
# Cross-compile for Raspberry Pi Zero 2 W (ARM64)
cd /Users/adam/src/gate-locker
GOOS=linux GOARCH=arm64 go build -o gate-locker-arm64

# Copy binary and config files to Pi
scp gate-locker-arm64 adam@<pi-ip-address>:/opt/gate-locker/gate-locker
scp config.json.pi adam@<pi-ip-address>:/opt/gate-locker/
scp gate-locker.service adam@<pi-ip-address>:/opt/gate-locker/
scp update-gate.sh adam@<pi-ip-address>:/opt/gate-locker/
```

## 4. Configure the Service

### Set Up Configuration
```bash
cd /opt/gate-locker
cp config.json.pi config.json
nano config.json
```

Update the `homebridge.url` field with your Homebridge server IP:
```json
"homebridge": {
  "enabled": true,
  "url": "http://YOUR_HOMEBRIDGE_IP:8081/status"
}
```

### Install Systemd Service
```bash
sudo cp gate-locker.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable gate-locker.service
```

### Configure GPIO Permissions
```bash
sudo usermod -a -G gpio adam
```

## 5. Start the Service

```bash
# Start the service
sudo systemctl start gate-locker.service

# Check status
sudo systemctl status gate-locker.service

# View logs
sudo journalctl -u gate-locker -f
```

## 6. GPIO Connections

Connect your hardware to these GPIO pins:

| Component | GPIO Pin | Physical Pin |
|-----------|----------|--------------|
| Button | GPIO 18 | Pin 12 |
| Red LED | GPIO 23 | Pin 16 |
| Green LED | GPIO 24 | Pin 18 |
| Unlock Relay | GPIO 25 | Pin 22 |
| Reed Switch | GPIO 21 | Pin 40 |

**Important:** All GPIO outputs should be connected through appropriate resistors (typically 220Ω for LEDs) and the relay should be driven through a transistor circuit for protection.

## 7. Testing

### Test API Endpoints
```bash
# Check status
curl http://localhost:8080/status

# Test unlock
curl -X POST http://localhost:8080/unlock

# Test command endpoint (for Homebridge)
curl -X POST http://localhost:8080/command \
  -H "Content-Type: application/json" \
  -d '{"command":"unlock"}'
```

### Test Mock Mode (Development)
If you want to test without GPIO hardware:
```bash
# Stop the service
sudo systemctl stop gate-locker

# Run in mock mode
cd /opt/gate-locker
./gate-locker

# Type commands: press, open, close, status, quit
```

## 8. Maintenance

### View Logs
```bash
# View recent logs
sudo journalctl -u gate-locker --no-pager

# Follow logs in real-time
sudo journalctl -u gate-locker -f

# View logs from specific time
sudo journalctl -u gate-locker --since "1 hour ago"
```

### Update the Service
```bash
cd /opt/gate-locker
chmod +x update-gate.sh
./update-gate.sh
```

### Manual Service Control
```bash
# Start service
sudo systemctl start gate-locker

# Stop service
sudo systemctl stop gate-locker

# Restart service
sudo systemctl restart gate-locker

# Check service status
sudo systemctl status gate-locker

# Disable auto-start
sudo systemctl disable gate-locker

# Re-enable auto-start
sudo systemctl enable gate-locker
```

## 9. Troubleshooting

### Service Won't Start
1. Check logs: `sudo journalctl -u gate-locker`
2. Verify Go installation: `go version`
3. Test manual run: `cd /opt/gate-locker && ./gate-locker`
4. Check file permissions: `ls -la /opt/gate-locker/`

### GPIO Permission Issues
```bash
# Add pi user to gpio group
sudo usermod -a -G gpio adam

# Reboot to apply changes
sudo reboot
```

### Network Connectivity Issues
1. Verify Pi can reach Homebridge: `ping YOUR_HOMEBRIDGE_IP`
2. Check port accessibility: `telnet YOUR_HOMEBRIDGE_IP 8081`
3. Verify Homebridge plugin is running and listening on port 8081

### Configuration Issues
1. Validate JSON: `cat config.json | python3 -m json.tool`
2. Check all required fields are present
3. Verify GPIO pin numbers match your hardware setup

## 10. Auto-Start Verification

The service is configured to:
- Start automatically on boot
- Restart automatically if it crashes (up to 5 times in 5 minutes)
- Wait for network connectivity before starting
- Log all output to systemd journal

To verify auto-start is working:
```bash
sudo reboot
# Wait for Pi to boot
ssh adam@<pi-ip-address>
sudo systemctl status gate-locker
```

The service should show as "active (running)".

## Security Notes

- Change the default Pi password
- Consider setting up SSH key authentication
- The service runs as the `adam` user (not root) for security
- Only necessary ports (8080) are exposed
- GPIO access is limited to the `gpio` group