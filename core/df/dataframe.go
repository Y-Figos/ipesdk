package df

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/Y-Figos/ipesdk/utils"
	lua "github.com/yuin/gopher-lua"
)

type ColumnInterface interface {
	HeaderName() string
	Type() reflect.Type
	Len() int
	DataSlice() []any
	AppendValue(any) error
	GetValue(int) any
	EmptyClone() ColumnInterface
	Map(map[any]any, ...string) ColumnInterface
	Unique() []any
	Apply(func(any) any, ...string) ColumnInterface
	LuaApply(*lua.LState, *lua.LFunction, ...string) (ColumnInterface, error)
}

type Column[T comparable] struct {
	Header string
	Data   []T
	GoType reflect.Type
}

func (c *Column[T]) EmptyClone() ColumnInterface {
	return NewColumn(c.Header, make([]T, 0))
}

func (c *Column[T]) GetValue(index int) any {
	return c.Data[index]
}

func (c Column[T]) HeaderName() string {
	return c.Header
}

func (c Column[T]) Type() reflect.Type {
	return c.GoType
}

func (c Column[T]) Len() int {
	return len(c.Data)
}

func (c Column[T]) DataSlice() []any {
	out := make([]any, len(c.Data))
	for i, v := range c.Data {
		out[i] = v
	}
	return out
}

func (c *Column[T]) AppendValue(val any) error {
	var target reflect.Type = c.GoType

	var finalValue any
	var err error

	switch target.Kind() {
	case reflect.Int:
		finalValue, err = convertToInt(val)
	case reflect.Float64:
		finalValue, err = convertToFloat(val)
	case reflect.Bool:
		finalValue, err = convertToBool(val)
	case reflect.String:
		finalValue, err = convertToString(val)
	case reflect.Interface:
		finalValue = val
	default:
		return fmt.Errorf("unsupported type: %v", target.Kind())
	}

	if err != nil {
		return fmt.Errorf("cannot convert %v (%T) to %v: %w", val, val, target, err)
	}

	
	converted, ok := finalValue.(T)
	if !ok {
		return fmt.Errorf("conversion succeeded but type assertion to T failed: %T", finalValue)
	}

	c.Data = append(c.Data, converted)
	return nil
}
func convertToInt(val any) (any, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to int", val)
	}
}

func convertToFloat(val any) (any, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return nil, fmt.Errorf("cannot convert %T to float64", val)
	}
}

func convertToBool(val any) (any, error) {
	switch v := val.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return nil, fmt.Errorf("cannot convert %T to bool", val)
	}
}

