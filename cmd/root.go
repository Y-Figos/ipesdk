package cmd

import (
	"fmt"
	"github.com/Y-Figos/ipesdk/core/file_handler"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "ipe",
	Short: "IPE - Integrated Pipeline Engine",
}
var AppFS file_handler.FolderStruct
func Execute(){
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
	}
}