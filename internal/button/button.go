package button

import (
	"machine"
	"time"
)

// --- Configuration ---
const (
	buttonPinNum    = 14
	buttonLedPinNum = 15
	longPressMs     = 800
	doublePressMs   = 400
	debounceMs      = 30
)

type PressType int

const (
	SHORT_PRESS PressType = iota
	LONG_PRESS
	DOUBLE_PRESS
)

var (
	button    = machine.Pin(buttonPinNum)
	buttonLed = machine.Pin(buttonLedPinNum)
)

func Init() {
	buttonLed.Configure(machine.PinConfig{Mode: machine.PinOutput})
	buttonLed.High()
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
}

// waitForButtonState waits until the button is in expectedState (true = HIGH, false = LOW),
// or until timeout expires. Returns true if the state was reached, false on timeout.
func waitForButtonState(expectedHigh bool, timeout time.Duration) bool {
	start := time.Now()

	for {
		current := button.Get() // true = HIGH, false = LOW[web:27]
		if current == expectedHigh {
			time.Sleep(time.Millisecond * debounceMs) // debounce
			return true
		}
		if timeout > 0 && time.Since(start) > timeout {
			return false
		}
		time.Sleep(time.Millisecond) // 1 ms poll
	}
}

// blocks until at least one completed press is detected.
func WaitForPress() PressType {
	// wait until LOW (pressed)
	waitForButtonState(false, 0)

	// Measure how long it stays pressed
	pressStart := time.Now()
	for !button.Get() { // while LOW
		if time.Since(pressStart) > time.Millisecond*longPressMs {
			// Long press detected; wait for release then return
			waitForButtonState(true, 0) // wait until HIGH (released)
			return LONG_PRESS
		}
		time.Sleep(time.Millisecond)
	}

	// Released before longPressMs -> short press
	firstReleaseTime := time.Now()

	// 2) Look for a second press within doublePressMs
	for time.Since(firstReleaseTime) < time.Millisecond*doublePressMs {
		if !button.Get() { // second press started (LOW)
			// Debounce and ensure it is actually pressed
			waitForButtonState(false, 0)
			// Wait for it to be released again
			waitForButtonState(true, 0)
			return DOUBLE_PRESS
		}
		time.Sleep(time.Millisecond)
	}

	// No second press in time -> single short press
	return SHORT_PRESS
}
