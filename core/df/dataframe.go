package df

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
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
}

type Column[T comparable] struct{
	Header string
	Data []T
	GoType reflect.Type
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
					log.Fatalln(err)
				}
			}else {

				err := newColumn.AppendValue(zero)
				if err != nil{
					log.Fatalln(err)
				}
			}
		}
	return newColumn
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
	for _, col := range df.Columns {
		row[col.HeaderName()] = col.DataSlice()[i]
	}
	return row
}

func (df *Dataframe) RowView(i int) RowView {
    return RowView{
        Cols:  df.Columns,
        Index: i,
    }
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

type RowView struct {
    Cols  map[string]ColumnInterface
    Index int
}

func (r RowView) Get(col string) any {
    return r.Cols[col].GetValue(r.Index)
}




