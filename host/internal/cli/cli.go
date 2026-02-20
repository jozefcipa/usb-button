package cli

import "flag"

var RunAsDaemon bool
var StopDaemon bool
var ListHIDDevices bool
var SendHexData string

func DefineAndParseArgs() {
	runAsDaemon := flag.Bool("daemon", false, "run listener in background; PID is written to a file so 'hid_listener stop' can stop it")
	listOnly := flag.Bool("list", false, "list HID devices and exit")
	sendHexData := flag.String("send", "", "send one HID output report (hex bytes, no spaces) and exit; e.g. -send 0201 = report ID 2, payload 0x01 (LED on)")

	flag.Parse()

	StopDaemon = len(flag.Args()) > 0 && flag.Args()[0] == "stop"
	ListHIDDevices = *listOnly
	RunAsDaemon = *runAsDaemon
	SendHexData = *sendHexData
}
