package graph

import (
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/engine"
	"github.com/yuin/gopher-lua"
	"log"
)

type ModuleStatus string

const (
	StatusPending ModuleStatus = "pending"
	StatusRunning ModuleStatus = "running"
	StatusSuccess ModuleStatus = "success"
	StatusFailed  ModuleStatus = "failed"
)

type OutputType string

const (
	OutputNil      OutputType = ""
	OutputNone      OutputType = "none"
	OutputDataframe OutputType = "dataframe"
	OutputCSV       OutputType = "csv"
	OutputJSON      OutputType = "json"
)

type NodeModule struct{
	ModuleName string
	ScriptPath string 
	Depends []*NodeModule
	Children []*NodeModule
 	DataOutput OutputType 
	Args []string
	Payload *df.Dataframe
	Status ModuleStatus
}

func (nm *NodeModule) registerDeps(L *lua.LState, dependecy *NodeModule ) {
	ud := L.NewUserData()
	ud.Value = dependecy.Payload
	L.SetMetatable(ud, L.GetTypeMetatable("dataframe"))
	L.SetGlobal("payload_" + dependecy.ModuleName, ud)
}

func (nm *NodeModule) Run() ModuleStatus {
	nm.Status = StatusRunning
	
	L := lua.NewState()
	defer L.Close()
	engine.CreateLuaEnv(L)
	for _, dependency := range nm.Depends {
		if dependency.Status == StatusFailed {
			log.Printf("dependecy %v of %v did not succeed",dependency.ModuleName, nm.ModuleName)
			nm.Status = StatusFailed
			return StatusFailed
		}
		if dependency.Payload != nil {
			nm.registerDeps(L, dependency)
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
	NRet:    1,     // expecting 1 return value
	Protect: true,  // handle errors
	})

	if err != nil {
		log.Printf("Error while executing main of %s: %v", nm.ModuleName, err)
		nm.Status = StatusFailed
		return StatusFailed
	}

	ret := L.Get(-1)
	L.Pop(1)   

	ud, ok := ret.(*lua.LUserData)
	if !ok {
		log.Printf("main() did not return a dataframe")
		nm.Status = StatusFailed
		return StatusFailed
	}

	newdf, ok := ud.Value.(*df.Dataframe)
	if !ok {
		log.Printf("Returned value is not a Dataframe")
		nm.Status = StatusFailed
		return StatusFailed
	}
	nm.DataOutput = OutputDataframe
	nm.Payload = newdf
	nm.Status = StatusSuccess
	return StatusSuccess
}
