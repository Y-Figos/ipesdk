package graph

import (
	"errors"
	"fmt"
	"log"
	"github.com/Y-Figos/ipesdk/utils"
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
	LuaManager	*engine.LuaManager
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
	ExportFlag 	bool
	Context *RuntimeContext
}


func (nm *NodeModule) Run() ModuleStatus {
	nm.Status = StatusRunning
	nm.LuaManager = engine.NewLuaManager()
	nm.LuaManager.L = lua.NewState()
	defer func() {
		if !nm.ExportFlag {
			nm.LuaManager.L.Close()
		}
	}()
	nm.LuaManager.CreateLuaEnv()
	nm.LuaManager.InjectContextTable(nm.Context.Global)
	
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
		nm.LuaManager.RegisterPayload(nm.ModuleName + "_input", dataframe)
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
			nm.LuaManager.RegisterPayload(name, dataframe)
    	}
	}
	}

	nm.LuaManager.LoadScript(nm.ScriptPath)
	ret, err := nm.LuaManager.CallGlobalFunc("main", 1)
	if err != nil{
		return StatusFailed
	}
	
	nm.Payloads = make(map[string]*df.Dataframe)

	switch val := ret[0].(type) {
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
	
	if nm.ExportFlag{
		if err := nm.Export(); err != nil{
			log.Printf("Error while exporting of %s: %v", nm.ModuleName, err)
			return StatusFailed
		}
	}
	
	nm.Status = StatusSuccess
	return nm.Status
}

func (nm *NodeModule) Export() error {
	defer nm.LuaManager.L.Close()
	if nm.DataOutput == "" {
		return errors.New("no Output was set")
	}
	factory, ok := adapters.OutAdapterRegistry[nm.DataOutput]
		if !ok {
			log.Printf("Adapter %v of %v do not exist", nm.Adapter, nm.ModuleName)
			return errors.New("output not valid")
	}
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
	if nm.Context.Global[nm.ModuleName+".export"], err = outadapter.ExportData(); err != nil {
		return fmt.Errorf("failed to export data: %w", err)
	} 
	
	return nil
}

func (nm *NodeModule) export_hook() error {

	ret,err := nm.LuaManager.CallGlobalFunc("export_hook", 1)
	if err != nil{
		return err
	}
	tbl, ok := ret[0].(*lua.LTable)
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

