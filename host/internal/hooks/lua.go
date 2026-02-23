package hooks

import (
	"fmt"
	"os"
	"os/exec"

	lua "github.com/yuin/gopher-lua"
)

var luaVM *lua.LState

func initLua() {
	if luaVM != nil {
		// Lua already initialized
		return
	}

	luaVM = lua.NewState()

	// Provide os_exec(cmd) to run a shell command and return its output/error
	luaVM.SetGlobal("os_exec", luaVM.NewFunction(func(L *lua.LState) int {
		// get the command string from the Lua script
		cmdStr := L.CheckString(1)

		// execute the command
		out, err := exec.Command("sh", "-c", cmdStr).CombinedOutput()

		// push the output to the Lua stack
		L.Push(lua.LString(string(out)))

		// push the error status to the Lua stack
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}

		return 2 // return 2 values to the Lua script
	}))
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
