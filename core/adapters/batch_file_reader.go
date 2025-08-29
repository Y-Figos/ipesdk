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
	FailedFiles   []string
}

func (bf *BatchFileReader) getFileList() ([]string, error) {
    path := filepath.Clean(bf.DirPath) // normalize slashes, remove trailing spaces
    info, err := os.Stat(path)
    if err != nil {
        return nil, fmt.Errorf("stat failed for %q: %w", path, err)
    }
    if !info.IsDir() {
        return nil, fmt.Errorf("path %q must be a directory, got file", path)
    }

    entries, err := os.ReadDir(path)
    if err != nil {
        return nil, fmt.Errorf("readdir failed for %q: %w", path, err)
    }

    files := make([]string, 0, len(entries))
    for _, entry := range entries {
        files = append(files, filepath.Join(path, entry.Name()))
    }
    return files, nil
}

func (bf *BatchFileReader) readBatchFiles(filesList []string, in_args map[string]any) ([]*df.Dataframe, error) {
	bf.FailedFiles = []string{}
	var dfList []*df.Dataframe // lista apenas com arquivos válidos

	for _, file := range filesList {
		in_args["filepath"] = file
		reader, err := bf.ReaderFactory(in_args)
		if err != nil {
			bf.FailedFiles = append(bf.FailedFiles, file) // só a path
			continue
		}

		dataframe, err := reader.GetData()
		if err != nil {
			bf.FailedFiles = append(bf.FailedFiles, file) // só a path
			continue
		}

		dfList = append(dfList, dataframe) // adiciona apenas os válidos
	}

	if len(dfList) == 0 {
		return nil, fmt.Errorf("no valid files read, failed: %v", bf.FailedFiles)
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

	for i := 1; i < len(dfList); i++ {
		mainDataframe.Append(dfList[i])
	}
	err = mainDataframe.PadColumns()
	if err != nil {
		return nil, err
	}
	return mainDataframe, nil
}
