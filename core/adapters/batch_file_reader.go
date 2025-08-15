package adapters

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Y-Figos/ipesdk/core/df"
	_ "github.com/Y-Figos/ipesdk/core/ports"
)

type BatchFileReader struct {
	ReaderFactory InputAdapterFactory
	DirPath       string
}

func (bf *BatchFileReader) getFileList() ([]string, error) {
	files := []string{}
	info, err := os.Stat(bf.DirPath)
	if err != nil {
		return nil, err // path might not exist
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path must be a directory")
	}
	dirEntry, err := os.ReadDir(bf.DirPath)
	if err != nil {
		return nil, err
	}
	for _, entry := range dirEntry {
		files = append(files, filepath.Join(bf.DirPath, entry.Name()))
	}
	return files, nil
}

func (bf *BatchFileReader) readBatchFiles(filesList []string, out_args map[string]any) ([]*df.Dataframe, error) {
	dfList := make([]*df.Dataframe, len(filesList))
	for index, file := range filesList {
		out_args["filepath"] = file
		reader, err := bf.ReaderFactory(out_args)
		if err != nil {
			return nil, err
		}
		dataframe, err := reader.GetData()
		if err != nil {
			return nil, err
		}
		dfList[index] = dataframe
	}
	return dfList, nil
}

func (bf *BatchFileReader) MergeFiles(out_args map[string]any) (*df.Dataframe, error) {
	files, err := bf.getFileList()
	if err != nil {
		return nil, err
	}
	dfList, err := bf.readBatchFiles(files, out_args)
	if err != nil {
		return nil, err
	}
	mainDataframe := dfList[0] // first dataframe is the main one

	for i := 1; i > len(dfList); i++ {
		mainDataframe.Append(dfList[i])
	}
	return mainDataframe, nil
}
