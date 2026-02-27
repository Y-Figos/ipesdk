# IPE SDK: Advanced Data Pipeline & Automation Engine

O **IPE SDK** é uma engine de automação extensível desenvolvida em **Go**, projetada para gerenciar fluxos complexos de processamento de dados e tarefas sistêmicas. Diferente de scripts lineares, o IPE utiliza uma estrutura de **Grafo Acíclico Dirigido (DAG)** para orquestrar a execução de tarefas interdependentes com alta performance e segurança.



## Diferenciais Técnicos e Arquitetura

O projeto foi construído sob princípios de **Engenharia de Plataforma**, focando em desacoplamento, performance concorrente e extensibilidade:

* **Native Concurrency:** O motor de execução utiliza **goroutines** e **sync.WaitGroup** para processar nós independentes de forma paralela. Ele identifica automaticamente camadas de execução no grafo, otimizando o tempo total do pipeline.
* **Dependency Management (DAG):** Implementação robusta de grafos para orquestração de tarefas. O motor realiza a ordenação topológica e validação de ciclos para garantir a integridade do fluxo de trabalho.
* **Extensibilidade via Lua:** Suporte nativo para scripts Lua através de um `LuaManager`. Isso permite que a lógica de negócio ou transformações de dados sejam injetadas dinamicamente sem a necessidade de recompilar o binário principal.
* **Data Abstraction (Dataframes):** Camada de manipulação de dados em memória inspirada em ferramentas de Data Science, permitindo transformações complexas, filtragens e inferência de tipos de forma eficiente.
* **Adapter Pattern:** Arquitetura modular com suporte a múltiplos formatos de entrada e saída (CSV, XLSX, Custom Logs), facilitando a ingestão de dados de sistemas legados de infraestrutura.
* **Professional CLI:** Interface de linha de comando construída com **Cobra CLI**, seguindo os padrões de ferramentas *industry-standard* como `kubectl` e `docker`.

## Stack Tecnológica

* **Linguagem Principal:** Go (Golang)
* **Scripting:** Lua (Gopher-Lua)
* **CLI Framework:** Cobra
* **Concorrência:** Canais (Channels) e WaitGroups para controle de fluxo.

## Estrutura do Projeto

* `/cmd`: Ponto de entrada da CLI e definição de comandos de execução e build.
* `/core/graph`: Motor do Grafo, lógica de ordenação topológica e execução concorrente.
* `/core/engine`: Runner responsável pela ponte entre o core em Go e os scripts Lua.
* `/core/df`: Implementação de Dataframes e motores de inferência de tipos de colunas.
* `/core/adapters`: Conectores modulares para diferentes fontes de dados (CSV, Excel, Logs).

## Roadmap de Evolução Técnica (v1.0)

Este projeto consolidou a prova de conceito (v0.1) de um motor de automação robusto baseado em DAGs. Para a arquitetura de uma versão voltada para produção (v1.0), as seguintes decisões de design (ADRs) foram mapeadas, focando em **Performance**, **Extensibilidade** e **Padronização**:

1.  **Transição para Nodes Nativos (Go-Only):** Descontinuação do motor de execução de scripts Lua. Toda a lógica de transformação de dados e nós de processamento passará a ser 100% nativa em Go. O objetivo é eliminar o *overhead* da ponte entre linguagens, garantindo maior *type-safety*, velocidade de execução e facilidade de *debug*.
2.  **Arquitetura de Plugins via gRPC:** Refatoração da camada de `adapters`. Para evitar a necessidade de recompilar o *core engine* toda vez que um novo formato de entrada/saída (CSV, Excel, APIs) for suportado, os adaptadores operarão como plugins independentes comunicando-se com o motor principal via **gRPC** (inspirado no modelo de plugins da HashiCorp).
3.  **Adoção do Go-Gota (Dataframes):** Substituição da implementação customizada de Dataframes (que dependia fortemente do pacote `reflect`) pela biblioteca `go-gota/gota`. Essa mudança delega a manipulação estrutural de dados para a ferramenta padrão da comunidade, eliminando gargalos de conversão de tipos em tempo de execução e otimizando a memória.
4.  **Observabilidade e Resiliência (Context Propagation):** Implementação de logs estruturados (padrão JSON), exportação de métricas via Prometheus para monitoramento do tempo de execução das camadas do Grafo, e uso extensivo de `context.Context` ponta a ponta para garantir cancelamentos graciosos (*graceful shutdown*) e controle de *timeouts*.

---
*Desenvolvido como projeto de conclusão de curso em Ciência da Computação (2025).*

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
