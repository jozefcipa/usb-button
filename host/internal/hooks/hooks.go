package hooks

import (
	"fmt"
	"os"
	"path"

	"github.com/jozefcipa/usb-button/protocol"
	lua "github.com/yuin/gopher-lua"
)

const luaScriptFile = "hid_listener.lua"

var (
	onSinglePress lua.LValue
	onDoublePress lua.LValue
	onLongPress   lua.LValue
)

func Configure() {
	initLua()

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

func HandleHIDEvent(pressType protocol.ButtonPressType) {
	if pressType == protocol.ShortPress && onSinglePress != nil {
		callLuaFn(onSinglePress)
	} else if pressType == protocol.DoublePress && onDoublePress != nil {
		callLuaFn(onDoublePress)
	} else if pressType == protocol.LongPress && onLongPress != nil {
		callLuaFn(onLongPress)
	}
}
