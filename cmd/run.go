package cmd

import (
	// "os"
	"path/filepath"
	"github.com/Y-Figos/ipesdk/internal/file_handler"
	"github.com/Y-Figos/ipesdk/core/graph"
	"github.com/spf13/cobra"
)

var testFlag bool
var export bool

var runTool = &cobra.Command{
	Use: "run",
	Short: "Build the IPE folder structure into the .ipe file format",
	Args: cobra.ExactArgs(1),
	RunE: func (cmd *cobra.Command, args []string) error{
		toolName := args[0] //Can be either a tool name or path for a test folder
		var manifest *file_handler.Manifest
		var err error

		if testFlag {
			file_handler.CreateManifest(toolName,"1.0")
			manifestPath := filepath.Join(toolName, "manifest.json")
			manifest, err = file_handler.ParseManifest(manifestPath)
			if err != nil {
				return err
			}
		}else {
			manifest,err = AppFS.GetInstalledTool(toolName)
			if err != nil {
				return err
			}
		}
		dag := graph.BuildGraph(manifest)

		dag.Run()
		
		return nil		
	} ,
}

func init() {
	runTool.Flags().BoolVarP(&testFlag, "test", "t", false, "Use for testing tools eg. still in folder")
	runTool.Flags().BoolVarP(&export, "export", "o", false, "Set to export modules output")
	rootCmd.AddCommand(runTool)
}