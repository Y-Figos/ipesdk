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
		FilePath:    "D:\\Huawei Projects\\InvToolProject\\Files\\BLZ\\Inventory_Board_20250505_102510.csv", 
		BatchSize:   10000,
		WorkerCount: 8,
	}

	newdf, err := teste.GetData()
	if err != nil{
		log.Fatalln(err)
	}
	log.Println(newdf.RowCount())

	// filtered := newdf.Filter(func(row map[string]any) bool { 
	// 	return row["Board Name"] == "MRRU" || row["Board Name"] == "FModule"
	// })
	
	

	log.Println(newdf.Columns["Subrack No."].Type())

}