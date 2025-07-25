package engine

import (
	"fmt"
	"path/filepath"
	"github.com/Y-Figos/ipesdk/internal"
	"github.com/Y-Figos/ipesdk/utils"
	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/yuin/gopher-lua"
)

type LuaManager struct{
	L *lua.LState
}
func NewLuaManager() *LuaManager {
	L := lua.NewState()
	lm := &LuaManager{L: L}
	lm.CreateLuaEnv()
	return lm 
}
func (lm *LuaManager) CreateLuaEnv() {
	
	internal.RegisterReaderType(lm.L)
	internal.RegisterDataFrameType(lm.L)
	internal.RegisterColumnType(lm.L)
	lm.L.SetGlobal("new_reader", lm.L.NewFunction(internal.RegisterCSVReader))

}

func (lm *LuaManager) InjectContextTable(globalContext map[string]any) {
	contextTable := lm.L.NewTable()

	for key, value := range globalContext {
		lm.L.SetField(contextTable, key, utils.ConvertAnytoLuaType(lm.L,value)) // your existing converter
	}

	lm.L.SetGlobal("context", contextTable)
}

func (lm *LuaManager) LoadScript(scriptPath string) error {
	fn, err := lm.L.LoadFile(scriptPath)
	if err != nil{
		return fmt.Errorf("error loading script %v: %w",filepath.Base(scriptPath), err)
	}
	lm.L.Push(fn)
	err = lm.L.PCall(0,lua.MultRet, nil)
	if err != nil{
		return fmt.Errorf("error passing script to Lua VM %v: %w",filepath.Base(scriptPath), err)
	}
	return nil
}

func (lm *LuaManager) CallGlobalFunc(name string, nrets int) ([]lua.LValue, error){
	fn := lm.L.GetGlobal(name)
	if fn.Type() != lua.LTFunction{
		return nil, fmt.Errorf("%v is not a function", name)
	}
	err := lm.L.CallByParam(lua.P{
		Fn: fn,
		NRet: nrets,
		Protect: true,
	})
	if err != nil {
		return nil, err
	}
	results := make([]lua.LValue, nrets)
	for i := nrets - 1; i >= 0; i-- {
		results[i] = lm.L.Get(-1)
		lm.L.Pop(1)
	}
	return results, nil

}

func (lm *LuaManager) RegisterPayload(name string, dataframe *df.Dataframe) {
    ud := lm.L.NewUserData()
    ud.Value = dataframe
    lm.L.SetMetatable(ud, lm.L.GetTypeMetatable("dataframe"))
    lm.L.SetGlobal(name, ud)
}
