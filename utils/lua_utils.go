package utils

import(
	"github.com/yuin/gopher-lua"
	"fmt"
)

func ConvertAnytoLuaType(val any) lua.LValue{
	var luaVal lua.LValue
    switch v := val.(type) {
    case string:
        luaVal = lua.LString(v)
    case int:
        luaVal = lua.LNumber(v)
    case float64:
        luaVal = lua.LNumber(v)
    case bool:
        luaVal = lua.LBool(v)
    default:
        luaVal = lua.LString(fmt.Sprintf("%v", v)) // fallback to string
    }

	return luaVal
}

func ConvertLuaTypeToGoType(val lua.LValue) any {
	switch v := val.(type) {
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	case *lua.LTable:
		goMap := make(map[any]any)
		v.ForEach(func(key, value lua.LValue) {
			goMap[key.String()] = ConvertLuaTypeToGoType(value)
		})
		return goMap
	case *lua.LFunction:
		return "function" 
	case *lua.LUserData:
		return v.Value 
	default:
		return nil
	}
}