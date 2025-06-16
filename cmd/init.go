package cmd

import (
	"os"
	"path/filepath"
	"fmt"
)

func Init(projectName string,wd string, modulesQty int ) error{
	if wd == ""{
		newWd, err := os.Getwd()
		if err != nil{
			return err
		}
		wd = newWd
	}
	luaContent := `function main()
	-- Your Code here
end
	`

	root := filepath.Join(wd,projectName)
	
	for i := 0; i < modulesQty; i++{
		jsonContent := fmt.Sprintf(`{"id":"Module_%v","adapter":"","depends":[], "export_as":""}`, i+1)
		folderName := fmt.Sprintf("Module_%v", i+1)
		moduleFolder := filepath.Join(root,"modules",folderName)
		if err := os.MkdirAll(moduleFolder, 0755); err != nil {
			return err
		}
		luaFile := filepath.Join(moduleFolder, "script.lua")
		config := filepath.Join(moduleFolder, "config.json")
		if err := os.WriteFile(config, []byte(jsonContent) , 0644); err != nil {
			return err
		}

		err := os.WriteFile(luaFile, []byte(luaContent), 0644)
		if err != nil {
		panic(err)
		}
	}
	return nil
}