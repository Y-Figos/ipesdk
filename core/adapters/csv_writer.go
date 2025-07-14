package adapters

import (
	"encoding/csv"
	"errors"
	"os"

	"github.com/Y-Figos/ipesdk/core/df"
)

type CSVWriter struct {
	Data 		*df.Dataframe
	FilePath    string
	File        *os.File
	Writer		*csv.Writer
}
func (c *CSVWriter) Open() error {
	file, err := os.Create(c.FilePath)
	if err != nil{
		return err
	}
	c.File = file
	c.Writer = csv.NewWriter(c.File)
	return nil
}
func (c *CSVWriter) Close() error {
	if c.File == nil{
		return errors.New("CSVWriter.Close(): file does not exist")
	}
	if err :=c.File.Close(); err != nil{
		return err
	}
	return nil
	
}

func (c *CSVWriter) WriteData() error{
	_, rows := c.Data.Shape()
	for i:=0; i < rows; i++ {
		if err := c.Writer.Write(c.Data.RowSlice(i)); err !=nil{
			return err
		}
	}
	return nil
}
func (c *CSVWriter) WriteHeaders() error{
	if err := c.Writer.Write(c.Data.ColumnOrder); err != nil{
		return err
	}
	return nil
}

func (c *CSVWriter) ExportData() error{

	if c.FilePath == ""{
		return errors.New("CSVWriter.ExportData(): Must specify Path")
	}
	err := c.Open()
	if err != nil{
		return err
	}
	defer c.Close()

	defer c.Writer.Flush()
	err = c.WriteHeaders()
	if err != nil{
		return err
	}
	err = c.WriteData()
	if err != nil{
		return err
	}
		
	return nil
}