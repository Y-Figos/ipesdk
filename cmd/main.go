package main

import (
	_ "fmt"
	"log"
	_ "runtime"
	_ "strings"
	_ "time"

	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/adapters"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df/infer"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/engine"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/internal/file_handler"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/internal/graph"
)
func main()  {



	manifest, err := file_handler.ParseManifest("D:\\Huawei Projects\\1- Project IPE\\IPEGOCORE\\toolTest\\manifest.json")
	if err != nil {
		log.Println(err)
	}

	dag := graph.BuildGraph(manifest)

	dag.Run()
	
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
	// 	return row["name"] != ""
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