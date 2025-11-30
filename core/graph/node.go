package graph

import (
	"errors"
	"log"

	"github.com/Y-Figos/ipesdk/core/adapters"
	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/Y-Figos/ipesdk/core/engine"
	lua "github.com/yuin/gopher-lua"
)

type ModuleStatus string

const (
	StatusPending ModuleStatus = "pending"
	StatusRunning ModuleStatus = "running"
	StatusSuccess ModuleStatus = "success"
	StatusFailed  ModuleStatus = "failed"
)

type OutputType string

type NodeModule struct {
	ModuleName string
	ScriptPath string
	Depends    []*NodeModule
	Children   []*NodeModule
	DataOutput string
	Adapter    string
	InArgs     map[string]any
	OutArgs     map[string]any
	Payloads    map[string]*df.Dataframe
	Status     ModuleStatus
}

func registerPayload(L *lua.LState, name string, dataframe *df.Dataframe) {
    ud := L.NewUserData()
    ud.Value = dataframe
    L.SetMetatable(ud, L.GetTypeMetatable("dataframe"))
    L.SetGlobal(name, ud)
}

func (nm *NodeModule) Run() ModuleStatus {
	nm.Status = StatusRunning

	L := lua.NewState()
	defer L.Close()
	engine.CreateLuaEnv(L)

	if nm.Adapter != "" {
		factory, ok := adapters.InputAdapterRegistry[nm.Adapter]
		if !ok {
			log.Printf("Adapter %v of %v do not exist", nm.Adapter, nm.ModuleName)
			return StatusFailed
		}
		adapter, err := factory(nm.InArgs)
		if err != nil {
			log.Printf("Adapter %v of %v: Error while creating adapter - %v ", nm.Adapter, nm.ModuleName, err)
			return StatusFailed
		}
		dataframe, err := adapter.GetData()
		if err != nil {
			log.Printf("Adapter %v of %v: Error while creating adapter - %v ", nm.Adapter, nm.ModuleName, err)
			return StatusFailed
		}
		nm.Payloads = map[string]*df.Dataframe{ //df.Dataframe is not a type
    	nm.ModuleName + "_input": dataframe,
		}
		registerPayload(L, nm.ModuleName + "_input", dataframe)
	}

	for _, dependency := range nm.Depends {
		if dependency.Status == StatusFailed {
			log.Printf("dependecy %v of %v did not succeed", dependency.ModuleName, nm.ModuleName)
			nm.Status = StatusFailed
			return StatusFailed
		}
		if dependency.Payloads != nil {
    	for name, dataframe := range dependency.Payloads {
        if nm.Payloads == nil {
            nm.Payloads = make(map[string]*df.Dataframe) //df.Dataframe is not a type
        }
        nm.Payloads[name] = dataframe
        registerPayload(L, name, dataframe)
    }
}
	}

	// Load the Lua script
	if err := L.DoFile(nm.ScriptPath); err != nil {
		log.Printf("[Module %s] Lua error: %v", nm.ModuleName, err)
		nm.Status = StatusFailed
		return StatusFailed
	}

	fn := L.GetGlobal("main")
	if fn.Type() != lua.LTFunction {
		log.Printf("main is not a function in module %s", nm.ModuleName)
		nm.Status = StatusFailed
		return StatusFailed
	}
	err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,    // expecting 1 return value
		Protect: true, // handle errors
	})

	if err != nil {
		log.Printf("Error while executing main of %s: %v", nm.ModuleName, err)
		nm.Status = StatusFailed
		return StatusFailed
	}

	ret := L.Get(-1)
	L.Pop(1)

	nm.Payloads = make(map[string]*df.Dataframe)

	switch val := ret.(type) {
	case *lua.LUserData:
		if df, ok := val.Value.(*df.Dataframe); ok {
			 nm.Payloads[nm.ModuleName + "payload"] = df
		} else {
			log.Printf("Returned value is not a Dataframe")
			nm.Status = StatusFailed
			return StatusFailed
		}

	case *lua.LTable:
		val.ForEach(func(key, value lua.LValue) {
			name := key.String()
			if ud, ok := value.(*lua.LUserData); ok {
				if df, ok := ud.Value.(*df.Dataframe); ok {
					nm.Payloads[name] = df
				}
			}
		})

default:
    log.Printf("main() did not return a dataframe or table")
    nm.Status = StatusFailed
    return StatusFailed
}
// depois do switch que preenche nm.Payloads

nm.Status = StatusSuccess

if nm.DataOutput != "" {
    if err := nm.Export(); err != nil {
        log.Printf("Error exporting module %s: %v", nm.ModuleName, err)
        nm.Status = StatusFailed
        return StatusFailed
    }
}

return StatusSuccess

}

func (nm *NodeModule) Export() error {
	if nm.DataOutput == "" {
		return errors.New("no Output was set")
	}
	factory, ok := adapters.OutAdapterRegistry[nm.DataOutput]
		if !ok {
			log.Printf("Adapter %v of %v do not exist", nm.Adapter, nm.ModuleName)
			return errors.New("output not valid")
	}
	dfToExport, ok := nm.Payloads["output"]
	if !ok {
		return errors.New("no 'output' payload to export")
	}
	outadapter, _ := factory(nm.OutArgs, dfToExport)
	outadapter.ExportData()
	return nil
}
