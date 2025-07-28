package adapters

import (
	"fmt"
	"path/filepath"
	"os"
	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/xuri/excelize/v2"
)

type XLSXWriter struct {
	Data 		map[string]*df.Dataframe
	FilePath   	string
	file       	*excelize.File
	SheetsOrder	[]string
	Template 	map[string]excelize.Style
	SaveMode 	string //Opional - default multisheet
	HeaderRow 	int
	StartColumn	int
	ActiveSheet string
}

func (x *XLSXWriter) Open() error {
	x.file = excelize.NewFile()
	return nil
}

func (x *XLSXWriter) createSheet(sheet string,data *df.Dataframe) (int,error) {
	index,err := x.file.NewSheet(sheet)
	if err != nil{
		return 0,fmt.Errorf("error while creating sheet: %v", err)
	}
	startCell, err := excelize.CoordinatesToCellName(x.StartColumn,x.HeaderRow)
	if err != nil{
		return 0,fmt.Errorf("error while inserting headers to sheet %v: %v",sheet, err)
	}
	headers := make([]interface{}, len(data.ColumnOrder))
	for i, h := range data.ColumnOrder {
		headers[i] = h
	}
	if err := x.file.SetSheetRow(sheet, startCell, &headers); err != nil {
		return 0,fmt.Errorf("error while inserting headers to sheet %v: %v",sheet, err)
	}

	rowCount := data.RowCount()
	for i := 0; i < rowCount; i++{
		row := data.RowSlice(i)
		rowData := make([]interface{}, len(row)) 
		for idx,value := range row {
			rowData[idx] = value	
		} 
		cell, _ := excelize.CoordinatesToCellName(x.StartColumn,x.HeaderRow+i+1)
		if err := x.file.SetSheetRow(sheet,cell, &rowData); err != nil{
			return	0,fmt.Errorf("error while inserting headers to sheet %v: %v",sheet, err)
		}
	}

	return index, nil 
}

func (x *XLSXWriter) ExportData() (map[string]any, error) {
	switch x.SaveMode {
		case "multi_sheet":
			x.Open()
			sheetMap := make(map[string]int, len(x.SheetsOrder))
			for _,sheet := range x.SheetsOrder{
			data, ok := x.Data[sheet]
			if !ok {
				return nil,fmt.Errorf("%v sheet not in payload, be sure the sheetnames and payloads have the same name", sheet)
			}
			index, err := x.createSheet(sheet, data); 
			if err != nil{
				return nil,err
			}
			sheetMap[sheet] = index
			}
			activeIndex, ok := sheetMap[x.ActiveSheet]
			if !ok {
				activeIndex = 1
			} 

			x.file.SetActiveSheet(activeIndex)
			if err := x.file.DeleteSheet("Sheet1"); err != nil{
				return nil,fmt.Errorf("error closing exported file: %v", err)
			}
			if err := x.file.SaveAs(x.FilePath); err != nil {
				return nil,fmt.Errorf("error closing exported file: %v", err)
			}
			return map[string]any{
			"mode":        "multi_sheet",
			"file_path":   x.FilePath,
			"sheet_names": x.SheetsOrder,
			"active":      x.ActiveSheet,
			}, nil
		case "multi_file":
			info, err := os.Stat(x.FilePath)
				if err != nil {
					return nil,err // path might not exist
				}
			if !info.IsDir(){
				return nil,fmt.Errorf("path must be a directory")
			}
			if _, err := os.Stat(x.FilePath); os.IsNotExist(err) {
				if err := os.MkdirAll(x.FilePath, os.ModePerm); err != nil {
					return nil,fmt.Errorf("failed to create output directory: %w", err)
				}
			}
			exportedPaths := []string{}
			for name,data := range x.Data {
				tempFileName := filepath.Join(x.FilePath,name +".xlsx")	
				if err := x.Open(); err != nil {
					return nil,fmt.Errorf("failed to open Excel file: %w", err)
				}
				index, err := x.createSheet(name, data) 
				if err != nil {
					return nil,err
				}
				x.file.SetActiveSheet(index)
				if err := x.file.DeleteSheet("Sheet1"); err != nil{
					return nil,fmt.Errorf("error closing exported file: %v", err)
				}
				if err := x.file.SaveAs(tempFileName); err != nil {
					return nil,err
				}
				exportedPaths = append(exportedPaths, tempFileName)
			}
			
			return map[string]any{
			"mode":         "multi_file",
			"directory":    x.FilePath,
			"file_paths":   exportedPaths,
			"sheet_count":  len(exportedPaths),
		}, nil
	default:
		return nil, fmt.Errorf("invalid save mode: %s", x.SaveMode)
	}
}

func (x *XLSXWriter) Close() error {
	//
	return nil
}
