# Firmware (RPi Pico)

TinyGo firmware for a Raspberry Pi Pico. A physical button (short / double / long press) is reported to the host over USB HID as **consumer reports**.

## What it does

- **Button** (GPIO 14): Detects short, double, and long press; sends one HID consumer report per completed press.
- **LED** (GPIO 15): Blinks by default. When the host sends the “LED on” byte on the keyboard report (report ID 2), the LED stays on as commanded, signaling the host software is running.
- **USB HID**: Uses TinyGo’s CDC+HID composite. Consumer report ID 3 carries the 16‑bit press type; keyboard report ID 2 is used for host→device LED control.

## File structure

- **`main.go`** — Entry point: init, wait for USB, then a single loop that polls host reports, LED blink, and button; dispatches host bytes to `handleHostReport` and sends consumer reports on button events.
- **`internal/button`** — Button on GPIO 14, debounce and state machine; `Init()` and `Poll()` return the press type (short/double/long) when a press is complete.
- **`internal/led`** — LED on GPIO 15: `Init()`, `On()`, `Off()`, `BlinkOn()`, `ShouldBlink()` / `BlinkLED()` for the default blink and host-driven solid state.
- **`internal/hid`** — HID device wrapper: `Init()`, `SendConsumerReport()`, `Flush()`, `GetReceivedReport()` for host output (e.g. keyboard report with LED byte). Uses the shared `protocol` package for report IDs and constants.

## Flow

1. `main()` initializes button, LED (starts blinking), and HID, then waits for USB ready.
2. Main loop (every ~2 ms):
   - Read any host report via `GetReceivedReport()`; if keyboard report with LED byte, call `handleHostReport()` (switch LED to solid on/off or blink).
   - If `led.ShouldBlink()` is true, call `led.BlinkLED()` to toggle the LED.
   - Call `button.Poll()`; on completed press, send the corresponding consumer report and `Flush()`.
3. No goroutines; one non-blocking loop so USB and button are both serviced.

# View logs
To view the RPi logs, find the USB device (`ls /dev/cu.usb*`) and connect to the console via the `screen` command.

For instance,

```
screen /dev/cu.usbmodem2101 115200
```