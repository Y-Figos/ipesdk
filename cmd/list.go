package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var listModules = &cobra.Command{
	Use: "list",
	Short: "List all installed .ipe tools",
	RunE: func (cmd *cobra.Command, args []string) error{
		reg, err := AppFS.ParseToolRegistry()
		if err != nil {
			return err
		} 
		for _, tool := range reg.ToolList{
			fmt.Printf("%s\nDesc: %s \n", tool.Name, tool.Desc)
		}
		return nil		
	} ,
}

func init() {
	rootCmd.AddCommand(listModules)
}