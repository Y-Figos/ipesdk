# IPE SDK: Data Pipeline & Automation Engine

The **IPE SDK** is an extensible automation engine developed in **Go**, designed to manage complex data processing workflows and systemic tasks. Unlike linear scripts, IPE utilizes a **Directed Acyclic Graph (DAG)** structure to orchestrate the execution of interdependent tasks with high performance and safety.

## Technical Differentiators and Architecture

The project was built under **Platform Engineering** principles, focusing on decoupling, concurrent performance, and extensibility:

* **Native Concurrency:** The execution engine uses **goroutines** and **sync.WaitGroup** to process independent nodes in parallel. It automatically identifies execution layers in the graph, optimizing the total pipeline time.
* **Dependency Management (DAG):** Robust graph implementation for task orchestration. The engine performs topological sorting and cycle validation to ensure workflow integrity.
* **Extensibility via Lua:** Native support for Lua scripts through a `LuaManager`. This allows business logic or data transformations to be injected dynamically without the need to recompile the main binary.
* **Data Abstraction (Dataframes):** In-memory data manipulation layer inspired by Data Science tools, allowing complex transformations, filtering, and type inference efficiently.
* **Adapter Pattern:** Modular architecture with support for multiple input and output formats (CSV, XLSX, Custom Logs), facilitating data ingestion from legacy infrastructure systems.
* **Professional CLI:** Command-line interface built with **Cobra CLI**, following industry-standard tool patterns like `kubectl` and `docker`.

## Technology Stack

* **Main Language:** Go (Golang)
* **Scripting:** Lua (Gopher-Lua)
* **CLI Framework:** Cobra
* **Concurrency:** Channels and WaitGroups for flow control.

## Project Structure

* `/cmd`: CLI entry point and definition of execution and build commands.
* `/core/graph`: Graph engine, topological sorting logic, and concurrent execution.
* `/core/engine`: Runner responsible for the bridge between the Go core and Lua scripts.
* `/core/df`: DataFrame implementation and column type inference engines.
* `/core/adapters`: Modular connectors for different data sources (CSV, Excel, Logs).

## Technical Evolution Roadmap (v1.0)

This project consolidated the proof of concept (v0.1) of a robust automation engine based on DAGs. For the architecture of a production-ready version (v1.0), the following design decisions (ADRs) were mapped, focusing on **Performance**, **Extensibility**, and **Standardization**:

1.  **Transition to Native Nodes (Go-Only):** Discontinuation of the Lua script execution engine. All data transformation logic and processing nodes will become 100% native in Go. The goal is to eliminate the overhead of the language bridge, ensuring greater type-safety, execution speed, and debugging ease.
2.  **Plugin Architecture via gRPC:** Refactoring of the `adapters` layer. To avoid the need to recompile the core engine every time a new input/output format (CSV, Excel, APIs) is supported, adapters will operate as independent plugins communicating with the main engine via **gRPC** (inspired by the HashiCorp plugin model).
3.  **Adoption of Go-Gota (Dataframes):** Replacement of the custom DataFrame implementation (which relied heavily on the `reflect` package) with the `go-gota/gota` library. This change delegates structural data manipulation to the community-standard tool, eliminating runtime type conversion bottlenecks and optimizing memory.
4.  **Observability and Resilience (Context Propagation):** Implementation of structured logging (JSON standard), metrics export via Prometheus for monitoring graph layer execution times, and extensive use of `context.Context` end-to-end to ensure graceful shutdowns and timeout control.

---
*Developed as a Computer Science final year project (2025).*

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
