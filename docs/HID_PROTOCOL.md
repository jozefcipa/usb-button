# HID protocol

This doc explains the HID protocol in order: enumeration (host gets descriptors, sets configuration), how the “agreement” on report format works (the report descriptor, no separate handshake), and what our code does (register HID device, then send custom events as report bytes).

## 1. Enumeration: how the host learns about the device

When you plug in the Pico, the host (your computer) runs **enumeration**:

1. It sends a control request for the **device descriptor**.  
   The device responds with VID, PID, device class, and other basic info.

2. It requests the **configuration descriptor** (and any interface/endpoint descriptors bundled with it).  
   That describes which interfaces exist (e.g. CDC for serial, HID for input).

3. For the HID interface, it requests the **HID report descriptor**.  
   The device responds with a description of the **report format**: report IDs, usage pages, and how many bytes each report has. This is the “agreement” – both sides use this format from then on. There is no separate “handshake” message; the report descriptor *is* the contract.

All of the above is handled by TinyGo’s USB stack and machine package. Our application code does not send these descriptor bytes ourselves; the stack responds to the host’s requests using the descriptors compiled into the firmware.

## 2. After enumeration: sending data (reports)

Once the configuration is set, the HID interface has an **interrupt IN** endpoint. The device uses it to send **input reports** to the host.

- We don’t send a special “ready” or “config” byte first. As soon as the host has enumerated the device and we have something to send, we can send a report.
- **Format is fixed by the report descriptor.**  
  Each packet we send must match what we declared: for example “Report ID 3, then 2 bytes (one 16‑bit value).” The host parses incoming reports using that same descriptor.

So the flow is:

1. Host enumerates device (gets descriptors, sets configuration).
2. Device and host both “agree” on the report layout via the report descriptor (no separate handshake).
3. Device starts sending **custom events as HID reports** – each report is a small block of bytes (report ID + data) that matches the descriptor.

## 3. What our code does

We don’t send descriptors manually. We:

- Register our raw HID device with TinyGo’s HID layer. The first handler registered also triggers USB HID setup: the stack will use the built‑in descriptor (CDC + HID with report IDs 1, 2, 3) and respond to the host’s descriptor requests during enumeration. We don’t send a “configure” byte; the host gets the descriptors and sets the configuration as in section 1.

- Build one HID input report that matches the **consumer** report in the descriptor:  
  `[report ID 3, usage_lo, usage_hi]`.

- Use usage values like `0x0001`, `0x0002`, `0x0003` for short / double / long. That report is queued in our device.

- Call `Flush()` so the data is sent on the interrupt IN endpoint (on RP2040 the stack does not call `TxHandler` from the main loop, so we send from application code).

When the host polls, it receives our report and interprets it using the same report descriptor. So: **we send custom events with data through HID** by sending these small reports; the “data” is the 16‑bit usage (and we could use keyboard or other report IDs for different layouts).
