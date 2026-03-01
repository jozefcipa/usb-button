# Custom USB descriptor (VID/PID and HID report format)

The firmware uses **TinyGo’s built-in USB descriptors**. They are **not** in this repo; they live in the TinyGo compiler’s source. To use a custom vendor/product ID or a custom HID report format (e.g. different report sizes, or an output report for consumer), you have to **edit TinyGo’s source** and build the compiler (or use a local copy of the descriptor).

## What “default TinyGo descriptor” means

When you build with `tinygo build -target=pico ...`, TinyGo links in a **CDC + HID** composite device:

- **Device descriptor** – VID, PID, etc. (e.g. Raspberry Pi VID `0x2E8A`, product `0x000A` for Pico).
- **Configuration** – CDC (serial) + one HID interface with **one** HID report descriptor that describes all three report IDs.
- **HID report descriptor** – A single blob that defines:
  - **Report ID 1** – Mouse: input only (buttons + X, Y, wheel).
  - **Report ID 2** – Keyboard: input (modifier + 6 keys) and **output** (1 byte, LEDs).
  - **Report ID 3** – Consumer: **input only** (one 16‑bit usage). No output is declared, so the host cannot send report ID 3 to the device with the default descriptor.

That descriptor is generated from Go code in TinyGo’s tree. Your application code (e.g. `internal/hid`) only sends and receives bytes; it does **not** define the descriptor. The host and device both rely on the descriptor as the “contract” for report IDs and sizes.

## Where to find it in the code (TinyGo repo)

The descriptor is built from:

| What | Where in TinyGo |
|------|-------------------|
| HID report descriptor (mouse, keyboard, consumer) | `src/machine/usb/descriptor/hid.go` |
| Helpers (HIDReportID, HIDInput, HIDOutput, etc.) | `src/machine/usb/descriptor/hidreport.go` |
| Device descriptor (VID/PID, etc.) | `src/machine/usb/descriptor/device.go` |
| Configuration / interfaces / endpoints | `src/machine/usb/descriptor/configuration.go`, `interface.go`, `endpoint.go`, etc. |

Clone TinyGo and open those files:

```bash
git clone https://github.com/tinygo-org/tinygo.git
cd tinygo
# e.g. release branch
git checkout release
```

In `src/machine/usb/descriptor/hid.go`, the `CDCHID` descriptor has a field `HID: map[uint16][]byte`. The key `2` is the **interface number**; the value is the **HID report descriptor** bytes (built from the `Append([][]byte{...})` list). There you’ll see:

- `HIDReportID(1)` + mouse items (Input only).
- `HIDReportID(2)` + keyboard items: Input (modifiers, keys) and **Output** (LEDs, 1 byte).
- `HIDReportID(3)` + consumer: one 16‑bit **Input** only; no `HIDOutput` for report 3.

So “default TinyGo descriptor” = that exact report descriptor (and the rest of the USB descriptors) coming from this TinyGo source.
