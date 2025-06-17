package adapters

import (
	"reflect"
	"sync"

	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df/infer"
)

type BaseInput struct{}

func (bi *BaseInput) SetupSchemaFromSample(headers []string, sampleData [][]string, newDf *df.Dataframe) {
	sampleMap := make(map[string][]string)
	typeMap := make(map[string]reflect.Type)
	for _, row := range sampleData {
		for columnIndex, value := range row {
			if columnIndex >= len(headers) {
				continue
			}
			header := headers[columnIndex]
			sampleMap[header] = append(sampleMap[header], value)
		}
	}

	for index, column := range sampleMap {
		typeMap[index] = infer.InferTypeFromSlice(column)
	}

	for _, header := range headers {
		newDf.Columns[header] = infer.CreateTypedColumn(header, typeMap[header])
	}

	for _, row := range sampleData {
		for columnIndex, data := range row {
			newDf.Columns[headers[columnIndex]].AppendValue(data)
		}
	}
}

func (bi *BaseInput) StartWorkers(batchChan chan [][]string, headers []string, df *df.Dataframe, workerCount int) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go bi.ColumnWriter(batchChan, headers, df, wg, mu)
	}
	return wg
}

func (bi *BaseInput) ColumnWriter(channel chan [][]string, headers []string, dataframe *df.Dataframe, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	for batch := range channel {
		mu.Lock()
		for _, row := range batch {
			for columnIndex, data := range row {

				dataframe.Columns[headers[columnIndex]].AppendValue(data)

			}
		}
		mu.Unlock()
	}

}
