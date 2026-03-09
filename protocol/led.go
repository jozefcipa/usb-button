package protocol

// Host sends only one byte to the firmware: 0x01 to turn the LED on.
const LedOn byte = 0x01

const LedOff byte = 0x00

const LedBlinkOn byte = 0x02
