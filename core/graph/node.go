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
	LuaManager *engine.LuaManager
	ModuleName string
	ScriptPath string
	Depends    []*NodeModule
	Children   []*NodeModule
	DataOutput string
	Adapter    string
	InArgs     map[string]any
	OutArgs    map[string]any
	Payloads   map[string]*df.Dataframe
	Status     ModuleStatus
	ExportFlag bool
	Context    *RuntimeContext
}

func (nm *NodeModule) readBatchFiles(reader adapters.InputAdapterFactory) (*df.Dataframe, error) {
	batchFileReader := adapters.BatchFileReader{
		ReaderFactory: reader,
		DirPath: nm.InArgs["filepath"].(string),
	}
	newDf, err := batchFileReader.MergeFiles(nm.InArgs)
	if err != nil {
		return nil, err
	}
	return newDf, nil
}

func (nm *NodeModule) getInputFromAdapter() (*df.Dataframe, error) {
	if nm.Adapter == "" {
		return nil, nil
	}
	factory, ok := adapters.InputAdapterRegistry[nm.Adapter]
	if !ok {
		return nil, fmt.Errorf("adapter %v of %v do not exist", nm.Adapter, nm.ModuleName)
	}

	if nm.InArgs["batch_read"].(bool){
		dataframe, err := nm.readBatchFiles(factory)
		if err != nil {
			return nil, err
		}
		return dataframe, nil
	}

	adapter, err := factory(nm.InArgs)
	if err != nil {
		return nil, fmt.Errorf("adapter %v of %v: Error while creating adapter - %v ", nm.Adapter, nm.ModuleName, err)
	}
	dataframe, err := adapter.GetData()
	if err != nil {
		return nil, fmt.Errorf("adapter %v of %v: Error while creating adapter - %v ", nm.Adapter, nm.ModuleName, err)
	}
	return dataframe, nil
}

func (nm *NodeModule) registerAdapterInput() error {
	dataframe, err := nm.getInputFromAdapter()
	if err != nil {
		return fmt.Errorf("adapter %v of %v: Error while creating adapter - %v ", nm.Adapter, nm.ModuleName, err)
	}
	if dataframe == nil {
		return nil
	}
	if nm.Payloads == nil {
		nm.Payloads = make(map[string]*df.Dataframe)
	}
	nm.Payloads[nm.ModuleName+"_input"] = dataframe
	nm.LuaManager.RegisterPayload(nm.ModuleName+"_input", dataframe)
	return nil
}

func (nm *NodeModule) resolveDependencies() error {
	for _, dependency := range nm.Depends {
		if dependency.Status == StatusFailed {
			return fmt.Errorf("dependecy %v of %v did not succeed", dependency.ModuleName, nm.ModuleName)
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
	return nil
}

func (nm *NodeModule) resolveMainReturn(ret []lua.LValue) error {
	if len(ret) == 0 || ret[0] == lua.LNil {
		return nil
	}
	switch val := ret[0].(type) {
	case *lua.LUserData:
		if df, ok := val.Value.(*df.Dataframe); ok {
			nm.Payloads[nm.ModuleName+"payload"] = df
		} else {
			log.Printf("Returned value is not a Dataframe")
			return fmt.Errorf("returned value is not a Dataframe")
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
		return nil
	}
	return fmt.Errorf("main() did not return a dataframe or table: %v", nm.ModuleName)
}

func (nm *NodeModule) Run() ModuleStatus {
	nm.Status = StatusRunning
	nm.LuaManager = engine.NewLuaManager(nm.Context.Global)

	defer nm.LuaManager.L.Close()

	err := nm.registerAdapterInput()
	if err != nil {
		log.Printf("error registering payload input %s", err)
		return StatusFailed
	}

	err = nm.resolveDependencies()
	if err != nil {
		log.Printf("error registering payload input %s", err)
		return StatusFailed
	}

	nm.LuaManager.LoadScript(nm.ScriptPath)

	ret, err := nm.LuaManager.CallGlobalFunc("main", 1)
	if err != nil {
		log.Printf("error calling main function input %s", err)
		return StatusFailed
	}
	log.Println(ret[0].String())
	err = nm.resolveMainReturn(ret)
	if err != nil {
		log.Printf("returned invalid value: %s", err)
		return StatusFailed
	}

	if nm.ExportFlag {
		if err := nm.Export(); err != nil {
			log.Printf("Error while exporting of %s: %v", nm.ModuleName, err)
			return StatusFailed
		}
	}
	nm.Status = StatusSuccess
	return nm.Status
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
	//Export hook runs here, and pushs results to nm.outArgs
	if useHook, ok := nm.OutArgs["use_export_hook"].(bool); ok && useHook {
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

	ret, err := nm.LuaManager.CallGlobalFunc("export_hook", 1)
	if err != nil {
		return err
	}
	tbl, ok := ret[0].(*lua.LTable)
	if !ok {
		return fmt.Errorf("export_hook of %s did not return a table: %w", nm.ModuleName, err)
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
