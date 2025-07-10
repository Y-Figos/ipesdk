package engine

import (
	"github.com/yuin/gopher-lua"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/internal"
)

func CreateLuaEnv(L *lua.LState) {
	
	internal.RegisterReaderType(L)
	internal.RegisterDataFrameType(L)
	internal.RegisterColumnType(L)
	L.SetGlobal("new_reader", L.NewFunction(internal.RegisterCSVReader))

}