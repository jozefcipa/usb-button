package protocol

import "fmt"

// Consumer usage values for button press types (report ID 3, 16-bit little-endian).
// Sent by firmware; host uses them to label reports and bind actions.

type ButtonPressType uint16

const (
	ShortPress  ButtonPressType = 0x0001
	DoublePress ButtonPressType = 0x0002
	LongPress   ButtonPressType = 0x0003
)

func BtnPressToHumanReadable(pressType ButtonPressType) string {
	switch pressType {
	case ShortPress:
		return "SHORT_PRESS"
	case DoublePress:
		return "DOUBLE_PRESS"
	case LongPress:
		return "LONG_PRESS"
	default:
		return fmt.Sprintf("unknown usage=0x%04X", pressType)
	}
}
