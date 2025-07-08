package main

import (
	"log"
	"os"
	_ "runtime"
	_ "strings"
	_ "time"
	
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/cmd"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/adapters"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df/infer"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/engine"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/internal"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/internal/file_handler"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/internal/graph"
	_"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/utils/cmd_utils"
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
	// err = fs.Extract("D:\\Huawei Projects\\1- Project IPE\\Files\\Teste.ipe")
	// if err != nil{
	// 	log.Printf("%v",err)
	// 	os.Exit(1)
	// }
	// os.Exit(0)

	

	// //Flags for Run Command
	// runFlags := flag.NewFlagSet("run", flag.ExitOnError)
	// runInput := runFlags.String("input", "","Input File Path" )
	// runOutput := runFlags.String("output", "","Output File Path" )
	
	// if len(os.Args) < 2{
	// 	fmt.Println("Usage: ipe [run|list|...]")
	// 	return
	// }
	// command := os.Args[1]
	// switch command {
	// 	case "run":
	// 		runFlags.Parse(os.Args[2:])
	// 		if runFlags.NArg() < 1 {
	// 			fmt.Println("Usage: ipe run [toolName | path/to/tool.ipe | path/to/toolDir] -input ... -output ...")
	// 			os.Exit(1)
	// 		}
	// 		toolArg := runFlags.Args()[0]
			
	// 		manifest, err := cmdutils.ResolveToolArg(toolArg, fs)
	// 		if err != nil{
	// 			log.Printf("error resolving run arguments: %s", err)
	// 			os.Exit(1)
	// 		}
			
	// 		err = cmd.RunIPE(manifest,*runInput,*runOutput)
	// 		if err != nil{
	// 			fmt.Printf("error while running .ipe file: %s", err)	
	// 		}
	// }
	// cmd.BuildIpeFromFolder("0.2","D:\\Huawei Projects\\1- Project IPE\\IPEGOCORE\\toolTest","D:\\Huawei Projects\\1- Project IPE\\Files\\Teste.ipe")
	// cmd.RunIPE("D:\\Huawei Projects\\1- Project IPE\\Files\\Teste.ipe")
	// cmd.Init("ProjectTest","",5)
	// file_handler.CreateManifest("D:\\Huawei Projects\\1- Project IPE\\IPEGOCORE\\toolTest", "0.2")
	// manifest, err := file_handler.ParseManifest("D:\\Huawei Projects\\1- Project IPE\\IPEGOCORE\\toolTest\\manifest.json")
	// if err != nil {
	// 	log.Println(err)
	// }

	// dag := graph.BuildGraph(manifest)
	
	// dag.Run()
	// dag.Nodes["Module_2"].Export()
	// start := time.Now()

	// Track memory usage
	// var memStatsStart, memStatsEnd runtime.MemStats
	// runtime.ReadMemStats(&memStatsStart)
	//engine.Teste("D:\\Huawei Projects\\1- Project IPE\\IPEGOCORE\\core\\engine\\example.lua")

	// nm := engine.NodeModule{
	// 	ModuleName: "Teste",
	// 	ScriptPath: "D:\\Huawei Projects\\1- Project IPE\\IPEGOCORE\\core\\engine\\example.lua",
	// }

	// nm.Run()

	// log.Println(nm.Payload.Shape())
	// teste := adapters.CSVReader{
	// 	FilePath:    "D:\\Huawei Projects\\1- Project IPE\\Files\\users_100.csv", 
	// 	BatchSize:   10000,
	// 	WorkerCount: 8,
	// }

	// newdf, err := teste.GetData()
	// if err != nil{
	// 	log.Fatalln(err)
	// }
	// log.Println(newdf.RowCount())

	// filtered := newdf.Filter(func(row map[string]any) bool { 
	// 	return row["name"] != "" && row["status"] == "active" 
	// })
	
	// filtered.Columns["TesteApply"] = filtered.Columns["age"].Apply(func(a any) any {
	// 	if a == "Eve"{
	// 		return "Is EVE!"
	// 	}
	// 	return "Not Eve"
	// })

	// filtered.Columns["TestColumn"] = filtered.Columns["id"].Map(map[any]any{1:"FIRST USER"}, "TestColumn")
	
	// log.Println(filtered.Columns["name"].Unique())
	// log.Println(filtered.Columns["TestColumn"])

}