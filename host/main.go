package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jozefcipa/usb-button/host/internal/cli"
	"github.com/jozefcipa/usb-button/host/internal/daemon"
	"github.com/jozefcipa/usb-button/host/internal/hid"
	"github.com/jozefcipa/usb-button/host/internal/hooks"
	"github.com/jozefcipa/usb-button/protocol"
)

// Default VID/PID for Raspberry Pi Pico (TinyGo default)
const (
	RPI_PICO_VID uint16 = 0x2E8A
	RPI_PICO_PID uint16 = 0x000A
)

func main() {
	// Define and parse CLI arguments
	cli.DefineAndParseArgs()

	// Handle CLI arguments first
	if cli.RunAsDaemon {
		daemon.Start()
		return
	}

	if cli.StopDaemon {
		daemon.Stop()
		return
	}

	// Debug command to list the available HID devices
	if cli.ListHIDDevices {
		hid.ListDevices()
		return
	}

	// Connect to RPi Pico
	rpiPico, err := hid.Connect(RPI_PICO_VID, RPI_PICO_PID)
	if err != nil {
		log.Fatalf("Failed to connect to the device: %v", err)
	}
	defer rpiPico.Close()

	// Debug command to send hex data directly to the device
	if cli.SendHexData != "" {
		hexStr := strings.TrimSpace(strings.ReplaceAll(cli.SendHexData, " ", ""))
		data, err := hex.DecodeString(hexStr)
		if err != nil {
			log.Fatalf("Failed to decode hex data: %v", err)
		}
		if err := hid.SendData(rpiPico, data); err != nil {
			log.Fatalf("Failed to send data: %v", err)
		}
		return
	}

	// Set up hooks for handling HID events
	hooks.Configure()

	// Send a "ready" report to the firmware
	if err := hid.SendData(rpiPico, []byte{
		// TinyGo doesn't define HIDReportConsumer as bidirectional,
		// so we use HIDReportIDKeyboard that accepts one byte of data from the host
		protocol.HIDReportIDKeyboard,
		protocol.LedOn,
	}); err != nil {
		log.Fatalf("Failed to send Ready report: %v", err)
	}

	// Listen for HID reports
	fmt.Fprintf(os.Stderr, "Listening for reports (VID=0x%04X PID=0x%04X). Press Ctrl+C to stop.\n", RPI_PICO_VID, RPI_PICO_PID)
	hidReports := hid.ListenForHIDReports(rpiPico)

	// Handle Ctrl+C and process interrupts to quit the program
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Inifite loop to handle events
	for {
		select {
		case <-quit:
			return
		case hidReport, ok := <-hidReports:
			if !ok {
				fmt.Fprintln(os.Stderr, "Error: HID reports channel closed")
				os.Exit(1)
			}

			pressType, err := hid.ValidateHIDReport(hidReport)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid HID report: %v\n", err)
				continue
			}

			hooks.HandleHIDEvent(pressType)
		}
	}
}
