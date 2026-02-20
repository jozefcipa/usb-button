# Protocol

This package defines the HID report values shared between firmware (Pico) and host, ensuring both sides use the same usage codes for button events.

**Consumer report** (report ID 3) carries one 16-bit usage. The constants in this package represent the values sent by the firmware and interpreted by the host.