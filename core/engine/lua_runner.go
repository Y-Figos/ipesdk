package engine

import (
	"log"
	"github.com/yuin/gopher-lua"
)

func greetings(L *lua.LState) int {
	name := L.ToString(1) 
	L.Push(lua.LString("Hello, " + name)) // Whats is this push function? and why cant i just push a normal string?
	return 1 
}

func Teste(path string) {
	L := lua.NewState()
	defer L.Close()

	L.SetGlobal("greetings", L.NewFunction(greetings)) // How exactly the greetings function getr the LState?
	// Load the Lua script
	if err := L.DoFile(path); err != nil {
        log.Fatal(err)
    }
}