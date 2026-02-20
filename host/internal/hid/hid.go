// Package hid provides HID device connection and report I/O for the USB button host app.
// It uses the consumer interface (report ID 3) to match the firmware. For the HID
// report format and usage codes, see the project docs and the protocol package.
//
// USB HID references:
//   - HID specification: https://www.usb.org/hid
//   - HID Usage Tables (usage pages and usages): https://usb.org/document-library/hid-usage-tables-14
package hid

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bearsh/hid"
	"github.com/jozefcipa/usb-button/protocol"
)

// HID usage page and usage values (from HID Usage Tables). One physical USB device
// can expose multiple HID interfaces (e.g. keyboard, mouse, consumer). We prefer the
// consumer interface because our firmware sends report ID 3 (consumer control).
//
// Reference: https://usb.org/document-library/hid-usage-tables-14

// TODO: This is still a bit unclear - figure out what exactly is going on here.
const (
	usagePageGenericDesktop = 0x01 // Usage Page: Generic Desktop (mouse, keyboard, etc.)
	usagePageConsumer       = 0x0C // Usage Page: Consumer (media keys, system control)
	usageMouse              = 0x02 // Generic Desktop: Mouse
	usageKeyboard           = 0x06 // Generic Desktop: Keyboard
	usageConsumerCtrl       = 0x01 // Consumer: Consumer Control
)

func Connect(vid, pid uint16) (*hid.Device, error) {
	devices := hid.Enumerate(vid, pid)
	if len(devices) == 0 {
		return nil, fmt.Errorf("no HID device found with VID=0x%04X PID=0x%04X (try -list to see devices)", vid, pid)
	}
	// One physical device can expose multiple "logical" HID entries. Try opening each in turn:
	// prefer consumer (report ID 3), then any other. On macOS some interfaces may be claimed
	// by the system (e.g. keyboard) and fail with "failed to open device"; another may work.
	var tryOrder []int
	for i := range devices {
		if devices[i].UsagePage == usagePageConsumer && devices[i].Usage == usageConsumerCtrl {
			tryOrder = append(tryOrder, i)
			break
		}
	}
	for i := range devices {
		if len(tryOrder) > 0 && tryOrder[0] == i {
			continue
		}
		tryOrder = append(tryOrder, i)
	}
	var lastErr error
	for _, i := range tryOrder {
		dev, err := devices[i].Open()
		if err == nil {
			return dev, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("could not open any HID interface (tried %d): %w", len(devices), lastErr)
}

func ListDevices() {
	all := hid.Enumerate(0, 0)
	if len(all) == 0 {
		fmt.Println("No HID devices found.")
		return
	}

	fmt.Printf("HID devices (VID:PID usage_page:usage path product):\n")
	fmt.Printf("  Same physical device can appear multiple times (one per HID usage pair, e.g. keyboard/mouse/consumer).\n")
	for i, d := range all {
		usage := usageLabel(d.UsagePage, d.Usage)
		fmt.Printf("  %d: 0x%04X:0x%04X 0x%02X:0x%02X %s  %q %q\n", i, d.VendorID, d.ProductID, d.UsagePage, d.Usage, usage, d.Path, d.Product)
	}
}

// usageLabel returns a short label for the given HID usage page and usage (for ListDevices output).
func usageLabel(usagePage, usage uint16) string {
	switch usagePage {
	case usagePageGenericDesktop:
		switch usage {
		case usageMouse:
			return "(mouse)"
		case usageKeyboard:
			return "(keyboard)"
		}
	case usagePageConsumer:
		if usage == usageConsumerCtrl {
			return "(consumer)"
		}
	}
	return ""
}

// SendData opens the device, sends the given hex bytes as one HID output report, then closes.
// Example: 0201 sends [0x02, 0x01] (report ID 2, payload 0x01 → Pico LED on).
func SendData(vid, pid uint16, hexStr string) error {
	hexStr = strings.TrimSpace(strings.ReplaceAll(hexStr, " ", ""))
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		log.Fatalf("Invalid hex data: %v (use e.g. 0201 for report ID 2, payload 0x01)", err)
	}
	if len(data) == 0 {
		log.Fatal("Need at least one byte (report ID)")
	}

	dev, err := Connect(vid, pid)
	if err != nil {
		return err
	}
	defer dev.Close()

	n, err := dev.Write(data)
	if err != nil {
		log.Fatalf("Failed to write data: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Sent %d bytes: % X\n", n, data[:n])

	return nil
}

func ListenForEvents(dev *hid.Device) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	reports := make(chan []byte, 8)
	go func() {
		buf := make([]byte, 64)
		for {
			n, err := dev.Read(buf)
			if err != nil {
				log.Printf("read: %v", err)
				close(reports)
				return
			}
			if n > 0 {
				cp := make([]byte, n)
				copy(cp, buf[:n])
				reports <- cp
			}
		}
	}()

	for {
		select {
		case <-quit:
			return
		case b, ok := <-reports:
			if !ok {
				return
			}
			printReport(b)
		}
	}
}

func printReport(b []byte) {
	// Our firmware sends consumer reports: [report_id=0x03, usage_lo, usage_hi]
	if len(b) >= 3 && b[0] == 0x03 {
		// Reconstruct 16-bit usage from little-endian bytes: low byte first, high byte second
		// Example: double press = 0x0002 → b[1]=0x02, b[2]=0x00 → 0x02 | 0x00 = 0x0002.
		usage := uint16(b[1]) | uint16(b[2])<<8
		label := btnPressToHumanReadable(protocol.ButtonPressType(usage))
		if label == "" {
			label = fmt.Sprintf("usage=0x%04X", usage)
		}
		fmt.Printf("report: %s [0x%02X, 0x%02X, 0x%02X]\n", label, b[0], b[1], b[2])
		return
	}
	// Raw bytes for any other report
	fmt.Printf("report (%d bytes): % X\n", len(b), b[:len(b)])
}

func btnPressToHumanReadable(usage protocol.ButtonPressType) string {
	switch usage {
	case protocol.ShortPress:
		return "SHORT_PRESS"
	case protocol.DoublePress:
		return "DOUBLE_PRESS"
	case protocol.LongPress:
		return "LONG_PRESS"
	default:
		return fmt.Sprintf("unknown usage=0x%04X", usage)
	}
}
