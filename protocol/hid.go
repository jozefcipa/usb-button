// TinyGo's default descriptor (CDC+HID) for the Pico declares these report IDs.
// The HID standard defines usage pages and report formats, not the numeric IDs;
// the assignment 1=mouse, 2=keyboard, 3=consumer is TinyGo's choice in that descriptor.
//
//   - Report ID 1: mouse (buttons + X, Y, wheel).
//   - Report ID 2: keyboard (modifier + 6 keys).
//   - Report ID 3: consumer (one 16-bit usage, e.g. volume).
package protocol

// KeyboardReportID is defined as Report ID 2 in the TinyGo HID descriptor.
// https://github.com/tinygo-org/tinygo/blob/db9f1182f5f2a64ea496752899626578d2b313a7/src/machine/usb/descriptor/hid.go#L138
// Payload format:
// SEND (fw -> host): 1 modifier byte + 1 reserved + 6 key bytes = 8 bytes after the ID.
// RECEIVE (host -> fw): 1 byte (LED state) after the report ID.
const HIDReportIDKeyboard = 0x02

// ConsumerReportID is defined as Report ID 3 in the TinyGo HID descriptor.
// https://github.com/tinygo-org/tinygo/blob/db9f1182f5f2a64ea496752899626578d2b313a7/src/machine/usb/descriptor/hid.go#L205
// Payload format:
// SEND (fw -> host): One 16-bit usage value; good for custom events (short/double/long = 1/2/3).
// RECEIVE (host -> fw): No receive reports defined for this report ID
const HIDReportIDConsumer = 0x03
