package df
import(
	"strings"
	"fmt"
	"reflect"
)

type ColumnInterface interface{
	HeaderName() string
	Type() reflect.Type
	Len() int
	DataSlice() []any
	AppendValue(any) error
}

type Column[T any] struct{
	Header string
	Data []T
	GoType reflect.Type
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

func NewColumn[T any](header string, data []T) Column[T]{
	return Column[T]{
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

func(df *Dataframe) RowCount() int {
	var maxRow int
	for _, column := range df.Columns{
		if len(column.DataSlice()) > maxRow {
			maxRow = len(column.DataSlice())
		}
	}
	return maxRow
}

func (df *Dataframe) Row(i int) map[string]any{
	row := make(map[string]any, len(df.Columns))
	for _, col := range df.Columns {
		row[col.HeaderName()] = col.DataSlice()[i]
	}
	return row
}



