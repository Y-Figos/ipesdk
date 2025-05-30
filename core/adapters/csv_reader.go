package adapters

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"sync"

	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
	_ "codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/ports"
)

type CSVReader struct {
	FilePath string
	BatchSize int
	WorkerCount int
}

func ColumnWriter(channel chan [][]string, headers []string, dataframe *df.Dataframe , wg *sync.WaitGroup,mu *sync.Mutex){
	defer wg.Done()
	
	for batch := range channel{
		mu.Lock()
		for _, row := range batch{
			for columnIndex, data := range row{
				
				dataframe.Columns[headers[columnIndex]].AppendValue(data)
				
			}
		}
		mu.Unlock() 
		
	}
	
}


func (c *CSVReader) GetData() (*df.Dataframe, error) {
	file, err := os.Open(c.FilePath)
	
	mu := &sync.Mutex{}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	batchChannel := make(chan [][]string, 100)
	var wg sync.WaitGroup
	
	reader := csv.NewReader(file)
	
	headers, err := reader.Read()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	newDf := df.Dataframe{
		ColumnOrder: headers,
		Columns:     make(map[string]df.ColumnInterface),
	}
	// Initialize all columns
	for _, header := range headers {
	col := df.NewColumn[string](header, []string{}) // Chore: Create type assertion Logic
	newDf.Columns[header] = col 
	}
	// Start workers
	for i := 0; i < c.WorkerCount; i++ {
		wg.Add(1)
		go ColumnWriter(batchChannel, headers, &newDf, &wg, mu)
	}

	// Read and send batches
	for {
		var batch [][]string

		for i := 0; i < c.BatchSize; i++ {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("read error: %v", err)
				continue
			}
			batch = append(batch, record)
		}
		if len(batch) > 0 {
        batchChannel <- batch
    	}
		if len(batch) == 0 {
			break
		}
		
	}

	close(batchChannel)
	wg.Wait()
	log.Println("All workers done")
	return &newDf, nil
}