package infer

import (
	"reflect"
	"strconv"
)

func InferTypeFromSlice(data []string) reflect.Type {
	intCount, floatCount, boolCount := 0, 0, 0
	totalCount := len(data)
	for _, val := range data {
		if _, err := strconv.Atoi(val); err == nil {
			intCount++
			continue
		}
		if _, err := strconv.ParseFloat(val, 64); err == nil {
			floatCount++
			continue
		}
		if _, err := strconv.ParseBool(val); err == nil {
			boolCount++
			continue
		}
	}

	switch {
	case intCount == totalCount:
		return reflect.TypeOf(0)
	case floatCount == totalCount || floatCount+intCount == totalCount:
		return reflect.TypeOf(0.0)
	case boolCount == totalCount:
		return reflect.TypeOf(true)
	default:
		return reflect.TypeOf("")
	}
}
