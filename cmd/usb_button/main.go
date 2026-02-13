package main

import (
	"machine"
	"time"

	"github.com/jozefcipa/usb-button/internal/button"
)

var led = machine.Pin(machine.LED)

func main() {
	// Configure button and LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	button.Init()

	// Infinite loop
	for {
		pressType := button.WaitForPress()
		handlePress(pressType)
	}
}

func handlePress(pressType button.PressType) {
	switch pressType {
	case button.SHORT_PRESS:
		// Single press: blink LED once
		led.High()
		time.Sleep(150 * time.Millisecond)
		led.Low()
		println("Single press")
	case button.DOUBLE_PRESS:
		// Double press: blink LED twice
		for i := 0; i < 2; i++ {
			led.High()
			time.Sleep(150 * time.Millisecond)
			led.Low()
			time.Sleep(150 * time.Millisecond)
		}
		println("Double press")
	case button.LONG_PRESS:
		// Long press: keep LED on longer
		led.High()
		time.Sleep(600 * time.Millisecond)
		led.Low()
		println("Long press")
	}
}
