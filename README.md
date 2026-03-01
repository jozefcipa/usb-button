# USB Button (RPi Pico)

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

## Layout

- **firmware/** – TinyGo code for the Pico (separate Go module).
  - `main.go` – Entry point: button/LED/HID init, then loop that polls host reports, LED blink, and button; sends consumer reports on press.
  - `internal/button` – Button and press-type detection (short / double / long).
  - `internal/led` – LED state and blink vs host-driven on/off.
  - `internal/hid` – Sending raw HID report bytes (consumer, keyboard output handling) and receiving host reports.
- **host/** – Regular Go code for your computer (separate Go module).
  - `main.go` – Finds the Pico HID device, prints reports, runs Lua hooks, can send output (e.g. LED), supports daemon mode.
  - `internal/hid` – HID open, read, write.
  - `internal/hooks` – Lua script loading and handler dispatch (e.g. `hid_listener.example.lua`).
  - `internal/cli` – CLI flags
  - `internal/daemon` – daemon start/stop.

- **protocol/** – Shared constants (report IDs, button/led usage values) used by firmware and host.

## Troubleshooting

### macOS: "Failed to connect to the device" (HID open failed)

1. **System Settings → Privacy & Security → Input Monitoring**  
   Add Terminal (or Cursor / your IDE), enable it, then quit and reopen the app.
2. If it still fails, run the listener with sudo (macOS sometimes requires root for HID):  
   `sudo ./.bin/hid_listener`
3. Unplug the Pico, plug it back in, then run again.