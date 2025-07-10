package main

import (
	"log"
	"os"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/cmd"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/internal/file_handler"
)
func main()  {
	fs := file_handler.FolderStruct{}
	err := fs.Init()
	if err != nil{
		log.Printf("error while initializing appdata")
		os.Exit(1)
	}
	cmd.AppFS = fs
	cmd.Execute()
}