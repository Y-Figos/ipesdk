package file_handler

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type ToolInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Desc	string `json:"description"`
}

type Node struct {
	Id        	string        	 `json:"id"`
	Adapter   	string        	 `json:"adapter"`
	Depends  	[]string      	 `json:"depends"`
	InputArgs 	map[string]any	 `json:"input_args"`
	OutputArgs 	map[string]any	 `json:"output_args"`
	ExportFlag	bool			 `json:"export"`
}

type Manifest struct {
	ManifestPath string   `json:"-"`
	Tool         ToolInfo `json:"tool"`
	NodeList     map[string]Node   `json:"nodes"`
}

func ParseManifest(manifestPath string) (*Manifest, error) {
	file, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error when reading manifest: %v", err)
	}
	data := Manifest{ManifestPath: manifestPath}
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("error parsing manifest: %v", err)
	}
	return &data, nil
}

func CreateManifest(projectFolder string, version string) (*Manifest, error) {
	nodeList := make(map[string]Node)
	modulesFolder := filepath.Join(projectFolder, "modules")
	err := filepath.WalkDir(modulesFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "config.json" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var node Node
			if err := json.Unmarshal(data, &node); err != nil {
				return fmt.Errorf("error parsing manifest: %v", err)
			}
			nodeList[node.Id] = node
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking the directory:", err)
	}
	newManifest := Manifest{
		Tool:     ToolInfo{Name: filepath.Base(projectFolder), Version: version},
		NodeList: nodeList,
	}
	jsonBytes, err := json.MarshalIndent(newManifest, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil,err
	}
	err = os.WriteFile(filepath.Join(projectFolder, "manifest.json"), jsonBytes, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return nil,err
	}
	return &newManifest, nil
}
