package adapters

import (
	"errors"
	"fmt"
	"log"

	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/xuri/excelize/v2"
)

type XLSXReader struct {
	Data      *df.Dataframe
	FilePath  string
	SheetName string
	file      *excelize.File
	HeaderRow int
	rows      [][]string
	BaseInput *BaseInput
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
		x.Close()
		log.Println(err)
		return nil, err
	}

	newDf := df.Dataframe{
		ColumnOrder: headers,
		Columns:     make(map[string]df.ColumnInterface),
	}
	samplesize := 100
	if samplesize > len(x.rows) {
		samplesize = len(x.rows)
	}

	sampleData, err := x.ReadSample(samplesize)
	if err != nil {
		return nil, err
	}

	x.BaseInput.SetupSchemaFromSample(headers, sampleData, &newDf)

	x.BaseInput.WriteRows(headers, x.rows[x.HeaderRow:], &newDf)

	return &newDf, nil
}

func (x *XLSXReader) Open() error {
	f, err := excelize.OpenFile(x.FilePath)
	if err != nil {
		return err
	}
	x.file = f
	row, err := x.file.GetRows(x.SheetName)
	if err != nil {
		return err
	}
	if len(row) <= 0 {
		return fmt.Errorf("file has no data")
	}
	x.rows = row
	return nil
}

func (x *XLSXReader) GetHeaders() ([]string, error) {
	if len(x.rows) > 0 {
		return x.rows[x.HeaderRow-1], nil
	}
	return nil, fmt.Errorf("file has no data")
}

func (x *XLSXReader) ReadSample(n int) ([][]string, error) {
	if len(x.rows) > 0 && n < len(x.rows) {
		return x.rows[1 : n+1], nil
	}
	return nil, fmt.Errorf("file has no data")
}

func (x *XLSXReader) Close() error {
	if err := x.file.Close(); err != nil {
		return err
	}
	return nil
}
