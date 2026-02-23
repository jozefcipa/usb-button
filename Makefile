build-fw:
	tinygo build -target=pico -o ./.bin/usb_button.uf2 ./firmware/main.go

build-host:
	go build -o ./.bin/hid_listener ./host/main.go

flash:
	tinygo flash -target=pico ./firmware/main.go

.PHONY: build build-fw flash
