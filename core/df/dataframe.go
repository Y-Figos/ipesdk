package df

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/Y-Figos/ipesdk/utils"
	"github.com/yuin/gopher-lua"
)

type ColumnInterface interface{
	HeaderName() string
	Type() reflect.Type
	Len() int
	DataSlice() []any
	AppendValue(any) error
	GetValue(int) any
	EmptyClone() ColumnInterface 
	Map(map[any]any, ...string) ColumnInterface
	Unique() []any 
	Apply(func(any) any, ...string ) ColumnInterface
	LuaApply( *lua.LState, *lua.LFunction, ...string ) (ColumnInterface, error)
}

type Column[T comparable] struct{
	Header 		string
	Data		[]T
	GoType 		reflect.Type
}

func (c *Column[T]) EmptyClone() ColumnInterface{
	return NewColumn(c.Header, make([]T, 0))
}

func (c *Column[T]) GetValue(index int) any {
	return c.Data[index]
}

func (c Column[T]) HeaderName() string{
	return c.Header
}

func (c Column[T]) Type() reflect.Type{
	return c.GoType
}

func (c Column[T]) Len() int{
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

	// Convert finalValue (any) into T using reflection
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

func NewColumn[T comparable](header string, data []T) *Column[T]{
	return &Column[T]{
		Header: header,
		Data: data,
		GoType:  reflect.TypeOf((*T)(nil)).Elem(),
	}
}

func (c *Column[T]) Map(mapper map[any]any, optional_header ...string) ColumnInterface{
	var zero T
	header := c.Header + "_mapped"
	if len(optional_header) > 0{
		header = optional_header[0]
	}

	newColumn := NewColumn(header, []any{})
		for _ ,data := range c.DataSlice() {
			if mapper[data] != nil{
				err := newColumn.AppendValue(mapper[data])
				if err != nil{
					log.Println(err)
				}
			}else {

				err := newColumn.AppendValue(zero)
				if err != nil{
					log.Println(err)
				}
			}
		}
	return newColumn
}

func (c* Column[T]) Apply(predicate func(data any) any, optional_header ...string) ColumnInterface{
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

func (c* Column[T]) LuaApply(L *lua.LState, predicate *lua.LFunction, optional_header ...string) (ColumnInterface, error){
	data := make([]any, len(c.Data))
	header := "new_" + c.Header
	if len(optional_header) > 0 {
		header = optional_header[0]
	}
	for i, value := range c.Data{
		luaArg := utils.ConvertAnytoLuaType(value)
		err := L.CallByParam(lua.P{
			Fn: predicate,
			NRet: 1,
			Protect: true,
		}, luaArg)
		if err != nil{
			return nil, fmt.Errorf("error applying function at row %d: %v", i, err)
		}
		result := L.Get(-1)
		L.Pop(1)

		data[i] = utils.ConvertLuaTypeToGoType(result)
	}
	return NewColumn(header, data), nil
}

func (c *Column[T]) Unique() []any{
	seen :=  make(map[T]struct{})
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

	for _, value := range c.Data{
		column.WriteString(fmt.Sprintf("\t%v\n", value))
	}

	return column.String()
}

type Dataframe struct {
	ColumnOrder []string
	Columns map[string]ColumnInterface
}

func (df Dataframe) String() string {
	var dfString strings.Builder

	// Adjust the width according to your data
	const colWidth = 10

	// Headers
	for _, col := range df.Columns {
		dfString.WriteString(fmt.Sprintf("%-*v|", colWidth, col.HeaderName()))
	}
	dfString.WriteString("\n")

	// Rows
	for i := 0; i < df.RowCount(); i++ {
		for _, value := range df.Row(i) {
			dfString.WriteString(fmt.Sprintf("%-*v|", colWidth, value))
		}
		dfString.WriteString("\n")
	}

	return dfString.String()
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
    // All columns same length - pick first
    for _, col := range df.Columns {
        return col.Len()
    }
    return 0
}

func (df *Dataframe) Row(i int) map[string]any{
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
			row = append(row, "") // or some default value
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
        // Build row without copying full columns
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

		// Construct row table for Lua
		for name, col := range df.Columns {
			goVal := col.GetValue(i)
			luaVal := utils.ConvertAnytoLuaType(goVal)
			L.SetField(luaRow, name, luaVal)
		}

		// Call the Lua function with row table
		err := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    1,
			Protect: true,
		}, luaRow)
		if err != nil {
			return nil, fmt.Errorf("error in filter predicate at row %d: %v", i, err)
		}

		// Get and evaluate return value
		ret := L.Get(-1)
		L.Pop(1) // pop result from stack

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
	if !reflect.DeepEqual(df.ColumnOrder, otherDf.ColumnOrder){
		return errors.New("dataframes must be equal to be appended")
	}
	for header, column := range df.Columns {
		for _, value  := range otherDf.Columns[header].DataSlice(){
			column.AppendValue(value)
		}
	}
	return nil
}

func (df *Dataframe) NewColumn(columnName string, newColumn ColumnInterface){
	df.ColumnOrder = append(df.ColumnOrder, columnName)
	df.Columns[columnName] = newColumn
}


