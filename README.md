# USB Button (RPi Pico)

![usb-hid](https://github.com/user-attachments/assets/c9b970fd-0b9c-4d87-8522-7a2e662e0449)


Firmware for a Raspberry Pi Pico with a physical button: short / double / long press are sent to the host as **raw HID reports** (consumer report ID 3 with a 16-bit event code). An LED on GPIO 15 blinks until the host signals ready, then follows host commands. You can customize the report payload and, with a TinyGo override, VID/PID and report format.

The repo is split into **firmware/** (TinyGo, for the Pico) and **host/** (regular Go, for the listener on your computer).

## Quick start

- **Hardware:** Button on GPIO 14, LED on GPIO 15 (see `firmware/internal/button` and `firmware/internal/led`).
- **Build & flash firmware:**  
  `make build-fw && make flash`
- **Host:** Build and run the listener (standard Go, not TinyGo):
  - `make build-host` then `./.bin/hid_listener`  
  - Use `-list` to list HID devices. Reports are 3 bytes: `[0x03, event_lo, event_hi]`.

## Documentation

- **[docs/HID_PROTOCOL.md](docs/HID_PROTOCOL.md)** – Explains the HID protocol: enumeration, report descriptor, and what the code does.
- **[docs/TINYGO_DEFAULTS.md](docs/TINYGO_DEFAULTS.md)** – Explains what part of the the USB HID communication is handled by TinyGo for us
- **[host/README.md](host/README.md)** - Describes the host code in more detail
- **[firmware/README.md](firmware/README.md)** - Describes the firmware code in more detail

## Troubleshooting

### macOS: "Failed to connect to the device" (HID open failed)

1. **System Settings → Privacy & Security → Input Monitoring**  
   Add Terminal (or Cursor / your IDE), enable it, then quit and reopen the app.
2. If it still fails, run the listener with sudo (macOS sometimes requires root for HID):  
   `sudo ./.bin/hid_listener`
3. Unplug the Pico, plug it back in, then run again.
