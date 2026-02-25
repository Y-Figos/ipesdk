package internal

import (
	"github.com/Y-Figos/ipesdk/core/adapters"

	lua "github.com/yuin/gopher-lua"
)

func RegisterCSVReader(L *lua.LState) int {
	ud := L.NewUserData()
	ud.Value = adapters.NewCSVReaderWithOptions("", 0, 0)
	L.SetMetatable(ud, L.GetTypeMetatable("CSVReader"))
	L.Push(ud)
	return 1
}

func SetPath(L *lua.LState) int {
	reader := checkUserDataAs[*adapters.CSVReader](L, 1)
	newPath := L.CheckString(2)
	reader.FilePath = newPath
	return 0
}

func getPath(L *lua.LState) int {
	reader := checkUserDataAs[*adapters.CSVReader](L, 1)
	L.Push(lua.LString(reader.FilePath))
	return 1
}

func getData(L *lua.LState) int {
	reader := checkUserDataAs[*adapters.CSVReader](L, 1)
	df, err := reader.GetData()
	if err != nil {
		L.RaiseError("get_data erro: %v", err)
	}
	wrap(L, df, "dataframe")
	return 1
}

func RegisterReaderType(L *lua.LState) {
	mt := L.NewTypeMetatable("CSVReader")
	L.SetGlobal("CSVReader", mt)

	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"set_path": SetPath,
		"get_path": getPath,
		"get_data": getData,
	}))
}

