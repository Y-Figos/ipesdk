package df

import (
	"fmt"
	"reflect"
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
}

type Column[T any] struct{
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
	converted, ok := val.(T)
	if !ok {
		return fmt.Errorf("cannot convert %v (%T) to %v", val, val, c.GoType)
	}
	c.Data = append(c.Data, converted)
	return nil
}

func NewColumn[T any](header string, data []T) *Column[T]{
	return &Column[T]{
		Header: header,
		Data: data,
		GoType:  reflect.TypeOf((*T)(nil)).Elem(),
	}
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

 


