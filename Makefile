build:
	tinygo build -target=pico cmd/usb_button/main.go

flash:
	tinygo flash -target=pico cmd/usb_button/main.go

.PHONY: flash