func convertToString(val any) (any, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func NewColumn[T comparable](header string, data []T) *Column[T] {
	return &Column[T]{
		Header: header,
		Data:   data,
		GoType: reflect.TypeOf((*T)(nil)).Elem(),
	}
}

func (c *Column[T]) Map(mapper map[any]any, optional_header ...string) ColumnInterface {
	var zero T
	header := c.Header + "_mapped"
	if len(optional_header) > 0 {
		header = optional_header[0]
	}

	newColumn := NewColumn(header, []any{})
	for _, data := range c.DataSlice() {
		if mapper[data] != nil {
			err := newColumn.AppendValue(mapper[data])
			if err != nil {
				log.Println(err)
			}
		} else {

			err := newColumn.AppendValue(zero)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return newColumn
}

func (c *Column[T]) Apply(predicate func(data any) any, optional_header ...string) ColumnInterface {
	header := "new_" + c.Header
	data := make([]any, len(c.Data))
	if len(optional_header) > 0 {
		header = optional_header[0]
	}
	for index, value := range c.Data {
		newValue := predicate(value)
		data[index] = newValue
	}
	return NewColumn(header, data)
}

func (c *Column[T]) LuaApply(L *lua.LState, predicate *lua.LFunction, optional_header ...string) (ColumnInterface, error) {
	data := make([]any, len(c.Data))
	header := "new_" + c.Header
	if len(optional_header) > 0 {
		header = optional_header[0]
	}
	for i, value := range c.Data {
		luaArg := utils.ConvertAnytoLuaType(L, value)
		err := L.CallByParam(lua.P{
			Fn:      predicate,
			NRet:    1,
			Protect: true,
		}, luaArg)
		if err != nil {
			return nil, fmt.Errorf("error applying function at row %d: %v", i, err)
		}
		result := L.Get(-1)
		L.Pop(1)

		data[i] = utils.ConvertLuaTypeToGoType(result)
	}
	return NewColumn(header, data), nil
}

func (c *Column[T]) Unique() []any {
	seen := make(map[T]struct{})
	var unique []any
	for _, value := range c.Data {
		if _, exist := seen[value]; !exist {
			seen[value] = struct{}{}
			unique = append(unique, value)
		}
	}
	return unique
}

func (c Column[T]) String() string {
	var column strings.Builder
	column.WriteString(fmt.Sprintf("\t%v\n", c.Header))

	for _, value := range c.Data {
		column.WriteString(fmt.Sprintf("\t%v\n", value))
	}

	return column.String()
}

type Dataframe struct {
	ColumnOrder []string
	Columns     map[string]ColumnInterface
}

func (df Dataframe) String() string {
	if len(df.Columns) == 0 {
		return "Empty DataFrame"
	}

	maxDisplayWidth := 10
	maxDisplayRows := 10
	maxDisplayColumns := 4

	colWidths := make(map[string]int, len(df.ColumnOrder))
	for _, colName := range df.ColumnOrder {
		maxLen := len(colName)
		for _, val := range df.Columns[colName].DataSlice() {
			strVal := fmt.Sprint(val)
			if len(strVal) > maxLen {
				maxLen = len(strVal)
			}
		}
		if maxLen > maxDisplayWidth {
			maxLen = maxDisplayWidth
		}
		colWidths[colName] = maxLen
	}

	displayCols := []string{}
	if len(df.ColumnOrder) <= maxDisplayColumns {
		displayCols = df.ColumnOrder
	} else {
		head := maxDisplayColumns / 2
		tail := maxDisplayColumns - head
		displayCols = append(displayCols, df.ColumnOrder[:head]...)
		displayCols = append(displayCols, "...")
		displayCols = append(displayCols, df.ColumnOrder[len(df.ColumnOrder)-tail:]...)
	}

	var out strings.Builder
	out.WriteString("\n")

	
	rowCount := df.RowCount()
	rowIdxWidth := len(fmt.Sprint(rowCount))

	out.WriteString(fmt.Sprintf("%*s  ", rowIdxWidth, "")) 
	for _, colName := range displayCols {
		if colName == "..." {
			out.WriteString("...  ")
			continue
		}
		width := colWidths[colName]
		out.WriteString(fmt.Sprintf("%-*s", width, truncate(colName, width)))
		out.WriteString("  ")
	}
	out.WriteString("\n")

	
	rowsToShow := []int{}
	if rowCount <= maxDisplayRows {
		for i := 0; i < rowCount; i++ {
			rowsToShow = append(rowsToShow, i)
		}
	} else {
		head := maxDisplayRows / 2
		tail := maxDisplayRows - head
		for i := 0; i < head; i++ {
			rowsToShow = append(rowsToShow, i)
		}
		rowsToShow = append(rowsToShow, -1) 
		for i := rowCount - tail; i < rowCount; i++ {
			rowsToShow = append(rowsToShow, i)
		}
	}

	
	for _, rowIdx := range rowsToShow {
		if rowIdx == -1 {
			out.WriteString(fmt.Sprintf("%*s  ...\n", rowIdxWidth, ""))
			continue
		}

		out.WriteString(fmt.Sprintf("%*d  ", rowIdxWidth, rowIdx))
		for _, colName := range displayCols {
			if colName == "..." {
				out.WriteString("...  ")
				continue
			}

			width := colWidths[colName]
			col := df.Columns[colName]

			var val string
			if rowIdx < col.Len() {
				val = fmt.Sprint(col.GetValue(rowIdx))
			} else {
				val = ""
			}

			out.WriteString(fmt.Sprintf("%-*s", width, truncate(val, width)))
			out.WriteString("  ")
		}
		out.WriteString("\n")
	}

	return out.String()
}

func truncate(s string, max int) string {
	if len(s) > max {
		if max <= 4 {
			return s[:max]
		}
		return s[:max-3] + "..."
	}
	return s
}

func (df *Dataframe) Shape() (int, int) {
	columns := df.ColumnCount()
	rows := df.RowCount()
	return columns, rows
}

func (df *Dataframe) ColumnCount() int {
	columns := len(df.Columns)
	return columns
}

func (df *Dataframe) RowCount() int {
	if len(df.Columns) == 0 {
		return 0
	}

	var maxiRowColumn int = 0
	for _, col := range df.Columns {
		x := col.Len()
		if x > maxiRowColumn {
			maxiRowColumn = x
		}
	}
	return maxiRowColumn
}

func (df *Dataframe) Row(i int) map[string]any {
	row := make(map[string]any, len(df.Columns))
	for _, col := range df.ColumnOrder {
		row[df.Columns[col].HeaderName()] = df.Columns[col].DataSlice()[i]
	}
	return row
}

func (df *Dataframe) RowSliceAny(i int) []any {
	row := make([]any, 0, len(df.Columns))
	for _, col := range df.ColumnOrder {
		row = append(row, df.Columns[col].DataSlice()[i])
	}
	return row
}
func (df *Dataframe) RowSlice(i int) []string {
	row := make([]string, 0, len(df.Columns))
	for _, col := range df.ColumnOrder {
		data := df.Columns[col].DataSlice()
		if i < len(data) {
			row = append(row, fmt.Sprint(data[i]))
		} else {
			row = append(row, "")
		}
	}
	return row
}

func (df *Dataframe) Filter(predicate func(map[string]any) bool) *Dataframe {

	newCols := make(map[string]ColumnInterface, len(df.Columns))

	for name, col := range df.Columns {
		newCols[name] = col.EmptyClone()
	}

	rowCount := df.RowCount()
	row := make(map[string]any, len(df.Columns))

	for i := 0; i < rowCount; i++ {
		
		for name, col := range df.Columns {
			row[name] = col.GetValue(i)
		}

		if predicate(row) {
			for name, col := range newCols {
				col.AppendValue(row[name])
			}
		}
	}
	return &Dataframe{
		ColumnOrder: df.ColumnOrder,
		Columns:     newCols,
	}
}

func (df *Dataframe) FilterLua(L *lua.LState, fn *lua.LFunction) (*Dataframe, error) {
	newCols := make(map[string]ColumnInterface, len(df.Columns))
	for name, col := range df.Columns {
		newCols[name] = col.EmptyClone()
	}

	rowCount := df.RowCount()

	for i := 0; i < rowCount; i++ {
		luaRow := L.NewTable()

		
		for name, col := range df.Columns {
			goVal := col.GetValue(i)
			luaVal := utils.ConvertAnytoLuaType(L, goVal)
			L.SetField(luaRow, name, luaVal)
		}

		err := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    1,
			Protect: true,
		}, luaRow)
		if err != nil {
			return nil, fmt.Errorf("error in filter predicate at row %d: %v", i, err)
		}

		ret := L.Get(-1)
		L.Pop(1)

		if lua.LVAsBool(ret) {
			for name, col := range newCols {
				col.AppendValue(df.Columns[name].GetValue(i))
			}
		}
	}

	return &Dataframe{
		ColumnOrder: df.ColumnOrder,
		Columns:     newCols,
	}, nil
}

func (df *Dataframe) Append(otherDf *Dataframe) error {
	if !reflect.DeepEqual(df.ColumnOrder, otherDf.ColumnOrder) {
		return errors.New("dataframes must be equal to be appended")
	}
	for header, column := range df.Columns {
		for _, value := range otherDf.Columns[header].DataSlice() {
			column.AppendValue(value)
		}
	}
	err := df.PadColumns()
	if err != nil {
		return err
	}
	return nil
}

func (df *Dataframe) NewColumn(columnName string, newColumn ColumnInterface) {
	df.ColumnOrder = append(df.ColumnOrder, columnName)
	df.Columns[columnName] = newColumn
}

func (df *Dataframe) PadColumns() error {
	if len(df.Columns) == 0 {
		return nil
	}

	
	maxRows := 0
	for _, col := range df.Columns {
		if col.Len() > maxRows {
			maxRows = col.Len()
		}
	}

	
	for name, col := range df.Columns {
		currentLen := col.Len()
		for i := currentLen; i < maxRows; i++ {
			var padValue any
			switch col.Type().Kind() {
			case reflect.Int:
				padValue = 0
			case reflect.Float64:
				padValue = 0.0
			case reflect.Bool:
				padValue = false
			case reflect.String:
				padValue = ""
			default:
				padValue = nil
			}
			if err := col.AppendValue(padValue); err != nil {
				return fmt.Errorf("failed to pad column %s: %v", name, err)
			}
		}
	}

	return nil
}
func (df *Dataframe) EmptyClone() *Dataframe {
    newColumns := make(map[string]ColumnInterface, len(df.Columns))
    
    
    for name, col := range df.Columns {
        newColumns[name] = col.EmptyClone()
    }
    
    
    return &Dataframe{
        ColumnOrder: append([]string{}, df.ColumnOrder...),
        Columns:     newColumns,
    }
}

func (df *Dataframe) Clone() *Dataframe {
    newColumns := make(map[string]ColumnInterface, len(df.Columns))
    
    
    for name, col := range df.Columns {
        newCol := col.EmptyClone()
        for _, val := range col.DataSlice() {
            newCol.AppendValue(val)
        }
        newColumns[name] = newCol
    }
    
    return &Dataframe{
        ColumnOrder: append([]string{}, df.ColumnOrder...),
        Columns:     newColumns,
    }

}
