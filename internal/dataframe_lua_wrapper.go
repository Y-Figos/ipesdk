package internal

import (
	"log"

	"github.com/Y-Figos/ipesdk/core/df"
	"github.com/Y-Figos/ipesdk/utils"
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
		// normalize index to 1-based for Lua
		luaVal := utils.ConvertAnytoLuaType(L, v)
		table.RawSetInt(i+1, luaVal)
	}

	L.Push(table)
	return 1
}

func columnData(L *lua.LState) int {
	column := checkUserDataAs[df.ColumnInterface](L, 1)
	data := column.DataSlice()

	table := L.NewTable()
	for i, v := range data {
		L.RawSet(table, lua.LNumber(i+1), utils.ConvertAnytoLuaType(L, v))
	}
	L.Push(table)
	return 1

}

func columnGet(L *lua.LState) int {
	column := checkUserDataAs[df.ColumnInterface](L, 1)

	index := L.CheckInt(2)

	value := column.GetValue(index)
	L.Push(utils.ConvertAnytoLuaType(L, value))
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
	case "new_column":
		L.Push(L.NewFunction(dataframeNewColumn))
		return 1
	case "append":
		L.Push(L.NewFunction(dataframeAppend))
		return 1
	}

	// Then, check if key is a column name
	col, exists := df.Columns[key]
	if exists {
		// Wrap the column as Lua userdata
		colUD := L.NewUserData()
		colUD.Value = col
		L.SetMetatable(colUD, L.GetTypeMetatable("column"))
		L.Push(colUD)
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func dataframeNewIndex(L *lua.LState) int {
	dfUd := checkUserDataAs[*df.Dataframe](L, 1)
	key := L.CheckString(2)
	val := L.CheckUserData(3)

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

func dataframeNewColumn(L *lua.LState) int {
	mydf := checkUserDataAs[*df.Dataframe](L, 1)
	colName := L.CheckString(2)
	col := checkUserDataAs[df.ColumnInterface](L, 3)
	log.Println(col.DataSlice()...)
	mydf.NewColumn(colName, col)
	return 0
}
func dataframeAppend(L *lua.LState) int {
	mydf := checkUserDataAs[*df.Dataframe](L, 1)
	otherdf := checkUserDataAs[*df.Dataframe](L, 2)
	mydf.Append(otherdf)
	return 0
}

