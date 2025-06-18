package adapters

import (
	"fmt"

	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/ports"
)

type InputAdapterFactory func(args map[string]any) (ports.InputInterface, error)

var InputAdapterRegistry = map[string]InputAdapterFactory{
	"csv": CSVReaderFactory,
	// "excel": ExcelAdapterFactory, etc.
}
type OutputAdapterFactory func(args map[string]any,payload *df.Dataframe) (ports.OutputInterface, error)
var OutAdapterRegistry = map[string]OutputAdapterFactory{
	"csv":CSVWriterFactory ,
	// "excel": ExcelAdapterFactory, etc.
}

func CSVWriterFactory(args map[string]any, payload *df.Dataframe) (ports.OutputInterface, error){
	return &CSVWriter{
		Data: payload,
		FilePath: args["file_path"].(string),
	}, nil
}

func CSVReaderFactory(args map[string]any) (ports.InputInterface, error) {
	filePath, ok := args["filepath"].(string)
	if !ok || filePath == ""{
		return nil, fmt.Errorf("'filepath' is required and must be a string")
	}

	batchSize := 100
	if val, ok := args["batch_size"].(float64); ok {
		batchSize = int(val)
	}
	workerCount := 4
	if val, ok := args["worker_count"].(float64); ok {
		workerCount = int(val)
	}

	return NewCSVReaderWithOptions(filePath, batchSize, workerCount), nil
}