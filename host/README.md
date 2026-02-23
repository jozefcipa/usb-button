# hid_listener

`hid_listener` is a host application for interacting with the USB button device (Raspberry Pi Pico). It can:

- Find the Pico USB button device.
- Listen for HID input reports (button presses) as they arrive (e.g., short, double, or long presses).
- Send HID output reports to the Pico (host → device).

## Usage

> **Note:** On Linux, you may need elevated permissions or custom udev rules.  
> On macOS, if open fails, add Terminal (or your preferred app) to **System Settings → Privacy & Security → Input Monitoring**.

### Common commands
(`hid_listener` binary can be found in `../.bin` directory after running `make build-host`)

```sh
./hid_listener                    # Listen for button reports
./hid_listener -list              # List available HID devices and exit
./hid_listener -send 0201         # Send Report ID 2, payload 0x01 (turn Pico LED on), then exit
./hid_listener -daemon            # Run listener in background (PID written to ~/.cache/hid_listener.pid)
./hid_listener stop               # Stop the background daemon (sends SIGTERM to PID from file)
```

## HID Library

Uses [github.com/bearsh/hid](https://github.com/bearsh/hid), which vendors an up-to-date hidapi. (Does not show macOS `kIOMasterPortDefault` deprecation warnings.)

## Handling Button Events with Lua

`hid_listener` supports event-based scripting with [Lua](https://www.lua.org/). You can attach your own custom actions for button events by creating a `hid_listener.lua` script in your home directory (example file [here](./hid_listener.example.lua)). 

When `hid_listener` starts, it looks for this file and loads it if present. The following Lua functions (if defined) are automatically called on their respective events:

- `function onSinglePress()`
- `function onDoublePress()`
- `function onLongPress()`

For example, to make a short press run a shell command:

```lua
function onSinglePress()
  os.execute("echo 'short press' >> /tmp/usb_button.log")
end
```
