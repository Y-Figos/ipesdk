package adapters

import (
	"fmt"
	"reflect"

	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/Y-Figos/ipesdk/core/ports"
)

type InputAdapterFactory func(args map[string]any) (ports.InputInterface, error)

var InputAdapterRegistry = map[string]InputAdapterFactory{
	"csv": CSVReaderFactory,
	"excel": ExcelReaderFactory,
	//
}
type OutputAdapterFactory func(args map[string]any,payload *df.Dataframe) (ports.OutputInterface, error)
var OutAdapterRegistry = map[string]OutputAdapterFactory{
	"csv":CSVWriterFactory ,
	// "excel": ExcelAdapterFactory, etc.
}

func CSVWriterFactory(args map[string]any, payload *df.Dataframe) (ports.OutputInterface, error){
	return &CSVWriter{
		Data: payload,
		FilePath: args["filepath"].(string),
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

func ExcelReaderFactory(args map[string]any) (ports.InputInterface, error) {
	filePath, ok := args["filepath"].(string)
	if !ok || filePath == ""{
		return nil, fmt.Errorf("'filepath' is required and must be a string")
	}
	headerRow, ok := args["header_row"].(float64)
	if !ok {
		return nil, fmt.Errorf("'headerRow' is required and must be a int: %v", reflect.TypeOf(args["header_row"]))
	}
	sheetName, ok := args["sheet_name"].(string)
	if !ok || sheetName == ""{
		return nil, fmt.Errorf("'sheet_name' is required and must be a string")
	}
	return &XLSXWriter{
		SheetName: sheetName,
		HeaderRow: int(headerRow),
		FilePath: filePath,
		BaseInput: &BaseInput{},
	}, nil;
}