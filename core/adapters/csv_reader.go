package adapters

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	_ "reflect"
	"errors"
	"github.com/Y-Figos/ipesdk/core/df"
)

type CSVReader struct {
	FilePath    string
	File        *os.File
	Reader      *csv.Reader
	BatchSize   int
	WorkerCount int
	SkipRow     int
	BaseInput   *BaseInput
}

func NewCSVReaderWithOptions(filePath string, batchSize int, workerCount int) *CSVReader {
	if batchSize <= 0 {
		batchSize = 100
	}
	if workerCount <= 0 {
		workerCount = 4
	}
	return &CSVReader{
		BaseInput: &BaseInput{},
		FilePath:    filePath,
		BatchSize:   batchSize,
		WorkerCount: workerCount,
	}
}

func (c *CSVReader) Open() error {
	file, err := os.Open(c.FilePath)
	if err != nil {
		return err
	}
	c.File = file
	c.Reader = csv.NewReader(c.File)
	return nil
}

func (c *CSVReader) Close() error {
	if c.File != nil {
		err := c.File.Close()
		c.File = nil
		return err
	}
	return nil
}

func (c *CSVReader) GetHeaders() ([]string, error) {
	headers, err := c.Reader.Read()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return headers, nil
}
func (c *CSVReader) ReadSample(n int) ([][]string, error) {
	sampleData := make([][]string, n)
	for i := 0; i < n; i++ {
		records, err := c.Reader.Read()
		if err != nil {
			return nil, err
		}
		sampleData[i] = records
	}
	return sampleData, nil
}

func (c *CSVReader) ReadBatch(n int) ([][]string, error) {
	var batch [][]string

	for i := 0; i < n; i++ {
		record, err := c.Reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("read error: %v", err)
			return nil, err
		}
		batch = append(batch, record)
	}

	return batch, nil

}

func (c *CSVReader) GetData() (*df.Dataframe, error) {
	
	err := c.Open()
	if err != nil {
		return nil, err
	}
	samplesize := 100
	batchChannel := make(chan [][]string, 100)
	defer c.Close()
	if c.BaseInput == nil {
		return nil, errors.New("BaseInput is not initialized")
	}
	headers, err := c.GetHeaders()
	if err != nil {
		c.Close()
		log.Println(err)
		return nil, err
	}

	newDf := df.Dataframe{
		ColumnOrder: headers,
		Columns:     make(map[string]df.ColumnInterface),
	}

	sampleData, err := c.ReadSample(samplesize)
	if err != nil {
		return nil, err
	}

	//Create BaseInput and create TyperInferer method
	c.BaseInput.SetupSchemaFromSample(headers, sampleData, &newDf)

	//Create BaseInput and create start Workers method
	wg := c.BaseInput.StartWorkers(batchChannel, headers, &newDf, c.WorkerCount)

	// Read and send batches
	for {
		batch, err := c.ReadBatch(c.BatchSize)
		if err != nil {
			return nil, err
		}
		if len(batch) > 0 {
			batchChannel <- batch
		}
		if len(batch) == 0 {
			break
		}
	}

	close(batchChannel)
	wg.Wait()
	return &newDf, nil
}
