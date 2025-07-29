package cmd

import (
	"os"

	"github.com/Y-Figos/ipesdk/core/file_handler"
	"github.com/spf13/cobra"
)

var ipeOutput string
var toolVersionFlag string

var buildTool = &cobra.Command{
	Use: "build",
	Short: "Build the IPE folder structure into the .ipe file format",
	Args: cobra.ExactArgs(1),
	RunE: func (cmd *cobra.Command, args []string) error{
		filePath := args[0]
		if _, err := os.Stat(filePath); os.IsNotExist(err){
			return err
		}else if err != nil{
			return err
		}
		if _, err := os.Stat(ipeOutput); os.IsNotExist(err){
			return err
		} else if err != nil{
			return err
		}
		if _, err := file_handler.CreateManifest(filePath,toolVersionFlag); err != nil {
			return err
		}
		file_handler.BuildIpeFromFolder(toolVersionFlag,filePath , ipeOutput)
		return nil		
	} ,
}

func init() {
	buildTool.Flags().StringVarP(&ipeOutput, "output", "o", "", "Set output of module")
	buildTool.Flags().StringVarP(&toolVersionFlag, "version", "v", "", "Set output of module")
	buildTool.MarkFlagRequired("output")
	buildTool.MarkFlagRequired("version")
	rootCmd.AddCommand(buildTool)
}