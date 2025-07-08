package file_handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)
type FolderStruct struct {
	AppData string
	ToolsFolder string
	ToolRegPath string
} 

type ToolRegistry struct {
	ToolList map[string]ToolInfo `json:"tool_list"`
}

func (fs *FolderStruct) Init() error {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return fmt.Errorf("APPDATA not found")
	}
	fs.ToolsFolder = filepath.Join(appData,"IPE", "tools")
	fs.AppData = filepath.Join(appData,"IPE")
	fs.ToolRegPath = filepath.Join(fs.AppData, "tool_registry.json")
	err := os.MkdirAll(fs.ToolsFolder, os.ModePerm)
	if err != nil{
		return err
	}

	if _, err := os.Stat(fs.ToolRegPath); os.IsNotExist(err){
		log.Println("tool registry does not exist, creating file")
		os.Create(fs.ToolRegPath)
		data, _ := json.MarshalIndent(&ToolRegistry{make(map[string]ToolInfo)},"","")
		os.WriteFile(fs.ToolRegPath, data, 0664)
	}
	log.Println("folder structure initialized succssesfuly")
	return nil
}


func (fs *FolderStruct) ParseToolRegistry() (*ToolRegistry, error) {
	data, err := os.ReadFile(fs.ToolRegPath)
	if err != nil{
		return nil,err
	}
	toolRegistry := ToolRegistry{make(map[string]ToolInfo)}

	err = json.Unmarshal(data, &toolRegistry)
	if err != nil{	
		return nil, fmt.Errorf("error parsing manifest: %v", err)
	}
	return &toolRegistry, nil
}

func (fs *FolderStruct) GetInstalledTool(toolName string) (*Manifest, error){
	toolRegistry, err := fs.ParseToolRegistry()
	if err != nil{
		return nil, fmt.Errorf("failed to parse Tool Registry File: %w", err)
	}
	tool, ok := toolRegistry.ToolList[toolName]
	if !ok {
		return nil, fmt.Errorf("tool not Installed: %w", err)
	}
	toolManifestPath := filepath.Join(fs.ToolsFolder, tool.Name, "manifest.json")
	log.Printf("Tool Found, loading %v \n", toolName)
	return ParseManifest(toolManifestPath)
}

func (fs *FolderStruct) RegisterTool(toolFolder string) error {
	toolRegistry, err := fs.ParseToolRegistry()
	if err != nil{
		return fmt.Errorf("failed to parse Tool Registry File: %w", err)
	}
	toolManifest, err := ParseManifest(filepath.Join(toolFolder, "manifest.json"))
	if err != nil {
		return fmt.Errorf("failed to parse Tool Manifest File: %w", err)
	}

	if toolManifest.Tool.Name == "" || toolManifest.Tool.Version == "" {
		return fmt.Errorf("invalid manifest")
	} 

	toolInfo := toolManifest.Tool
	id := toolInfo.Name
	
	toolRegistry.ToolList[id] = toolInfo
	
	data,err := json.MarshalIndent(toolRegistry,"","")
	if err != nil {
		return fmt.Errorf("failed to Marshal Registry File: %w", err)
	}
	
	err = os.WriteFile(fs.ToolRegPath,data, 0664)
	if err != nil{
		return fmt.Errorf("failed to Write Registry File: %w", err)
	}
	return nil
}

func (fs *FolderStruct) Extract(ipeFile string) error {
	toolName := strings.TrimSuffix(filepath.Base(ipeFile), filepath.Ext(ipeFile))
	toolFolder := filepath.Join(fs.ToolsFolder, toolName)
	err := ExtractIPE(ipeFile, toolFolder)
	if err != nil {
		log.Fatal("Extraction failed:", err)
	}
	err = fs.RegisterTool(toolFolder)
	if err != nil {
		return fmt.Errorf("failed to register tool: %w", err)
	}
	return nil
}

