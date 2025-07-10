package engine

import (
	"github.com/yuin/gopher-lua"
	"github.com/Y-Figos/ipesdk/internal"
)

func CreateLuaEnv(L *lua.LState) {
	
	internal.RegisterReaderType(L)
	internal.RegisterDataFrameType(L)
	internal.RegisterColumnType(L)
	L.SetGlobal("new_reader", L.NewFunction(internal.RegisterCSVReader))

}