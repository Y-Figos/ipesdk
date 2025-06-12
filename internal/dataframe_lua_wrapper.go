package internal

import (
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/utils"
	lua "github.com/yuin/gopher-lua"
)


// Register function for Column Type
func RegisterColumnType(L *lua.LState) {
	mt := L.NewTypeMetatable("column")
	L.SetGlobal("column", mt)

	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":    columnGet,
		"data":   columnData,
		"len":    columnLen,
		"unique": columnUnique,
		"map":    columnMap,
		"apply":  columnApply,
	}))
}

// Column Type function Lua Mappers
func columnApply(L *lua.LState) int {
	column := checkUserDataAs[df.ColumnInterface](L, 1)
	fn := L.CheckFunction(2)

	newColumn, err := column.LuaApply(L, fn)
	if err != nil {
		L.RaiseError("apply error: %v", err)
	}
	ud := wrap(L, newColumn, "column")
	L.Push(ud)
	return 1
}

func columnMap(L *lua.LState) int {
	column := checkUserDataAs[df.ColumnInterface](L, 1)
	table := L.CheckTable(2)
	// mapParam := make(map[any]any)
	converted := utils.ConvertLuaTypeToGoType(table).(map[any]any)
	newcolumn := column.Map(converted)
	ud := wrap(L, newcolumn, "column")
	L.Push(ud)
	return 1
}

func columnUnique(L *lua.LState) int {
	column := checkUserDataAs[df.ColumnInterface](L, 1)

	data := column.Unique()
	table := L.NewTable()
	for i, v := range data {
		L.RawSet(table, lua.LNumber(i+1), utils.ConvertAnytoLuaType(v))
	}
	L.Push(table)
	return 1
}

func columnData(L *lua.LState) int {
	column := checkUserDataAs[df.ColumnInterface](L, 1)
	data := column.DataSlice()

	table := L.NewTable()
	for i, v := range data {
		L.RawSet(table, lua.LNumber(i+1), utils.ConvertAnytoLuaType(v))
	}
	L.Push(table)
	return 1

}

func columnGet(L *lua.LState) int {
	column := checkUserDataAs[df.ColumnInterface](L, 1)

	index := L.CheckInt(2)

	value := column.GetValue(index)
	L.Push(utils.ConvertAnytoLuaType(value))
	return 1
}

func columnLen(L *lua.LState) int {
	column := checkUserDataAs[df.ColumnInterface](L, 1)
	L.Push(lua.LNumber(column.Len()))
	return 1
}

// Register function for DataFrame Type
func RegisterDataFrameType(L *lua.LState) {
	mt := L.NewTypeMetatable("dataframe")
	L.SetGlobal("dataframe", mt)

	L.SetField(mt, "__index", L.NewFunction(dataframeIndex))
	L.SetField(mt, "__newindex", L.NewFunction(dataframeNewIndex))
}
// Dataframe Type function Lua Mappers
func dataframeShape(L *lua.LState) int {
	df := checkUserDataAs[*df.Dataframe](L, 1)
	cols, rows := df.Shape()

	L.Push(lua.LNumber(cols))
	L.Push(lua.LNumber(rows))
	return 2
}

func dataframeIndex(L *lua.LState) int {
	ud := L.CheckUserData(1) // the Dataframe userdata
	key := L.CheckString(2)  // the index key

	df, ok := ud.Value.(*df.Dataframe)
	if !ok {
		L.ArgError(1, "Expected Dataframe userdata")
		return 0
	}

	// First, check if key is a known method name
	switch key {
	case "shape":
		L.Push(L.NewFunction(dataframeShape))
		return 1
	case "filter":
		L.Push(L.NewFunction(dataframeFilter))
		return 1

	}

	// Then, check if key is a column name
	col, exists := df.Columns[key]
	if exists {
		// Wrap the column as Lua userdata
		colUD := L.NewUserData()
		colUD.Value = col
		L.SetMetatable(colUD, L.GetTypeMetatable("column")) // assuming you have a "column" metatable
		L.Push(colUD)
		return 1
	}

	// key not found, return nil
	L.Push(lua.LNil)
	return 1
}

func dataframeNewIndex(L *lua.LState) int {
	dfUd := checkUserDataAs[*df.Dataframe](L, 1)
	key := L.CheckString(2)
	val := L.CheckUserData(3) // your column wrapper

	// Optionally check val's type:
	col, ok := val.Value.(df.ColumnInterface)
	if !ok {
		L.RaiseError("expected ColumnInterface")
		return 0
	}

	// Save the column into the DataFrame's internal map
	dfUd.Columns[key] = col

	return 0
}

func dataframeFilter(L *lua.LState) int {
	df := checkUserDataAs[*df.Dataframe](L, 1)
	fn := L.CheckFunction(2)

	newDf, err := df.FilterLua(L, fn)
	if err != nil {
		L.RaiseError("filter failed: %v", err)
		return 0
	}
	ud := wrap(L, newDf, "dataframe")
	L.Push(ud)
	return 1
}

