package adapters

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Y-Figos/ipesdk/core/df"
)

type LogReader struct {
	Data       *df.Dataframe
	tableSlice [][]string
	Filepath   string
	BaseInput  BaseInput
	siteName   string
}

func (lr *LogReader) Open() error {
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

	alarmLines, err := lr.getSectionAfterPattern(lines, "alt", `Total: \d+ MOs`)
	if err != nil {
		return err
	}
	lr.siteName, err = lr.getSiteName(alarmLines)
	if err != nil {
		return err
	}

	err = lr.checkAlarms(alarmLines)
	if err != nil {
		return fmt.Errorf("0 alarms triggered, site Ok")
	}

	cleanedText := lr.cleanTextAlarms(alarmLines)
	RawTableLinesAlarm, err := lr.getSectionAfterPattern(cleanedText, "Date & Time (Local) S Specific Problem;MO (AdditionalText/PC, AI)", `>>> Total: (?P<AlarmCount>\d+)`)
	if err != nil {
		return err
	}

	RawTableLinesSector, err := lr.getSectionAfterPattern(cleanedText, "MO;Attribute;Value", `Total: \d+ MOs`)
	if err != nil {
		return err
	}

	tableLinesAlarm, err := lr.getTableLines(RawTableLinesAlarm)
	if err != nil {
		return err
	}
	tableLinesSector, err := lr.getTableLines(RawTableLinesSector)
	if err != nil {
		return err
	}
	tableSliceAlarm := lr.getTable(tableLinesAlarm)
	tableSliceSector := lr.getTable(tableLinesSector)
	lr.tableSlice = lr.mergeTablesSideBySide(tableSliceAlarm, tableSliceSector)
	return nil
}

func (lr *LogReader) GetData() (*df.Dataframe, error) {

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
	newDf.NewColumn("Users", newDf.Columns["Value"])
	newDf.ColumnOrder = []string{"Site Name", "Date & Time (Local) S Specific Problem", "MO (AdditionalText/PC, AI)", "MO", "Attribute", "Users"}
	return &newDf, nil
}

func (lr *LogReader) getSiteName(lines []string) (string, error) {
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

func (lr *LogReader) fillSiteValue(siteName string, rowCount int) df.ColumnInterface {
	dataSlice := []string{}
	for i := 0; i < rowCount; i++ {
		dataSlice = append(dataSlice, siteName)
	}

	return df.NewColumn("Site Name", dataSlice)
}

func (lr *LogReader) ReadSample(n int) ([][]string, error) {
	return nil, nil
}
func (lr *LogReader) GetHeaders() ([]string, error) {
	if len(lr.tableSlice) == 0 {
		return nil, fmt.Errorf("error at file %v, no header found", lr.Filepath)
	}
	return lr.tableSlice[0], nil
}
func (lr *LogReader) Close() error {
	return nil
}
func (lr *LogReader) getSectionAfterPattern(lines []string, startPattern, endPattern string) ([]string, error) {
	startre, err := regexp.Compile(regexp.QuoteMeta(startPattern))
	if err != nil {
		return nil, err
	}
	endre, err := regexp.Compile(endPattern)
	if err != nil {
		return nil, err
	}
	startIndex := -1
	endIndex := -1

	for index, line := range lines {
		if startIndex == -1 && startre.MatchString(line) {
			startIndex = index
		} else if startIndex != -1 && endre.MatchString(line) {
			endIndex = index
			break
		}
	}

	if startIndex == -1 {
		return nil, fmt.Errorf("start pattern not found")
	}
	if endIndex == -1 {
		return nil, fmt.Errorf("end pattern not found after start")
	}
	return lines[startIndex : endIndex+1], nil
}

func (lr *LogReader) cleanTextAlarms(lines []string) []string {
	cleanLines := []string{}
	re := regexp.MustCompile(`\s{2,}`)
	for _, line := range lines {
		cleanTab := re.ReplaceAllString(line, ";")
		cleanLines = append(cleanLines, cleanTab)
	}
	return cleanLines
}

func (lr *LogReader) getTableLines(lines []string) ([]string, error) {
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

func (lr *LogReader) getTable(lines []string) [][]string {
	table := [][]string{}
	for _, lines := range lines {
		table = append(table, strings.Split(lines, ";"))
	}
	return table
}

func (lr *LogReader) checkAlarms(lines []string) error {
	pattern := `>>> Total: (?P<AlarmCount>\d+)`
	re := regexp.MustCompile(pattern)

	for _, line := range lines {
		match := re.FindStringSubmatch(line)
		result := make(map[string]string)
		if match != nil {
			for i, name := range re.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}
			alarmCount, err := strconv.Atoi(result["AlarmCount"])
			if alarmCount <= 0 || err != nil {

				return fmt.Errorf("0 alarms, site ok")
			}
		}

	}
	return nil
}
func (lr *LogReader) mergeTablesSideBySide(left, right [][]string) [][]string {
	// Helper to get max width of a ragged table
	maxWidth := func(t [][]string) int {
		w := 0
		for _, r := range t {
			if len(r) > w {
				w = len(r)
			}
		}
		return w
	}
	// Pad a row to a fixed width (copy-safe)
	pad := func(row []string, width int) []string {
		out := make([]string, width)
		copy(out, row)
		return out
	}

	leftRows, rightRows := len(left), len(right)
	leftWidth, rightWidth := maxWidth(left), maxWidth(right)

	maxRows := leftRows
	if rightRows > maxRows {
		maxRows = rightRows
	}

	result := make([][]string, 0, maxRows)

	for i := 0; i < maxRows; i++ {
		var lrow, rrow []string
		if i < leftRows {
			lrow = pad(left[i], leftWidth)
		} else {
			lrow = make([]string, leftWidth)
		}
		if i < rightRows {
			rrow = pad(right[i], rightWidth)
		} else {
			rrow = make([]string, rightWidth)
		}

		// Concatenate fixed-width halves; columns won't shift
		row := append(lrow, rrow...)
		result = append(result, row)
	}

	return result
}
