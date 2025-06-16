package file_handler

import (
	"encoding/json"
	"fmt"
	"os"
)

type ToolInfo struct {
	Name string `json:"name"`
	Version string `json:"version"`
}

type Node struct {
	Id string `json:"id"`
	Adapter string `json:"adapter"`
	Export string `json:"export_as"`
	Depends []string `json:"depends"`
}

type Manifest struct {
	ManifestPath string
	Tool ToolInfo `json:"tool"`
	NodeList []Node `json:"nodes"`
}

func ParseManifest(manifestPath string) (*Manifest, error){
	file, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error when reading manifest: %v", err)
	}
	data := Manifest{ManifestPath: manifestPath}
	if err := json.Unmarshal(file, &data); err != nil{
		return nil, fmt.Errorf("error parsing manifest: %v", err)
	}
	return &data, nil
}