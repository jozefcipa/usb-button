# USB, HID, and this project — learning summary

This document summarizes step by step what this project does and how it fits into USB and HID: from basic concepts to the exact flow in our code.

---

## 1. What is USB?

**USB (Universal Serial Bus)** is a standard for connecting devices to a host (usually a computer). It defines:

- **Physical layer**: cables, connectors, electrical signalling.
- **Protocol layer**: how the host and device exchange packets (control, bulk, interrupt, isochronous).
- **Software model**: the host is in charge. It discovers the device by **enumeration**, learns what the device is via **descriptors**, then chooses a **configuration** and talks over **endpoints** (pipes for data).

When you plug in a device, the host asks “who are you?” and “what can you do?”. The device answers with descriptors (binary blobs). No custom driver is needed if the device declares a **class** the OS already supports (e.g. HID, CDC). Our Pico declares itself as a composite device (CDC + HID), so the host can use it as a serial port and as an HID device.

**References:** [USB.org](https://www.usb.org/), USB specification.

---

## 2. What is HID?

**HID (Human Interface Device)** is a **USB device class** (class code 0x03). It was designed for keyboards, mice, joysticks, and similar “human input” devices, but it is widely used for any small, periodic data exchange where you want **no custom driver**: the OS has a generic HID driver that reads a **report descriptor** and then knows how to parse **reports** (small packets).

Important ideas:

- **Report descriptor** — A blob the device sends during enumeration that describes the **format** of its reports: how many bytes, what they mean (usage page/usage), input vs output. It is the “contract” between device and host; there is no separate handshake.
- **Reports** — The actual data. Each report is a small packet (often a few bytes). For **input reports**, the device sends to the host (e.g. key pressed); for **output reports**, the host sends to the device (e.g. keyboard LEDs). Reports are sent over **interrupt endpoints** (polled by the host at a fixed interval).
- **Usage pages and usages** — Defined in the [HID Usage Tables](https://usb.org/document-library/hid-usage-tables-14). They tell the host *what* the data represents (e.g. “Consumer” page, “Consumer Control” usage, one 16‑bit value). The report descriptor ties each field to a usage page/usage and to a size (Report Size/Report Count) and range (Logical Min/Max).

So: **HID = “device class that describes its reports via a descriptor, then sends/receives those reports.”** Our firmware sends **consumer** reports (report ID 3, one 16‑bit usage) to signal button events; the host app reads those reports and can send **keyboard output** (report ID 2, 1 byte) to control the LED.

### 2.1 Custom usage values and avoiding collisions

The HID Usage Tables assign specific meanings to many usage values. On the **Consumer** usage page (0x0C), examples include:

| Usage (hex) | Assigned meaning        |
|-------------|-------------------------|
| 0x01        | Consumer Control        |
| 0x02        | Numeric Key Pad         |
| 0x03        | Programmable Buttons    |
| 0xCD        | Play/Pause              |
| 0xE9        | Volume Increment        |
| 0xEA        | Volume Decrement        |

In this project we use consumer usages **0x0001**, **0x0002**, **0x0003** for short / double / long press. Those values **collide** with the standard assignments above (e.g. 0x0001 = Consumer Control, 0xE9/0xEA = volume). For a custom device this is often acceptable (we ignore standard semantics), but if you want to avoid collisions:

- **Check the Usage Tables** — Pick a usage that is unassigned or in a reserved range for your usage page.
- **Use a reserved/vendor range** — The spec reserves ranges for vendor use; use values from such a range for your custom semantics.
- **Use the Vendor usage page (0xFF00)** — Define your own semantics on the Vendor page so they do not conflict with standard Consumer/Keyboard usages.

**References:** [USB HID specification](https://www.usb.org/hid), [HID Usage Tables](https://usb.org/document-library/hid-usage-tables-14).

---

## 3. Other options besides HID

USB offers several ways for a device to talk to the host:

| Option | Typical use | Driver / complexity |
|--------|-------------|----------------------|
| **HID** | Keyboards, mice, gamepads, custom “button-like” devices | Built-in in OS; no custom driver. Reports are small and format is described by the report descriptor. |
| **CDC (Communications Device Class)** | Serial port (UART over USB) | Often appears as a COM port / tty; host uses serial APIs. TinyGo uses it for `println` and serial I/O. |
| **Vendor-specific class** | Custom protocols | Host usually needs a custom driver or libusb-style access. |
| **Mass storage (MSC)** | USB stick / disk | Built-in; device exposes a filesystem. |
| **Other classes** | Audio, video, etc. | Depends on the class. |

This project uses **HID** so that:

- No custom driver is required on the host.
- We can send small, well-defined “events” (button press type) and receive small commands (LED on/off) using the existing HID stack and a simple host app that opens the device by VID/PID and reads/writes report bytes.

We also have **CDC** in the same device (TinyGo’s default “CDC + HID” composite) for serial debug (e.g. `println` on the Pico).

---

## 4. How HID works at a high level

### 4.1 Enumeration

When the Pico is plugged in:

1. **Host** requests the **device descriptor** → gets VID, PID, device class, etc.
2. **Host** requests the **configuration descriptor** (and interface/endpoint descriptors) → learns there is e.g. one CDC interface and one HID interface.
3. **Host** requests the **HID report descriptor** for the HID interface → gets the exact format of all reports (report IDs, usage pages, sizes). This is the “agreement”; no separate handshake.
4. **Host** sets the **configuration** → device is ready. The HID interface now has an **interrupt IN** endpoint (device → host) and usually an **interrupt OUT** endpoint (host → device) if the descriptor declared output reports.

All of this is handled by the USB stack (on the Pico, TinyGo’s machine/USB code). Our application code does not send descriptors manually.

### 4.2 Sending and receiving data

- **Device → host (input reports):** The device places report bytes (e.g. `[0x03, lo, hi]`) on the HID interrupt IN endpoint. The host **polls** that endpoint periodically; when it polls, it gets the latest report (or nothing if the device had nothing to send). Our firmware builds the report and sends it (e.g. via `SendConsumerReport` + `Flush()`).
- **Host → device (output reports):** The host sends report bytes on the HID interrupt OUT endpoint. The device receives them; in our code, TinyGo’s HID layer calls our `RxHandler` with the received bytes, and we store the last one and expose it via `GetReceivedReport()`.

There is no “connection” step in the application: once the host has set the configuration, we just send/receive report bytes that match the descriptor.

---

## 5. What TinyGo does with USB and HID for us

TinyGo provides the **USB stack** and **HID integration** for the RP2040 (Pico):

- **Descriptors:** TinyGo compiles a **default composite descriptor** (CDC + HID) into the firmware. It lives in TinyGo’s source, not in our repo: [TinyGo descriptor (hid.go)](https://github.com/tinygo-org/tinygo/blob/release/src/machine/usb/descriptor/hid.go). That descriptor defines:
  - **Report ID 1** — mouse (buttons + X, Y, wheel).
  - **Report ID 2** — keyboard (modifier + 6 keys, plus 1-byte output for LEDs).
  - **Report ID 3** — consumer (one 16‑bit usage, input only in the default).
  The **numeric** report IDs (1, 2, 3) are TinyGo’s choice; the HID standard defines usage pages and report *formats*, not which ID is used for which type.

- **Endpoints and handling:** TinyGo sets up the USB endpoints (including HID interrupt IN/OUT) and, when we register an HID handler with `hid.SetHandler(dev)`, uses our `TxHandler` / `RxHandler` for sending and receiving. On RP2040, `TxHandler` is not invoked from the main loop by the stack, so we **flush** reports explicitly from application code (`Flush()`).

- **No custom driver:** Because the device declares the standard HID class and a standard-looking report descriptor (keyboard, mouse, consumer), the host OS loads its generic HID driver. Our host app just opens the device by VID/PID and reads/writes report bytes.

So: **TinyGo gives us the descriptors, the endpoints, and the hook to feed in report bytes; we only build the right bytes and call Flush (and handle received reports in the main loop).**

---

## 6. How our code uses HID to communicate with the host

### 6.1 Shared protocol (usage values)

- The **protocol** package defines the 16‑bit **consumer usage** values that mean “short press”, “double press”, “long press” (e.g. `0x0001`, `0x0002`, `0x0003`). Firmware and host both use these so labels and behavior stay in sync.

### 6.2 Firmware (Pico) — device side

1. **Setup**
   - Create a raw HID device: `dev := rawhid.New()` and register it: `hid.SetHandler(dev)`. This triggers TinyGo’s HID setup and the built-in report descriptor (CDC + HID with report IDs 1, 2, 3).
   - Initialize the button (GPIO 14/15, poll-based state machine).
   - Wait until USB is ready: `machine.USBDev.InitEndpointComplete`.

2. **Main loop (non-blocking)**
   - **Host → device:** Call `dev.GetReceivedReport()`. If the host sent an output report (e.g. keyboard report ID 2 with 1 byte), we get it and act on it (e.g. LED on/off). Only the last report is buffered.
   - **Button → report:** Call `button.Poll()`. If a complete press is detected (short/double/long), map it to a usage from the **protocol** package, call `dev.SendConsumerReport(usage)`, then `dev.Flush()` so the 3-byte report `[0x03, usage_lo, usage_hi]` is sent on the HID IN endpoint.
   - Small sleep (e.g. 2 ms) so the loop doesn’t spin and USB stays responsive.

3. **Report format**
   - We send **consumer** reports only: first byte = report ID `0x03`, next two bytes = 16‑bit usage (little-endian). The host receives these and interprets them using the same descriptor (and our protocol labels).

### 6.3 Host (computer) — host side

1. **Finding the device**
   - Enumerate HID devices by VID/PID (e.g. Raspberry Pi Pico default `0x2E8A` / `0x000A`). One physical device can expose several “logical” HID interfaces (keyboard, mouse, consumer). We **prefer the consumer** interface (usage page 0x0C, usage 0x01) because our firmware sends report ID 3; we open that interface (or fall back to another if opening fails, e.g. if the OS claimed the keyboard).

2. **Receiving button events (device → host)**
   - Read from the opened HID device in a loop. Each read returns a report (e.g. 3 bytes: `[0x03, lo, hi]`). We treat `b[0] == 0x03` as a consumer report, reconstruct the 16‑bit usage as `uint16(b[1]) | uint16(b[2])<<8`, and map it to a label (e.g. SHORT_PRESS) using the **protocol** package.

3. **Sending commands (host → device)**
   - To control the LED, the host sends an **output** report: e.g. `[0x02, 0x01]` (report ID 2, payload 0x01 = LED on). The firmware receives it in `RxHandler`, stores it, and the main loop reads it via `GetReceivedReport()` and sets the LED.

### 6.4 End-to-end flow (one button press)

| Step | Where | What happens |
|------|--------|-------------------------------|
| 1 | Host | Has already enumerated the Pico and opened the consumer (or other) HID interface; is reading in a loop. |
| 2 | Firmware | User presses button; `button.Poll()` returns e.g. SHORT_PRESS. |
| 3 | Firmware | `sendRawReport(dev, SHORT_PRESS)` → `SendConsumerReport(protocol.ShortPress)` → report `[0x03, 0x01, 0x00]` queued. |
| 4 | Firmware | `dev.Flush()` → bytes sent on HID interrupt IN endpoint. |
| 5 | Host | Next read returns `[0x03, 0x01, 0x00]`; usage = 0x0001; prints e.g. “report: SHORT_PRESS [0x03, 0x01, 0x00]”. |

No custom driver, no handshake beyond the descriptor: we just send and receive report bytes that match the TinyGo default HID report descriptor.

---

## 7. References (from the project)

- [USB HID specification](https://www.usb.org/hid)
- [HID Usage Tables](https://usb.org/document-library/hid-usage-tables-14)
- [TinyGo USB HID descriptor (hid.go)](https://github.com/tinygo-org/tinygo/blob/release/src/machine/usb/descriptor/hid.go)
