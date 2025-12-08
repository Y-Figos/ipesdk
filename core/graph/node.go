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

// ============================================================================
// Leitura de batch (vários arquivos)
// ============================================================================

func (nm *NodeModule) readBatchFiles(reader adapters.InputAdapterFactory) (*df.Dataframe, error) {
	dir, ok := nm.InArgs["filepath"].(string)
	if !ok || dir == "" {
		return nil, fmt.Errorf("readBatchFiles: 'filepath' must be a non-empty string")
	}

	batchFileReader := adapters.BatchFileReader{
		ReaderFactory: reader,
		DirPath:       dir,
	}

	newDf, err := batchFileReader.MergeFiles(nm.InArgs)
	if err != nil {
		return nil, err
	}

	if len(batchFileReader.FailedFiles) != 0 {
		log.Printf("readBatchFiles: some files failed during read: %#v", batchFileReader.FailedFiles)

		failPath := []string{}
		failReason := []string{}
		for k, v := range batchFileReader.FailedFiles {
			failPath = append(failPath, k)
			failReason = append(failReason, v)
		}

		columnLog := df.NewColumn("Files", failPath)
		columnReason := df.NewColumn("Reason", failReason)
		dflog := df.Dataframe{
			Columns: make(map[string]df.ColumnInterface),
		}
		dflog.NewColumn("Fails", columnLog)
		dflog.NewColumn("Reason", columnReason)

		if nm.Payloads == nil {
			nm.Payloads = make(map[string]*df.Dataframe)
		}
		nm.Payloads["Faillog_"+nm.ModuleName] = &dflog
	}

	return newDf, nil
}

// ============================================================================
// Input Adapter
// ============================================================================

func (nm *NodeModule) getInputFromAdapter() (*df.Dataframe, error) {
	// Nó sem adapter de entrada (ex: só recebe deps)
	if nm.Adapter == "" {
		return nil, nil
	}

	factory, ok := adapters.InputAdapterRegistry[nm.Adapter]
	if !ok {
		return nil, fmt.Errorf("adapter %v of %v do not exist", nm.Adapter, nm.ModuleName)
	}

	// batch_read = true  → lê pasta
	if flag, ok := nm.InArgs["batch_read"].(bool); ok && flag {
		dataframe, err := nm.readBatchFiles(factory)
		if err != nil {
			return nil, err
		}
		return dataframe, nil
	}

	// modo normal: arquivo único
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

	// Nó sem input de arquivo (ex: só depende de outros nós)
	if dataframe == nil {
		return nil
	}

	if nm.Payloads == nil {
		nm.Payloads = make(map[string]*df.Dataframe)
	}

	varName := nm.ModuleName + "_input"
	nm.Payloads[varName] = dataframe

	// expõe no Lua como, por exemplo, Module_1_input
	nm.LuaManager.RegisterPayload(varName, dataframe)

	return nil
}

// ============================================================================
// Dependências (payloads herdados de outros módulos)
// ============================================================================

func (nm *NodeModule) resolveDependencies() error {
	for _, dependency := range nm.Depends {

		if dependency.Status == StatusFailed {
			log.Printf("dependency %v of %v failed", dependency.ModuleName, nm.ModuleName)
			nm.Status = StatusFailed
			return fmt.Errorf("dependency %s failed", dependency.ModuleName)
		}

		if dependency.Payloads != nil {

			log.Println("------------------------------------------------")
			log.Printf("[DEBUG] %s receiving payloads from the parent module %s:", nm.ModuleName, dependency.ModuleName)

			for name, dataframe := range dependency.Payloads {

				if nm.Payloads == nil {
					nm.Payloads = make(map[string]*df.Dataframe)
				}

				log.Printf("    -> %s (rows=%d)", name, dataframe.RowCount())
				nm.Payloads[name] = dataframe

				// expõe no Lua (variáveis globais)
				engine.RegisterPayload(nm.LuaManager.L, name, dataframe)
			}
			log.Println("------------------------------------------------")
		}
	}

	return nil
}

// ============================================================================
// Tratamento do retorno do main() em Lua
// ============================================================================

