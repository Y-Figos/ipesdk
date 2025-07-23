package graph

import (
	"errors"
	"fmt"
	"log"

	"github.com/Y-Figos/ipesdk/core/adapters"
	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/Y-Figos/ipesdk/core/engine"
	"github.com/Y-Figos/ipesdk/utils"
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
	ModuleName 	string
	ScriptPath 	string
	Depends     []*NodeModule
	Children    []*NodeModule
	DataOutput 	string
	Adapter    	string
	InArgs     	map[string]any
	OutArgs     map[string]any
	Payloads    map[string]*df.Dataframe
	Status     	ModuleStatus
	L			*lua.LState
	ExportFlag 	bool
}

func registerPayload(L *lua.LState, name string, dataframe *df.Dataframe) {
    ud := L.NewUserData()
    ud.Value = dataframe
    L.SetMetatable(ud, L.GetTypeMetatable("dataframe"))
    L.SetGlobal(name, ud)
}

func (nm *NodeModule) Run() ModuleStatus {
	nm.Status = StatusRunning

	nm.L = lua.NewState()
	defer func() {
		if !nm.ExportFlag {
			nm.L.Close()
		}
	}()
	engine.CreateLuaEnv(nm.L)

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
		registerPayload(nm.L, nm.ModuleName + "_input", dataframe)
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
            nm.Payloads = make(map[string]*df.Dataframe)
        }
        nm.Payloads[name] = dataframe
        registerPayload(nm.L, name, dataframe)
    }
}
	}

	// Load the Lua script
	if err := nm.L.DoFile(nm.ScriptPath); err != nil {
		log.Printf("[Module %s] Lua error: %v", nm.ModuleName, err)
		nm.Status = StatusFailed
		return StatusFailed
	}

	fn := nm.L.GetGlobal("main")
	if fn.Type() != lua.LTFunction {
		log.Printf("main is not a function in module %s", nm.ModuleName)
		nm.Status = StatusFailed
		return StatusFailed
	}
	err := nm.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,    // expecting 1 return value
		Protect: true, // handle errors
	})

	if err != nil {
		log.Printf("Error while executing main of %s: %v", nm.ModuleName, err)
		nm.Status = StatusFailed
		return StatusFailed
	}

	ret := nm.L.Get(-1)
	nm.L.Pop(1)

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
	nm.Status = StatusSuccess
	return StatusSuccess
}

func (nm *NodeModule) Export() error {
	defer nm.L.Close()
	if nm.DataOutput == "" {
		return errors.New("no Output was set")
	}
	factory, ok := adapters.OutAdapterRegistry[nm.DataOutput]
		if !ok {
			log.Printf("Adapter %v of %v do not exist", nm.Adapter, nm.ModuleName)
			return errors.New("output not valid")
	}
	// dfToExport, ok := nm.Payloads["output"]
	// if !ok {
	// 	return errors.New("no 'output' payload to export")
	// }
	
	//Export hook runs here, and pushs results to nm.outArgs
	if nm.OutArgs["use_export_hook"].(bool){
		err := nm.export_hook()
		if err != nil {
			return err
		}
	}
	outadapter, err := factory(nm.OutArgs, nm.Payloads)
	if err != nil {
		return fmt.Errorf("failed to create output adapter: %w", err)
	}
	if outadapter == nil {
		return errors.New("output adapter is nil")
	}
	if err := outadapter.ExportData(); err != nil {
		return fmt.Errorf("failed to export data: %w", err)
	}
	
	return nil
}

func (nm *NodeModule) export_hook() error {

	fn := nm.L.GetGlobal("export_hook")
	if fn.Type() != lua.LTFunction {
		return fmt.Errorf("export_hook is not a function in module %s", nm.ModuleName)
	}

	err := nm.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,    // expecting 1 return value
		Protect: true, // handle errors
	})

	if err != nil {
		return fmt.Errorf("error while executing export_hook of %s: %v", nm.ModuleName, err)
	}

	ret := nm.L.Get(-1)
	nm.L.Pop(1)
	tbl, ok := ret.(*lua.LTable)
	if !ok{
		return  fmt.Errorf("export_hook of %s did not return a table: %w", nm.ModuleName, err)
	}
	if nm.OutArgs == nil {
    nm.OutArgs = make(map[string]interface{})
	}
	tbl.ForEach(func(key, value lua.LValue) {
			name := key.String()
			convertedValue := utils.ConvertLuaTypeToGoType(value)
			nm.OutArgs[name] = convertedValue
	})
	
	return nil
}