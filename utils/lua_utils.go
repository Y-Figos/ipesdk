package utils

import(
	"github.com/yuin/gopher-lua"
	"fmt"
)

func ConvertAnytoLuaType(L *lua.LState,val any) lua.LValue{
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
    case []string:
		tbl := L.NewTable()
		for i, s := range v {
			tbl.RawSetInt(i+1, lua.LString(s)) // Lua is 1-indexed
		}
		return tbl
	case []any:
		tbl := L.NewTable()
		for i, item := range v {
			tbl.RawSetInt(i+1, ConvertAnytoLuaType(L, item))
		}
		return tbl
	case map[string]any:
		tbl := L.NewTable()
		for key, val := range v {
			tbl.RawSetString(key, ConvertAnytoLuaType(L, val)) // recursive call
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", v))
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