func (nm *NodeModule) resolveMainReturn(ret []lua.LValue) error {
	if len(ret) == 0 || ret[0] == lua.LNil {
		// main() não retornou nada → ok
		return nil
	}

	if nm.Payloads == nil {
		nm.Payloads = make(map[string]*df.Dataframe)
	}

	switch val := ret[0].(type) {

	case *lua.LUserData:
		// main() retorna diretamente um dataframe
		if dfPtr, ok := val.Value.(*df.Dataframe); ok {
			nm.Payloads[nm.ModuleName+"_payload"] = dfPtr
			return nil
		}
		return fmt.Errorf("returned LUserData is not *df.Dataframe")

	case *lua.LTable:
		// main() retorna tabela: { nome = dataframe, ... }
		val.ForEach(func(key, value lua.LValue) {
			name := key.String()
			if ud, ok := value.(*lua.LUserData); ok {
				if dfPtr, ok := ud.Value.(*df.Dataframe); ok {
					nm.Payloads[name] = dfPtr
					log.Printf("[DEBUG resolveMainReturn] %s payload '%s' rows=%d",
                    nm.ModuleName, name, dfPtr.RowCount())
				}
			}
		})
		return nil
	}

	return fmt.Errorf("main() did not return a dataframe or table: %v", nm.ModuleName)
}

// ============================================================================
// Execução do módulo
// ============================================================================

func (nm *NodeModule) Run() ModuleStatus {
    nm.Status = StatusRunning

    // cria VM Lua com contexto global compartilhado
    nm.LuaManager = engine.NewLuaManager(nm.Context.Global)
    defer nm.LuaManager.L.Close()

    // 1) Lê adapter de entrada (se tiver) e expõe no Lua
    if nm.InArgs != nil {
        if _, hasPath := nm.InArgs["filepath"]; hasPath {
            if err := nm.registerAdapterInput(); err != nil {
                log.Printf("error registering payload input %s", err)
                nm.Status = StatusFailed
                return nm.Status
            }
        }
    }

    // 2) Puxa payloads das dependências e registra no Lua
    if err := nm.resolveDependencies(); err != nil {
        log.Printf("error resolving dependencies %s", err)
        nm.Status = StatusFailed
        return nm.Status
    }

    // 3) Carrega script Lua do módulo
    if err := nm.LuaManager.LoadScript(nm.ScriptPath); err != nil {
        log.Printf("error loading lua script for %s: %s", nm.ModuleName, err)
        nm.Status = StatusFailed
        return nm.Status
    }

    // 4) Chama main()
    ret, err := nm.LuaManager.CallGlobalFunc("main", 1)
    if err != nil {
        log.Printf("error calling main function: %s", err)
        nm.Status = StatusFailed
        return nm.Status
    }

    if len(ret) > 0 {
        log.Println(ret[0].String())
    }

    // 5) Interpreta retorno do main e guarda em Payloads
    if err := nm.resolveMainReturn(ret); err != nil {
        log.Printf("returned invalid value: %s", err)
        nm.Status = StatusFailed
        return nm.Status
    }

    // 6) Exporta, se for nó de saída
    if nm.ExportFlag {
        if err := nm.Export(); err != nil {
            log.Printf("Error while exporting of %s: %v", nm.ModuleName, err)
            nm.Status = StatusFailed
            return nm.Status
        }
    }

    nm.Status = StatusSuccess
    return nm.Status
}

// ============================================================================
// Export
// ============================================================================

func (nm *NodeModule) Export() error {
	if nm.DataOutput == "" {
		return errors.New("no Output was set")
	}

	factory, ok := adapters.OutAdapterRegistry[nm.DataOutput]
	if !ok {
		log.Printf("Adapter %v of %v do not exist", nm.DataOutput, nm.ModuleName)
		return errors.New("output not valid")
	}

	// export_hook opcional (pré-processa OutArgs via Lua)
	if useHook, ok := nm.OutArgs["use_export_hook"].(bool); ok && useHook {
		if err := nm.export_hook(); err != nil {
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

	exportResult, err := outadapter.ExportData()
	if err != nil {
		return fmt.Errorf("failed to export data: %w", err)
	}

	if nm.Context != nil && nm.Context.Global != nil {
		nm.Context.Global[nm.ModuleName+".export"] = exportResult
	}

	return nil
}

// ============================================================================
// export_hook
// ============================================================================

func (nm *NodeModule) export_hook() error {
	ret, err := nm.LuaManager.CallGlobalFunc("export_hook", 1)
	if err != nil {
		return err
	}

	tbl, ok := ret[0].(*lua.LTable)
	if !ok {
		return fmt.Errorf("export_hook of %s did not return a table", nm.ModuleName)
	}

	if nm.OutArgs == nil {
		nm.OutArgs = make(map[string]any)
	}

	tbl.ForEach(func(key, value lua.LValue) {
		name := key.String()
		convertedValue := utils.ConvertLuaTypeToGoType(value)
		nm.OutArgs[name] = convertedValue
	})

	return nil
}
