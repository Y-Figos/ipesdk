package infer

import (
	"reflect"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
)

func CreateTypedColumn(header string,columnType reflect.Type) df.ColumnInterface {
	switch {
		case columnType == reflect.TypeOf(0):
			return df.NewColumn(header, []int{})
		case columnType == reflect.TypeOf(0.0):
			return df.NewColumn(header, []float64{})
		case columnType == reflect.TypeOf(true):
			return df.NewColumn(header, []bool{})
		default:
			return df.NewColumn(header, []string{})
	}
}