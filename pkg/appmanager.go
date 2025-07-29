package pkg

import (
	"context"
	"github.com/Y-Figos/ipesdk/core/file_handler"
	"github.com/Y-Figos/ipesdk/core/graph"
)

type IPEAppManager struct{
	FolderStruct	*file_handler.FolderStruct
	ctx				*context.Context
}

func NewIPEAppManager(ctx context.Context) *IPEAppManager {
	return &IPEAppManager{
		FolderStruct: &file_handler.FolderStruct{},
		ctx: &ctx,
	}
}

func (ipe *IPEAppManager) Startup() error {
	err := ipe.FolderStruct.Init()
	if err != nil{
		return err
	}
	return nil
}

func (ipe *IPEAppManager) GetToolList() (*file_handler.ToolRegistry,error) {
	toolreg, err := ipe.FolderStruct.ParseToolRegistry()
	if err != nil{
		return nil, err
	}
	return toolreg, nil
}

func (ipe *IPEAppManager) InstallTool(ipePath string) error {
	if err := ipe.FolderStruct.Extract(ipePath); err != nil{
		return err
	}	
	return nil
}

func (ipe *IPEAppManager) RunTool(toolName string) error {
	manifest, err := ipe.FolderStruct.GetInstalledTool(toolName)
	if err != nil{
		return err
	}
	dag := graph.BuildGraph(manifest)
	dag.Run()
	return nil
}

