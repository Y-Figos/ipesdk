# IPE SDK Alpha Todo

* [X] Create Dataframe like type
  * [X] Create CSV Reader
  * [X] Implement Type Inference
  * [ ] String() Implementation
  * [X] filter
  * [X] map
  * [X] apply
  * [X] unique
  * [X] append
* [X] Create Basic input, output ports and adapters (XLSX, csv)
* [X] Implement Engine v1 (Run selected pipelines dynamically)
* [X] Implement pipeline v1 reader (detect lua files dynamically)
* [ ] Add config.json for non technical settings of pipeline
* [X] Implement .ipe builder
* [X] Implement .ipe Extractor
* [X] Create CLI for running pipelines

# Chores

* [X] Refactor CSV Reader Type Inference to be more reusable across other adapters





---

# DataFrame Operations

This document outlines the key operations you can perform with your custom `DataFrame` type, including `Filter`, `Apply`, and `Map`. These operations enable functional-style data transformation similar to those in Python's pandas, but written in Go.

---

## `Filter` Method

### Description

Filters the rows in a `DataFrame` based on a predicate function. Returns a new `DataFrame` with only the matching rows.

### Signature

```go
func (df *DataFrame) Filter(predicate func(row map[string]any) bool) *DataFrame
```

### Parameters

* `predicate`: Function that receives a row as a `map[string]any` and returns `true` to keep the row.

### Returns

* A new `DataFrame` with only the rows that satisfy the condition.

### Example

```go
// Keep only users over 18 with a verified email
adults := df.Filter(func(row map[string]any) bool {
	return row["age"].(int) > 18 && row["email_verified"].(bool)
})
```

---

## `Apply` Method

### Description

Transforms each value in a column using a custom function. Returns a new column.

### Signature

```go
func (col *Column) Apply(transform func(value any) any) *Column
```

### Parameters

* `transform`: A function that takes a value and returns a transformed value.

### Returns

* A new `Column` with transformed values.

### Example

```go
// Categorize age into groups
ageGroup := df.Columns["age"].Apply(func(val any) any {
	age := val.(int)
	switch {
	case age < 13:
		return "Child"
	case age < 20:
		return "Teen"
	case age < 65:
		return "Adult"
	default:
		return "Senior"
	}
})
df.NewColumn("age_group", ageGroup)
```

---

## `Map` Method

### Description

Replaces values in a column based on a lookup map. Unmatched values fall back to a default.

### Signature

```go
func (col *Column) Map(mapping map[any]any, defaultValue any) *Column
```

### Parameters

* `mapping`: A map of original values to new values.
* `defaultValue`: Value to use if the original value isn't in the map.

### Returns

* A new `Column` with mapped values.

### Example

```go
// Convert status codes into readable strings
statusLabels := df.Columns["status_code"].Map(
	map[any]any{
		1: "Active",
		2: "Inactive",
		3: "Banned",
	},
	"Unknown",
)
df.NewColumn("status_label", statusLabels)
```

---

## Full Example

```go
package main

import (
	"fmt"
	df "path/to/your/dataframe/package"
)

func main() {
	// Construct initial DataFrame manually
	users := &df.Dataframe{
		ColumnOrder: []string{"name", "age", "status_code", "email_verified"},
		Columns: map[string]df.ColumnInterface{
			"name": df.NewColumn("name", []string{"Alice", "Bob", "Eve", ""}),
			"age": df.NewColumn("age", []int{25, 17, 68, 30}),
			"status_code": df.NewColumn("status_code", []int{1, 2, 3, 99}),
			"email_verified": df.NewColumn("email_verified", []bool{true, false, true, true}),
		},
	}

	// Step 1: Filter valid users (name != "" and email verified)
	validUsers := users.Filter(func(row map[string]any) bool {
		return row["name"].(string) != "" && row["email_verified"].(bool)
	})

	// Step 2: Create "age_group" column using Apply
	ageGroup := validUsers.Columns["age"].Apply(func(val any) any {
		age := val.(int)
		switch {
		case age < 13:
			return "Child"
		case age < 20:
			return "Teen"
		case age < 65:
			return "Adult"
		default:
			return "Senior"
		}
	})
	validUsers.NewColumn("age_group", ageGroup)

	// Step 3: Map status_code to labels
	statusLabel := validUsers.Columns["status_code"].Map(
		map[any]any{
			1: "Active",
			2: "Inactive",
			3: "Banned",
		},
		"Unknown",
	)
	validUsers.NewColumn("status_label", statusLabel)

	// Step 4: Print result
	fmt.Println(validUsers.String())
}
```

---

## Suggested Usage Guidelines

Here are best practices when using the `DataFrame` package:

| Task                       | Recommended Tool                 |
| -------------------------- | -------------------------------- |
| Row filtering              | `df.Filter(...)`               |
| Conditional transformation | `col.Apply(...)`               |
| Lookup-based replacement   | `col.Map(...)`                 |
| New column creation        | `df.NewColumn("name", col)`    |
| Row inspection/debugging   | `df.String()`, `df.Row(i)`   |
| Data validation            | `col.Unique()`, `df.Shape()` |

---


