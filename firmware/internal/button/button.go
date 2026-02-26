package button

import (
	"machine"
	"time"

	"github.com/jozefcipa/usb-button/protocol"
)

const button = machine.Pin(14)

const (
	LONG_PRESS_MS   = 800
	DOUBLE_PRESS_MS = 400
	DEBOUNCE_MS     = 30
)

type pollState uint8

const (
	stateIdle pollState = iota
	stateDebouncePress
	statePressed
	stateReleasedWaitDouble
	stateDebounceSecond
	stateSecondPressed
	stateSecondReleased
)

var (
	ps          pollState = stateIdle
	psAt        time.Time
	pressStart  time.Time
	releaseTime time.Time
)

func Init() {
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
}

// Poll runs one tick of the button state machine. Call it repeatedly from your
// main loop so you can also process USB (e.g. GetReceivedReport) without blocking.
// Returns (pressType, true) when a complete short/double/long press is detected, otherwise (0, false).
func Poll() (protocol.ButtonPressType, bool) {
	now := time.Now()
	low := !button.Get() // pressed = LOW

	switch ps {
	case stateIdle:
		if low {
			psAt = now
			ps = stateDebouncePress
		}
		return 0, false

	case stateDebouncePress:
		if !low {
			ps = stateIdle
			return 0, false
		}
		if time.Since(psAt) >= time.Duration(DEBOUNCE_MS)*time.Millisecond {
			pressStart = now
			ps = statePressed
		}
		return 0, false

	case statePressed:
		if low {
			return 0, false
		}
		// Released
		if time.Since(pressStart) >= time.Duration(LONG_PRESS_MS)*time.Millisecond {
			ps = stateIdle
			return protocol.LongPress, true
		}
		releaseTime = now
		ps = stateReleasedWaitDouble
		return 0, false

	case stateReleasedWaitDouble:
		if time.Since(releaseTime) > time.Duration(DOUBLE_PRESS_MS)*time.Millisecond {
			ps = stateIdle
			return protocol.ShortPress, true
		}
		if low {
			psAt = now
			ps = stateDebounceSecond
		}
		return 0, false

	case stateDebounceSecond:
		if !low {
			ps = stateReleasedWaitDouble
			return 0, false
		}
		if time.Since(psAt) >= time.Duration(DEBOUNCE_MS)*time.Millisecond {
			ps = stateSecondPressed
		}
		return 0, false

	case stateSecondPressed:
		if low {
			return 0, false
		}
		psAt = now
		ps = stateSecondReleased
		return 0, false

	case stateSecondReleased:
		if time.Since(psAt) < time.Duration(DEBOUNCE_MS)*time.Millisecond {
			return 0, false
		}
		ps = stateIdle
		return protocol.DoublePress, true
	}

	return 0, false
}
