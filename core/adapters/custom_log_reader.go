package adapters

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Y-Figos/ipesdk/core/df"
)

// TODO
type LogReaderTim struct {
	Data           *df.Dataframe
	tableSlice     [][]string
	Filepath       string
	HeadersPattern string
	BaseInput      BaseInput
	siteName       string
}

func (lr *LogReaderTim) Open() error {
	file, err := os.Open(lr.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	fileData, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if len(fileData) == 0 {
		return fmt.Errorf("no File Data")
	}
	lines := strings.Split(string(fileData), "\n")
	sdiLines, err := lr.getSectionAfterPattern(lines, "sdi")
	if err != nil {
		return err
	}
	lr.siteName, err = lr.getSiteName(sdiLines)
	if err != nil {
		return err
	}
	cleanedText := lr.cleanText(sdiLines)
	RawTableLines, err := lr.getSectionAfterPattern(cleanedText, lr.HeadersPattern)
	if err != nil {
		return err
	}
	tableLines, err := lr.getTableLines(RawTableLines)
	if err != nil {
		return err
	}
	lr.tableSlice = lr.getTable(tableLines)
	return nil
}

func (lr *LogReaderTim) GetData() (*df.Dataframe, error) {
	err := lr.Open()
	if err != nil {
		return nil, err
	}
	headers, err := lr.GetHeaders()
	if err != nil {
		return nil, err
	}
	newDf := df.Dataframe{
		ColumnOrder: headers,
		Columns:     make(map[string]df.ColumnInterface),
	}
	lr.BaseInput.SetupSchemaFromSample(headers, lr.tableSlice[1:], &newDf)
	newDf.NewColumn("Site Name", lr.fillSiteValue(lr.siteName, newDf.RowCount()))
	return &newDf, nil
}

func (lr *LogReaderTim) getSiteName(lines []string) (string, error) {
	re, err := regexp.Compile(`((\w{2,}-\w{1,})|(\w{1,}))>`)
	if err != nil {
		return "", err
	}

	for _, line := range lines {
		if re.MatchString(line) {
			match := re.FindStringSubmatch(line)
			return match[1], nil
		}
	}
	return "", nil
}

func (lr *LogReaderTim) fillSiteValue(siteName string, rowCount int) df.ColumnInterface {
	dataSlice := []string{}
	for i := 0; i < rowCount; i++ {
		dataSlice = append(dataSlice, siteName)
	}

	return df.NewColumn("Site Name", dataSlice)
}

func (lr *LogReaderTim) ReadSample(n int) ([][]string, error) {
	return nil, nil
}
func (lr *LogReaderTim) GetHeaders() ([]string, error) {
	return lr.tableSlice[0], nil
}
func (lr *LogReaderTim) Close() error {
	return nil
}
func (lr *LogReaderTim) getSectionAfterPattern(lines []string, pattern string) ([]string, error) {
	re, err := regexp.Compile(regexp.QuoteMeta(pattern))
	if err != nil {
		return nil, err
	}

	for index, line := range lines {
		if re.MatchString(line) {
			return lines[index:], nil
		}
	}
	return nil, nil
}

func (lr *LogReaderTim) cleanText(lines []string) []string {
	cleanLines := []string{}
	for _, line := range lines {
		cleanTab := strings.ReplaceAll(line, " ", "")
		cleanLines = append(cleanLines, strings.ReplaceAll(cleanTab, "\t", ""))
	}
	return cleanLines
}

func (lr *LogReaderTim) getTableLines(lines []string) ([]string, error) {
	validTableSlice := []string{}
	for _, line := range lines {
		if strings.Contains(line, "===") || strings.Contains(line, "---") || strings.Contains(line, "q!!") {
			continue
		}
		if line == "" {
			continue
		}
		validTableSlice = append(validTableSlice, line)
	}
	return validTableSlice, nil
}

func (lr *LogReaderTim) getTable(lines []string) [][]string {
	table := [][]string{}
	for _, lines := range lines {
		table = append(table, strings.Split(lines, ";"))
	}
	return table
}
