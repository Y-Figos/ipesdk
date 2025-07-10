package cmd

import (
	"os"
	"path/filepath"
	"fmt"
	"github.com/spf13/cobra"
)


var wd string
var modulesQty int

var InitModule = &cobra.Command{
	Use: "init",
	Short: "Initialize a .ipe folder structure for devlopment",
	Args: cobra.ExactArgs(1),
	RunE: func (cmd *cobra.Command, args []string) error {
		projectName := args[0]
		if wd == ""{
		newWd, err := os.Getwd()
		if err != nil{
			return err
		}
		wd = newWd
	}
	luaContent := `function main()
	-- Your Code here
	return {}
end
	`

	root := filepath.Join(wd,projectName)
	
	for i := 0; i < modulesQty; i++{
		jsonContent := fmt.Sprintf(`{"id":"Module_%v","adapter":"","depends":[], "export_as":"","input_args":{}, "output_args":{}}`, i+1)
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
	},
}

func init(){
	InitModule.Flags().StringVarP(&wd, "wd", "w", "", "Working directory (optional)")
	InitModule.Flags().IntVarP(&modulesQty, "modules", "n", 0, "Number of modules to create")
	InitModule.MarkFlagRequired("modules")
	rootCmd.AddCommand(InitModule)
}