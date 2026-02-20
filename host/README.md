# hid_listener

`hid_listener` is a host application for interacting with the USB button device (Raspberry Pi Pico). It can:

- Find the Pico USB button device.
- Print HID input reports as they arrive (e.g., short, double, or long presses).
- Send HID output reports to the Pico (host → device).

## Usage

> **Note:** On Linux, you may need elevated permissions or custom udev rules.  
> On macOS, if open fails, add Terminal (or your preferred app) to **System Settings → Privacy & Security → Input Monitoring**.

### Common commands
(`hid_listener` binary can be found in `../.bin` directory after running `make build-host`)

```sh
./hid_listener                    # Listen for button reports
./hid_listener -send 0201         # Send Report ID 2, payload 0x01 (turn Pico LED on), then exit
./hid_listener -daemon            # Run listener in background (PID written to ~/.cache/hid_listener.pid)
./hid_listener stop               # Stop the background daemon (sends SIGTERM to PID from file)
```

### Options

- `-vid` / `-pid`: Override the USB vendor/product ID (defaults: 0x2E8A, 0x000A)
- `-list`: List HID devices and exit

## Library

Uses [github.com/bearsh/hid](https://github.com/bearsh/hid), which vendors an up-to-date hidapi. (Does not show macOS `kIOMasterPortDefault` deprecation warnings.)
