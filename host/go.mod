module github.com/jozefcipa/usb-button/host

go 1.23

require (
	github.com/bearsh/hid v1.6.0
	github.com/jozefcipa/usb-button/protocol v0.0.0
)

replace github.com/jozefcipa/usb-button/protocol => ../protocol

require (
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/sys v0.30.0 // indirect
)
