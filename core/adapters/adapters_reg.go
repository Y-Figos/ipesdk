package adapters

import (
	"fmt"
	"reflect"

	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/Y-Figos/ipesdk/core/ports"
)

type InputAdapterFactory func(args map[string]any) (ports.InputInterface, error)

var InputAdapterRegistry = map[string]InputAdapterFactory{
	"csv":        CSVReaderFactory,
	"excel":      ExcelReaderFactory,
	"tim_reader": TimLogReader,
	//
}

type OutputAdapterFactory func(args map[string]any, payload map[string]*df.Dataframe) (ports.OutputInterface, error)

var OutAdapterRegistry = map[string]OutputAdapterFactory{
	//"csv": CSVWriterFactory,
	"excel": ExcelAdapterFactory,
}

// Refactor this later to comply with new factory signature, csv output temporary dropped
func CSVWriterFactory(args map[string]any, payload *df.Dataframe) (ports.OutputInterface, error) {
	return &CSVWriter{
		Data:     payload,
		FilePath: args["filepath"].(string),
	}, nil
}

func CSVReaderFactory(args map[string]any) (ports.InputInterface, error) {
	filePath, ok := args["filepath"].(string)
	if !ok || filePath == "" {
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
	if !ok || filePath == "" {
		return nil, fmt.Errorf("'filepath' is required and must be a string")
	}
	headerRow, ok := args["header_row"].(float64)
	if !ok {
		return nil, fmt.Errorf("'headerRow' is required and must be a int: %v", reflect.TypeOf(args["header_row"]))
	}
	ColumnDefault := 1
	ColumnStart, ok := args["start_column"].(float64)
	if !ok {
		return nil, fmt.Errorf("'headerRow' is required and must be a int: %v", reflect.TypeOf(args["header_row"]))
	}
	if ColumnStart != 0 {
		ColumnDefault = int(ColumnStart)
	}

	sheetName, ok := args["sheet_name"].(string)
	if !ok || sheetName == "" {
		sheetName = ""
	}
	return &XLSXReader{
		SheetName:   sheetName,
		HeaderRow:   int(headerRow),
		FilePath:    filePath,
		BaseInput:   &BaseInput{},
		StartColumn: ColumnDefault,
	}, nil
}

func ExcelAdapterFactory(args map[string]any, payload map[string]*df.Dataframe) (ports.OutputInterface, error) {
	filePath, ok := args["filepath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("'filepath' is required and must be a string")
	}
	HeaderDefault := 1
	headerRow, ok := args["header_row"].(float64)
	if !ok {
		headerRow = 1
	}
	if headerRow != 0 {
		HeaderDefault = int(headerRow)
	}

	ColumnDefault := 1
	ColumnStart, ok := args["start_column"].(float64)
	if !ok {
		ColumnDefault = 1
	}
	if ColumnStart != 0 {
		ColumnDefault = int(ColumnStart)
	}

	saveMode, ok := args["save_mode"].(string)
	if !ok {
		saveMode = ""
	}
	var sheetOrderString []string
	if saveMode == "" {
		saveMode = "multi_sheet"
		sheetOrder, ok := args["sheet_order"].([]interface{})
		if !ok || sheetOrder == nil {
			sheetOrder = nil
		}
		var err error
		sheetOrderString, err = InterfaceSliceToStringSlice(sheetOrder)
		if err != nil {
			return nil, err
		}
	}
	return &XLSXWriter{
		Data:        payload,
		FilePath:    filePath,
		SheetsOrder: sheetOrderString,
		HeaderRow:   HeaderDefault,
		StartColumn: ColumnDefault,
		SaveMode:    saveMode,
	}, nil
}

// Helper Function to convert interface slices to strings
func InterfaceSliceToStringSlice(raw []interface{}) ([]string, error) {
	strs := make([]string, len(raw))
	for i, v := range raw {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string", i)
		}
		strs[i] = s
	}
	return strs, nil
}

func TimLogReader(args map[string]any) (ports.InputInterface, error) {
	filePath, ok := args["filepath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("'filepath' is required and must be a string")
	}
	header_pattern, ok := args["header_pattern"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("'filepath' is required and must be a string")
	}

	return &LogReaderTim{
		Filepath:       filePath,
		HeadersPattern: header_pattern,
	}, nil
}
