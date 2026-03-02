package led

import (
	"machine"
	"time"
)

type LedState int

const (
	LedStateBlink LedState = iota
	LedStateOn
	LedStateOff
)

var lastBlinkToggle time.Time
var ledState LedState
var blinkLedOn bool

// Blink interval when waiting for host to signal "ready"
const blinkInterval = 500 * time.Millisecond
const led = machine.Pin(15)

func Init() {
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low() // LED off after startup
	blinkLedOn = false
	ledState = LedStateOff
}

func On() {
	ledState = LedStateOn
	led.High()
}

func Off() {
	ledState = LedStateOff
	led.Low()
}

// Enable blink mode
func BlinkOn() {
	lastBlinkToggle = time.Now()
	ledState = LedStateBlink
}

// Check if it's time to toggle the blink LED
func ShouldBlink() bool {
	if ledState != LedStateBlink {
		return false
	}
	if time.Since(lastBlinkToggle) >= blinkInterval {
		blinkLedOn = !blinkLedOn
		return true
	}
	return false
}

// Toggle the blink LED
func BlinkLED() {
	led.Set(blinkLedOn)
	lastBlinkToggle = time.Now()
}
