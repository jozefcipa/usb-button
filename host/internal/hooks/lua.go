package hooks

import (
	"fmt"
	"os"

	lua "github.com/yuin/gopher-lua"
)

var luaVM *lua.LState

func initLua() {
	if luaVM != nil {
		// Lua already initialized
		return
	}

	luaVM = lua.NewState()
}

func callLuaFn(fn lua.LValue) {
	if luaVM == nil || fn == nil || fn.Type() != lua.LTFunction {
		return
	}

	if err := luaVM.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Lua handler error: %v\n", err)
	}
}
