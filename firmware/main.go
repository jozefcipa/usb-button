package main

import (
	"machine"
	"time"

	"github.com/jozefcipa/usb-button/firmware/internal/button"
	"github.com/jozefcipa/usb-button/firmware/internal/hid"
	"github.com/jozefcipa/usb-button/firmware/internal/led"
	"github.com/jozefcipa/usb-button/protocol"
)

func main() {
	// Initialize peripherals
	button.Init()
	led.Init()
	led.BlinkOn()

	// Initialize USB HID device
	usbHid := hid.Init()

	waitForUSBHostReady()
	println("USB connected. Waiting for host to signal 'ready'...")

	// Single non-blocking loop: poll USB (host→Pico) and button (no goroutines on Pico)
	for {
		// Process any HID output reports from the host (e.g. LED command).
		if reportID, payload, ok := usbHid.GetReceivedReport(); ok {
			handleHostReport(reportID, payload)
		}

		// LED: blink by default until host sends LedCmdSolidOn, then keep LED as set by host.
		if led.ShouldBlink() {
			led.BlinkLED()
		}

		// Non-blocking button check (short/double/long).
		if pressType, ok := button.Poll(); ok {
			println("Button press detected:", protocol.BtnPressToHumanReadable(pressType))
			usbHid.SendConsumerReport(uint16(pressType))
			usbHid.Flush()
			time.Sleep(50 * time.Millisecond)
		}
		time.Sleep(2 * time.Millisecond) // poll interval; keeps USB responsive
	}
}

// handleHostReport processes an HID output report from the host. Report ID 2 = keyboard LED (1 byte).
// In blink mode only protocol.TurnLedOn (0x0001) is accepted: it switches to solid mode and turns LED on.
// In solid mode: 0 = LED off, non-zero = LED on.
func handleHostReport(reportID byte, payload []byte) {
	if reportID != protocol.HIDReportIDKeyboard || len(payload) < 1 {
		return
	}
	b := payload[0]
	if b == protocol.LedOn {
		led.On()
		println("LED turned on from host")
		return
	}
	if b == protocol.LedBlinkOn {
		led.BlinkOn()
		println("LED blink mode enabled from host")
		return
	}
}

// Wait for USB to be initialized (host has enumerated and set config)
func waitForUSBHostReady() {
	for {
		if machine.USBDev.InitEndpointComplete {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
