package adapters

import(
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/ports"
	"fmt"
)

type InputAdapterFactory func(args map[string]any) (ports.InputInterface, error)

var InputAdapterRegistry = map[string]InputAdapterFactory{
	"csv": CSVAdapterFactory,
	// "excel": ExcelAdapterFactory, etc.
}

func CSVAdapterFactory(args map[string]any) (ports.InputInterface, error) {
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