package main

import (
	_ "fmt"
	"log"
	_ "runtime"
	_ "strings"
	_ "time"

	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/adapters"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
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

	filtered := newdf.Filter(func(row map[string]any) bool {
		return row["Board Name"] == "MRRU" 
	}) 

	log.Println(filtered.RowCount())
	// Memory after
	// runtime.ReadMemStats(&memStatsEnd)
	// elapsed := time.Since(start)

	// fmt.Println("Execution Time:", elapsed)

	// // Print memory stats (in MB)
	// alloc := float64(memStatsEnd.Alloc) / (1024 * 1024)
	// totalAlloc := float64(memStatsEnd.TotalAlloc) / (1024 * 1024)
	// sys := float64(memStatsEnd.Sys) / (1024 * 1024)
	// peak := float64(memStatsEnd.Alloc - memStatsStart.Alloc) / (1024 * 1024)

	// fmt.Printf("Memory Allocated: %.2f MB\n", alloc)
	// fmt.Printf("Total Memory Allocated: %.2f MB\n", totalAlloc)
	// fmt.Printf("System Memory: %.2f MB\n", sys)
	// fmt.Printf("Peak Usage During Call: %.2f MB\n", peak)

}