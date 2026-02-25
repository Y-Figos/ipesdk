package adapters

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/Y-Figos/ipesdk/core/df"
	_ "github.com/Y-Figos/ipesdk/core/ports"
)

type BatchFileReader struct {
	ReaderFactory InputAdapterFactory
	DirPath       string
	FailedFiles   map[string]string
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
	bf.FailedFiles = make(map[string]string)
	var dfList []*df.Dataframe

	for _, file := range filesList {
		localArgs := maps.Clone(in_args)
		localArgs["filepath"] = file

		reader, err := bf.ReaderFactory(localArgs)
		if err != nil {
			bf.FailedFiles[file] = fmt.Sprintf("%v",err)
			continue
		}

		dataframe, err := reader.GetData()
		if err != nil {
			bf.FailedFiles[file] = fmt.Sprintf("%v",err) 
			continue
		}

		dfList = append(dfList, dataframe)
	}

	if len(dfList) == 0 {
		return nil, fmt.Errorf("no valid files read, failed: %v", bf.FailedFiles)
	}

	return dfList, nil
}

func (bf *BatchFileReader) MergeFiles(out_args map[string]any) (*df.Dataframe, error) {
    files, err := bf.getFileList()
    if err != nil {
        return nil, fmt.Errorf("getFileList failed: %w", err)
    }

    dfList, err := bf.readBatchFiles(files, out_args)
    if err != nil {
        return nil, fmt.Errorf("readBatchFiles failed: %w", err)
    }

    fmt.Printf("Processing %d files\n", len(dfList))
    
    // Log sizes of all dataframes before merge
    for i, df := range dfList {
        fmt.Printf("DataFrame[%d] has %d rows and %d columns\n", 
            i, df.RowCount(), len(df.ColumnOrder))
    }

    mainDataframe := dfList[0].Clone()
    fmt.Printf("After clone: mainDataframe has %d rows\n", mainDataframe.RowCount())
    
    // Append remaining dataframes with size tracking
    for i := 1; i < len(dfList); i++ {
        beforeRows := mainDataframe.RowCount()
        err := mainDataframe.Append(dfList[i])
        if err != nil {
            return nil, fmt.Errorf("failed to append dataframe %d: %w", i, err)
        }
        afterRows := mainDataframe.RowCount()
        expectedRows := beforeRows + dfList[i].RowCount()
        
        fmt.Printf("Merge step %d: before=%d + df[%d]=%d = expected:%d, actual:%d\n",
            i, beforeRows, i, dfList[i].RowCount(), expectedRows, afterRows)
            
        if afterRows != expectedRows {
            fmt.Printf("WARNING: Row count mismatch at step %d\n", i)
        }
    }

    err = mainDataframe.PadColumns()
    if err != nil {
        return nil, fmt.Errorf("PadColumns failed: %w", err)
    }

    fmt.Printf("Final dataframe: %d rows\n", mainDataframe.RowCount())
    return mainDataframe, nil
}

