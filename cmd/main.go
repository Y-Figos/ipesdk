package main

import (
	_ "fmt"
	"log"
	_ "runtime"
	_ "strings"
	_ "time"

	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/adapters"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df/infer"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/engine"
)
func main()  {
	// start := time.Now()

	// Track memory usage
	// var memStatsStart, memStatsEnd runtime.MemStats
	// runtime.ReadMemStats(&memStatsStart)

	teste := adapters.CSVReader{
		FilePath:    "D:\\Huawei Projects\\1- Project IPE\\Files\\users_100.csv", 
		BatchSize:   10000,
		WorkerCount: 8,
	}

	newdf, err := teste.GetData()
	if err != nil{
		log.Fatalln(err)
	}
	log.Println(newdf.RowCount())

	filtered := newdf.Filter(func(row map[string]any) bool { 
		return row["id"] == 1
	})
	
	filtered.Columns["TestColumn"] = filtered.Columns["id"].Map(map[any]any{1:"FIRST USER"}, "TestColumn")
	
	log.Println(filtered.Columns["name"])
	log.Println(filtered.Columns["TestColumn"])

}