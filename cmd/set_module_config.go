package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Y-Figos/ipesdk/core/file_handler"
	"github.com/spf13/cobra"
)

var inputPath string
var outputPath string

var setModules = &cobra.Command{
	Use:   "set",
	Short: "Set modules input and output .ipe tools",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		toolName := args[0]
		moduleName := args[1]
		manifest, err := AppFS.GetInstalledTool(toolName)
		if err != nil {
			return fmt.Errorf("tool does not exist or no installed")

		}
		if inputPath != "" {
			manifest.NodeList[moduleName].InputArgs["filepath"] = inputPath
		}
		if outputPath != "" {
			manifest.NodeList[moduleName].OutputArgs["filepath"] = outputPath
		}
		newmanifest := filepath.Join(AppFS.ToolsFolder, toolName, "manifest.json")
		file_handler.WriteManifest(newmanifest, manifest)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	setModules.Flags().StringVarP(&inputPath, "input", "i", "", "Set input of module")
	setModules.Flags().StringVarP(&outputPath, "output", "o", "", "Set output of module")
	rootCmd.AddCommand(setModules)
}
