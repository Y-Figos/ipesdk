package file_handler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
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
	}
	log.Println("folder structure initialized succssesfuly")
	return nil
}


func (fs *FolderStruct) parseToolRegistry() (*ToolRegistry, error) {
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

func (fs *FolderStruct) RegisterTool(toolFolder string) error {
	toolRegistry, err := fs.parseToolRegistry()
	if err != nil{
		return fmt.Errorf("failed to parse Tool Registry File: %w", err)
	}
	toolManifest, err := ParseManifest(toolFolder)
	if err != nil {
		return fmt.Errorf("failed to parse Tool Manifest File: %w", err)
	}

	if toolManifest.Tool.Name == "" || toolManifest.Tool.Desc == "" || toolManifest.Tool.Version == "" {
		return fmt.Errorf("invalid manifest")
	} 

	toolInfo := toolManifest.Tool
	id := fmt.Sprintf("%s@%s",toolInfo.Name, toolInfo.Version)
	
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
	//Later to be added to a specific function
	toolName := strings.TrimSuffix(filepath.Base(ipeFile), filepath.Ext(ipeFile))
	
	toolFolder := filepath.Join(fs.ToolsFolder, toolName)
	
	archive, err := zip.OpenReader(ipeFile)
	if err != nil {
    	return fmt.Errorf("failed to open zip file: %w", err)
	}	
	defer archive.Close()
	
	for _, f := range archive.File {
		if f.Name == "" {
			continue
		}
		relPath := filepath.Join(toolFolder, f.Name)
		if !strings.HasPrefix(relPath, filepath.Clean(toolFolder)+string(os.PathSeparator)) {
            return fmt.Errorf("invalid file path")
        }	
		if f.FileInfo().IsDir() {
            log.Println("creating directory...")
            os.MkdirAll(relPath, os.ModePerm)
            continue
        }
		if err := os.MkdirAll(filepath.Dir(relPath), os.ModePerm); err != nil {
            return fmt.Errorf("failed to create tool directories: %w", err)
        }
		dstFile, err := os.OpenFile(relPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
    	if err != nil {
        	return err
    	}	
		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}
		
		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}
		fileInArchive.Close()
		dstFile.Close()
		}

	return nil
}
