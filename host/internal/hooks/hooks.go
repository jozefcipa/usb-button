package hooks

import (
	"fmt"
	"os"
	"path"

	hidHelper "github.com/bearsh/hid"
	"github.com/jozefcipa/usb-button/host/internal/hid"
	"github.com/jozefcipa/usb-button/protocol"
	lua "github.com/yuin/gopher-lua"
)

const luaScriptFile = "hid_listener.lua"

// SendLEDCmd sends one byte (LED command) to the device. Used by Lua helpers led_on, led_off, led_blink.
type SendLEDCmd func(cmd byte) error

var (
	onSinglePress lua.LValue
	onDoublePress lua.LValue
	onLongPress   lua.LValue
	sendLEDCmd    SendLEDCmd
)

func Configure(dev *hidHelper.Device) {
	sendLEDCmd = func(cmd byte) error {
		return hid.SendData(dev, []byte{protocol.HIDReportIDKeyboard, cmd})
	}
	initLua()

	// Expose LED control to Lua so scripts can turn the Pico LED on, off, or blink.
	registerLEDHelpers()

	// Load the user script from ~/hid_listener.lua
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lua: failed to get user home directory: %v\n", err)
		return
	}

	// Interpret the user script
	luaScriptPath := path.Join(homeDir, luaScriptFile)
	if err := luaVM.DoFile(luaScriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "Lua: failed to load %s: %v\n", luaScriptPath, err)
		luaVM = nil // disables Lua for rest of session
		return
	}

	// Get the function handlers from the user script
	onSinglePress = luaVM.GetGlobal("onSinglePress")
	onDoublePress = luaVM.GetGlobal("onDoublePress")
	onLongPress = luaVM.GetGlobal("onLongPress")
}

// registerLEDHelpers registers led_on(), led_off(), led_blink() in the Lua VM.
func registerLEDHelpers() {
	if luaVM == nil || sendLEDCmd == nil {
		return
	}
	luaVM.SetGlobal("led_on", luaVM.NewFunction(func(L *lua.LState) int {
		if sendLEDCmd != nil {
			_ = sendLEDCmd(protocol.LedOn)
		}
		return 0
	}))
	luaVM.SetGlobal("led_off", luaVM.NewFunction(func(L *lua.LState) int {
		if sendLEDCmd != nil {
			_ = sendLEDCmd(protocol.LedOff)
		}
		return 0
	}))
	luaVM.SetGlobal("led_blink", luaVM.NewFunction(func(L *lua.LState) int {
		if sendLEDCmd != nil {
			_ = sendLEDCmd(protocol.LedBlinkOn)
		}
		return 0
	}))
}

func HandleHIDEvent(pressType protocol.ButtonPressType) {
	if pressType == protocol.ShortPress && onSinglePress != nil {
		callLuaFn(onSinglePress)
	} else if pressType == protocol.DoublePress && onDoublePress != nil {
		callLuaFn(onDoublePress)
	} else if pressType == protocol.LongPress && onLongPress != nil {
		callLuaFn(onLongPress)
	}
}
