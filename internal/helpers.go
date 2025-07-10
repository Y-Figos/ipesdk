package internal

import(
	"fmt"
	"github.com/yuin/gopher-lua"
)

func wrap(L *lua.LState, value any, metatableName string) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = value
	L.SetMetatable(ud, L.GetTypeMetatable(metatableName))
	L.Push(ud)
	return ud
}

func checkUserDataAs[T any](L *lua.LState, index int) T {
	ud := L.CheckUserData(index)
	val, ok := ud.Value.(T)
	if !ok {
		L.ArgError(index, fmt.Sprintf("expected %T", *new(T)))
	}
	return val
}