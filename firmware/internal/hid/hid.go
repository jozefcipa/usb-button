// This package sends raw HID report bytes over USB so you can fully control the report payload
//
// # USB HID protocol
//
// The host sees your device via descriptors:
//   - Device descriptor: VID (idVendor), PID (idProduct), device class, etc.
//   - Configuration descriptor: which interfaces exist (e.g. CDC + HID).
//   - HID interface: class 0x03, with a HID report descriptor that defines
//     the format of input/output reports (report ID, size, usage page).
//
// Data is sent in "reports": small packets. Each report has a report ID (if
// the descriptor uses multiple reports) and a payload. The report descriptor
// defines the layout (e.g. "report ID 2: 1 modifier byte + 6 key bytes").
// What you send must match what the descriptor declared; the host uses the
// same descriptor to parse the bytes.
//
// # Usage pages and report format
//
// The HID specification separates "what the data means" from "how many bytes".
// Both are described in the report descriptor and in the HID Usage Tables.
//
// Usage pages are numeric categories that tell the host what kind of data the
// report carries. They are defined in the HID Usage Tables (see link below).
// For example: 0x01 = Generic Desktop (mouse, keyboard, joystick), 0x07 =
// Keyboard (key codes, modifier bits), 0x0C = Consumer (media keys, system
// control, power). Within a page, "usages" identify specific controls (e.g.
// Consumer Control 0x01, Volume Increment 0xE9). The descriptor ties each
// field to a usage page and usage so the host can interpret values correctly.
//
// Report format is the concrete layout of each report: how many bits or bytes
// each field has (Report Size / Report Count), whether it is input or output,
// and the value range (Logical Minimum/Maximum). For instance, the standard
// keyboard report format is: 1 byte modifier (8 bits), 1 byte reserved, 6
// bytes key codes; consumer reports often use one 16-bit usage. The device
// and host must use the same format; the report descriptor is the contract.
//
// Reference: HID Usage Tables, https://usb.org/document-library/hid-usage-tables-14

package hid

import (
	"machine"
	"machine/usb/hid"

	"github.com/jozefcipa/usb-button/protocol"
)

// ConsumerReportSize is the total report length for report ID 3: ID + 2 bytes.
const ConsumerReportSize = 3

// MaxReceivedPayload is the maximum payload length we store for host→device reports.
const MaxReceivedPayload = 8

// Device implements the HID device interface so it can be registered with hid.SetHandler.
// It also stores the last output report received from the host (see GetReceivedReport).
type Device struct {
	// report is the next report to send (report ID + payload). Nil when idle.
	report []byte
	// host→device: last received output report (report ID + payload)
	rxReportID byte
	rxPayload  [MaxReceivedPayload]byte
	rxLen      uint8
	rxPending  bool
}

func Init() *Device {
	dev := &Device{}
	hid.SetHandler(dev)
	return dev
}

// TxHandler is called by the HID stack when the host is ready for input.
// On RP2040 the stack may not call this from the main loop; use Flush() after
// queueing a report to send it immediately from application code.
func (d *Device) TxHandler() bool {
	if len(d.report) == 0 {
		return false
	}

	hid.SendUSBPacket(d.report)
	d.report = nil

	return true
}

// RxHandler is called when the host sends an HID output report (e.g. keyboard LED report ID 2).
func (d *Device) RxHandler(b []byte) bool {
	if len(b) < 1 {
		return false
	}

	d.rxReportID = b[0]
	payloadLen := len(b) - 1
	if payloadLen > MaxReceivedPayload {
		payloadLen = MaxReceivedPayload
	}
	d.rxLen = uint8(payloadLen)
	copy(d.rxPayload[:], b[1:1+payloadLen])
	d.rxPending = true
	return true
}

// Returns the last output report received from the host, if any.
// The first byte of an HID output is the report ID (e.g. 0x02 for keyboard LED); the rest is payload.
// Call this from your main loop to process host→device data. Only one report is buffered; if the host
// sends multiple reports before you call GetReceivedReport, only the latest is kept.
func (d *Device) GetReceivedReport() (reportID byte, payload []byte, ok bool) {
	if !d.rxPending {
		return 0, nil, false
	}
	id := d.rxReportID
	payload = make([]byte, d.rxLen)
	copy(payload, d.rxPayload[:d.rxLen])
	d.rxPending = false
	return id, payload, true
}

// SendConsumerReport queues a consumer (report ID 3) report with the given
// 16-bit usage. The host may interpret it as a media key; for custom events
// use values like 0x0001, 0x0002, 0x0003 and handle them in your host app.
// Call Flush() after this to send the report immediately (required on RP2040
// where TxHandler is not invoked from the main loop).
func (d *Device) SendConsumerReport(usage uint16) {
	d.report = make([]byte, ConsumerReportSize)
	d.report[0] = protocol.HIDReportIDConsumer
	d.report[1] = byte(usage)      // usage_lo: low byte of 16-bit usage (little-endian)
	d.report[2] = byte(usage >> 8) // usage_hi: high byte
}

// Flush sends any queued report immediately via the HID IN endpoint.
// Call this after SendConsumerReport so the host receives the data
// on RP2040 the USB stack does not call TxHandler from the main loop.
func (d *Device) Flush() {
	if !machine.USBDev.InitEndpointComplete || len(d.report) == 0 {
		return
	}
	hid.SendUSBPacket(d.report)
	d.report = nil
}
