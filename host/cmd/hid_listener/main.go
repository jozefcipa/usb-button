package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/jozefcipa/usb-button/host/internal/cli"
	"github.com/jozefcipa/usb-button/host/internal/daemon"
	"github.com/jozefcipa/usb-button/host/internal/hid"
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

	if cli.ListHIDDevices {
		hid.ListDevices()
		return
	}

	if cli.SendHexData != "" {
		if err := hid.SendData(RPI_PICO_VID, RPI_PICO_PID, cli.SendHexData); err != nil {
			log.Fatalf("Failed to send data: %v", err)
		}
		return
	}

	// Connect to the device
	dev, err := hid.Connect(RPI_PICO_VID, RPI_PICO_PID)
	if err != nil {
		if runtime.GOOS == "darwin" {
			printMacOSHIDHelp()
		}
		log.Fatalf("Failed to connect to the device: %v", err)
	}
	defer dev.Close()

	// Listen for events
	fmt.Fprintf(os.Stderr, "Listening for reports (VID=0x%04X PID=0x%04X). Press Ctrl+C to stop.\n", RPI_PICO_VID, RPI_PICO_PID)
	hid.ListenForEvents(dev)

	// TODO: Act upon events, add some logic here, or Lua scripting for custom logic?
}

func printMacOSHIDHelp() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "macOS HID open failed. Try in order:")
	fmt.Fprintln(os.Stderr, "  1. System Settings → Privacy & Security → Input Monitoring")
	fmt.Fprintln(os.Stderr, "     → Add Terminal (or Cursor / your IDE), enable it, then quit and reopen Terminal.")
	fmt.Fprintln(os.Stderr, "  2. If it still fails, run with sudo (macOS sometimes requires root for HID):")
	fmt.Fprintln(os.Stderr, "     sudo ./hid_listener")
	fmt.Fprintln(os.Stderr, "  3. Unplug the Pico, plug it back in, then run again.")
	fmt.Fprintln(os.Stderr, "")
}
