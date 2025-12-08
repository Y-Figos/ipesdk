package adapters

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/xuri/excelize/v2"
)

type XLSXWriter struct {
	Data        map[string]*df.Dataframe
	FilePath    string
	file        *excelize.File
	SheetsOrder []string
	Template    map[string]excelize.Style
	SaveMode    string // Optional - default multi_sheet
	HeaderRow   int
	StartColumn int
	ActiveSheet string
}

func (x *XLSXWriter) Open() error {
	x.file = excelize.NewFile()
	return nil
}

// NOVO createSheet — só trabalha por coluna e loga o que tá escrevendo
func (x *XLSXWriter) createSheet(sheet string, data *df.Dataframe) (int, error) {
	if data == nil {
		return 0, fmt.Errorf("nil dataframe recebido para sheet %s", sheet)
	}

	log.Printf("[XLSXWriter] Creating sheet %s (rows=%d, cols=%d)",
		sheet, data.RowCount(), len(data.ColumnOrder))

	index, err := x.file.NewSheet(sheet)
	if err != nil {
		return 0, fmt.Errorf("error while creating sheet: %v", err)
	}

	// ===== HEADERS =====
	startCell, err := excelize.CoordinatesToCellName(x.StartColumn, x.HeaderRow)
	if err != nil {
		return 0, fmt.Errorf("error while inserting headers to sheet %v: %v", sheet, err)
	}

	headers := make([]interface{}, len(data.ColumnOrder))
	for i, h := range data.ColumnOrder {
		headers[i] = h
	}
	log.Printf("[XLSXWriter] Sheet %s headers: %v", sheet, headers)

	if err := x.file.SetSheetRow(sheet, startCell, &headers); err != nil {
		return 0, fmt.Errorf("error while inserting headers to sheet %v: %v", sheet, err)
	}

	// ===== ROWS =====
	rowCount := data.RowCount()
	colCount := len(data.ColumnOrder)

	for i := 0; i < rowCount; i++ {
		rowData := make([]interface{}, colCount)

		for j, colName := range data.ColumnOrder {
			col := data.Columns[colName]
			var v any

			if col != nil && i < col.Len() {
				v = col.GetValue(i)
			}

			if v == nil {
				rowData[j] = ""
			} else {
				rowData[j] = v
			}
		}

		// debug nas 3 primeiras linhas pra garantir que não tá vindo vazio
		if i < 3 {
			log.Printf("[XLSXWriter] Sheet=%s Row=%d -> %v", sheet, i, rowData)
		}

		cell, _ := excelize.CoordinatesToCellName(x.StartColumn, x.HeaderRow+i+1)
		if err := x.file.SetSheetRow(sheet, cell, &rowData); err != nil {
			return 0, fmt.Errorf("error while inserting row %d to sheet %v: %v", i, sheet, err)
		}
	}

	return index, nil
}

func (x *XLSXWriter) ExportData() (map[string]any, error) {
	switch x.SaveMode {
	case "multi_sheet":
		// um único arquivo com várias abas
		if err := x.Open(); err != nil {
			return nil, fmt.Errorf("failed to open Excel file: %w", err)
		}

		if len(x.SheetsOrder) == 0 {
			// fallback: usa as chaves do map em ordem aleatória
			for name := range x.Data {
				x.SheetsOrder = append(x.SheetsOrder, name)
			}
		}

		sheetMap := make(map[string]int, len(x.SheetsOrder))
		for _, sheet := range x.SheetsOrder {
			data, ok := x.Data[sheet]
			if !ok {
				return nil, fmt.Errorf("%v sheet not in payload, be sure the sheetnames and payloads have the same name", sheet)
			}
			index, err := x.createSheet(sheet, data)
			if err != nil {
				return nil, err
			}
			sheetMap[sheet] = index
		}

		activeIndex := 0
		if x.ActiveSheet != "" {
			if idx, ok := sheetMap[x.ActiveSheet]; ok {
				activeIndex = idx
			}
		}
		if activeIndex == 0 && len(sheetMap) > 0 {
			// pega qualquer índice
			for _, idx := range sheetMap {
				activeIndex = idx
				break
			}
		}

		if activeIndex > 0 {
			x.file.SetActiveSheet(activeIndex)
		}

		// deleta Sheet1 se existir (API nova: (int, error))
		if idx, err := x.file.GetSheetIndex("Sheet1"); err == nil && idx != -1 {
			if err := x.file.DeleteSheet("Sheet1"); err != nil {
				return nil, fmt.Errorf("error deleting default Sheet1: %v", err)
			}
		}

		log.Printf("[XLSXWriter] Salvando arquivo em: %s", x.FilePath)
		if err := x.file.SaveAs(x.FilePath); err != nil {
			return nil, fmt.Errorf("error closing exported file: %v", err)
		}

		return map[string]any{
			"mode":        "multi_sheet",
			"file_path":   x.FilePath,
			"sheet_names": x.SheetsOrder,
			"active":      x.ActiveSheet,
		}, nil

	case "multi_file":
		// um arquivo por dataframe dentro de um diretório
		info, err := os.Stat(x.FilePath)
		if err != nil {
			// se não existe, tenta criar
			if os.IsNotExist(err) {
				if err := os.MkdirAll(x.FilePath, os.ModePerm); err != nil {
					return nil, fmt.Errorf("failed to create output directory: %w", err)
				}
			} else {
				return nil, err
			}
		} else if !info.IsDir() {
			return nil, fmt.Errorf("path must be a directory")
		}

		exportedPaths := []string{}
		for name, data := range x.Data {
			tempFileName := filepath.Join(x.FilePath, name+".xlsx")

			if err := x.Open(); err != nil {
				return nil, fmt.Errorf("failed to open Excel file: %w", err)
			}
			index, err := x.createSheet(name, data)
			if err != nil {
				return nil, err
			}
			x.file.SetActiveSheet(index)

			// deleta Sheet1 se existir (API nova: (int, error))
			if idx, err := x.file.GetSheetIndex("Sheet1"); err == nil && idx != -1 {
				if err := x.file.DeleteSheet("Sheet1"); err != nil {
					return nil, fmt.Errorf("error deleting default Sheet1: %v", err)
				}
			}

			log.Printf("[XLSXWriter] Saving file to: %s", tempFileName)
			if err := x.file.SaveAs(tempFileName); err != nil {
				return nil, err
			}
			exportedPaths = append(exportedPaths, tempFileName)
		}

		return map[string]any{
			"mode":        "multi_file",
			"directory":   x.FilePath,
			"file_paths":  exportedPaths,
			"sheet_count": len(exportedPaths),
		}, nil

	default:
		return nil, fmt.Errorf("invalid save mode: %s", x.SaveMode)
	}
}

func (x *XLSXWriter) Close() error {
	// nada especial por enquanto
	return nil
}
