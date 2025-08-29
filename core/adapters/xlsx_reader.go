package adapters

import (
	"errors"
	"fmt"
	"log"

	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/xuri/excelize/v2"
)

type XLSXReader struct {
	Data        *df.Dataframe
	FilePath    string
	SheetName   string
	file        *excelize.File
	HeaderRow   int
	StartColumn int
	rows        [][]string
	BaseInput   *BaseInput
}

func (x *XLSXReader) GetData() (*df.Dataframe, error) {
	err := x.Open()
	if err != nil {
		return nil, err
	}
	defer x.Close()

	if x.BaseInput == nil {
		return nil, errors.New("BaseInput is not initialized")
	}

	headers, err := x.GetHeaders()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	newDf := df.Dataframe{
		ColumnOrder: headers,
		Columns:     make(map[string]df.ColumnInterface),
	}

	samplesize := 10
	dataRows := len(x.rows) - x.HeaderRow // number of rows after header
	if dataRows < 1 {
		return nil, fmt.Errorf("no data rows found after header")
	}
	if samplesize > dataRows {
		samplesize = dataRows
	}

	sampleData, err := x.ReadSample(samplesize)
	if err != nil {
		return nil, err
	}

	x.BaseInput.SetupSchemaFromSample(headers, sampleData, &newDf)

	// Write the rest of the rows after the header + sample
	x.BaseInput.WriteRows(headers, x.rows[x.HeaderRow+samplesize:], &newDf)

	log.Println("RowCount before padding:", newDf.RowCount())
	newDf.PadColumns()
	log.Println("RowCount after padding:", newDf.RowCount())

	return &newDf, nil
}

func (x *XLSXReader) Open() error {
	f, err := excelize.OpenFile(x.FilePath)
	if err != nil {
		return err
	}
	x.file = f

	sheetMap := f.GetSheetMap()

	// --- fallback logic ---
	if x.SheetName == "" || !sheetExists(f, x.SheetName) {
		if len(sheetMap) == 0 {
			return fmt.Errorf("file has no sheets")
		}

		if len(sheetMap) == 1 {
			// only one sheet in the file, use it regardless of active sheet
			for _, s := range sheetMap {
				x.SheetName = s
			}
		} else {
			// multiple sheets: fallback to active sheet
			activeIdx := f.GetActiveSheetIndex()
			activeSheet, ok := sheetMap[activeIdx]
			if !ok {
				return fmt.Errorf("active sheet index %d not found in sheet map", activeIdx)
			}
			x.SheetName = activeSheet
		}
	}
	// --- end fallback ---

	rows, err := f.GetRows(x.SheetName)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return fmt.Errorf("sheet %q has no data", x.SheetName)
	}

	// Determine max columns after trimming start column
	maxCols := 0
	trimmed := make([][]string, 0, len(rows))
	for _, r := range rows {
		var rowSlice []string
		if len(r) >= x.StartColumn {
			rowSlice = r[x.StartColumn-1:]
		} else {
			rowSlice = []string{}
		}
		if len(rowSlice) > maxCols {
			maxCols = len(rowSlice)
		}
		trimmed = append(trimmed, rowSlice)
	}

	// Pad rows with fewer columns with empty strings
	for i, r := range trimmed {
		if len(r) < maxCols {
			pad := make([]string, maxCols-len(r))
			trimmed[i] = append(r, pad...)
		}
		if len(trimmed[i]) != maxCols {
			return fmt.Errorf("row %d has len=%d, expected %d", i, len(trimmed[i]), maxCols)
		}
	}
	x.rows = trimmed
	return nil
}

func (x *XLSXReader) GetHeaders() ([]string, error) {
	if len(x.rows) >= x.HeaderRow {
		return x.rows[x.HeaderRow-1], nil
	}
	return nil, fmt.Errorf("getheaders,file has no data")
}

func (x *XLSXReader) ReadSample(n int) ([][]string, error) {
	if len(x.rows) <= x.HeaderRow {
		return nil, fmt.Errorf("readsample,file has no data")
	}

	// data rows start at index = HeaderRow
	start := x.HeaderRow
	end := start + n
	if end > len(x.rows) {
		end = len(x.rows)
	}

	return x.rows[start:end], nil
}

func (x *XLSXReader) Close() error {
	if err := x.file.Close(); err != nil {
		return err
	}
	return nil
}

func sheetExists(f *excelize.File, name string) bool {
	for _, s := range f.GetSheetMap() {
		if s == name {
			return true
		}
	}
	return false
}
