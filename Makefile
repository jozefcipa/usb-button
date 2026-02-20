build-fw:
	tinygo build -target=pico -o ./.bin/usb_button.uf2 ./firmware/cmd/usb_button

build-host:
	go build -o ./.bin/hid_listener ./host/cmd/hid_listener

flash:
	tinygo flash -target=pico ./firmware/cmd/usb_button

.PHONY: build build-fw flash
