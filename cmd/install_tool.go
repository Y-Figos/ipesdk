package cmd

import (
	"github.com/spf13/cobra"
)

var installTool = &cobra.Command{
	Use: "install",
	Short: "Build the IPE folder structure into the .ipe file format",
	Args: cobra.ExactArgs(1),
	RunE: func (cmd *cobra.Command, args []string) error{
		filePath := args[0]
		
		if err := AppFS.Extract(filePath); err != nil {
			return err
		}
		return nil		
	} ,
}

func init() {
	rootCmd.AddCommand(installTool)

}
