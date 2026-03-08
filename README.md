# USB Button
![USB HID with RPi Pico](https://github.com/user-attachments/assets/16ce869d-3b2f-4b0b-a99f-3cc95a9b11a8)

> USB Button is a simple project built on a Raspberry Pi Pico with a physical button that talks to your computer over *USB HID*. Press the button (short, double, or long) and the Pico sends those events to a host application on your PC or Mac.
> 
> The host application listens for those events and acts upon them by executing custom actions defined in *Lua*.

## Quick start

The repo has two parts: `firmware/` (TinyGo on the Pico) and `host/` (regular Go on your computer) for talking to the custom hardware.


- **Hardware:** Button on GPIO 14, LED on GPIO 15 (see [button.go](./firmware/internal/button/button.go) and [led.go](./firmware//internal//led//led.go)).
- **Build & flash firmware:**  
  `make build-fw && make flash`
- **Host:** Build and run the listener (standard Go, not TinyGo):
  - `make build-host` then `sudo .bin/hid_listener`
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
   `sudo .bin/hid_listener`
3. Unplug the Pico, plug it back in, then run again.
