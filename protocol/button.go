package protocol

// Consumer usage values for button press types (report ID 3, 16-bit little-endian).
// Sent by firmware; host uses them to label reports and bind actions.

type ButtonPressType uint16

const (
	ShortPress  ButtonPressType = 0x0001
	DoublePress ButtonPressType = 0x0002
	LongPress   ButtonPressType = 0x0003
)
