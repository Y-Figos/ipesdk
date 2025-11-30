# README.dev.md — IPE SDK Developer Guide

This document describes the full flow for creating, configuring, building, installing, parameterizing, and running modules in the **IPE SDK**, with a focus on **CSV → CSV** modules that use Lua scripts for data transformation.

It is intended for **developers** and complements the main project README.

---

## 1. High-level architecture

The IPE (Integrated Pipeline Engine) works with:

- **Modules** defined as folders (`modules/Module_X`)
- **Adapters** (input/output) that handle data ingestion and export
- **Lua scripts** that process dataframes
- An installed **manifest.json** that describes the modules
- A dynamically built **DAG** of execution
- An **engine** that runs Lua, chains modules, and exports results

A module can read files, transform data, and produce outputs for other modules or for the user.

---

## 2. Creating a new module

To create a new project with **1 module**:

```bash
ipe init <project_name> -n 1
```

Example:

```bash
ipe init example -n 1
```

This generates:

```text
<project_name>/
  modules/
    Module_1/
      config.json
      script.lua
```

---

## 3. Configuring `config.json` for a CSV → CSV module

File:  
`modules/Module_1/config.json`

Example configuration for a **CSV → CSV** module:

```json
{
  "id": "Module_1",
  "adapter": "csv",
  "depends": [],
  "export_as": "csv",
  "input_args": {
    "batch_read": true,
    "header_row": 1,
    "start_column": 1
  },
  "output_args": {
    "header_row": 1
  }
}
```

Important notes:

- Do **not** put `file_path` here – this is configured later with `ipe set`.
- `export_as` must be set so the engine knows this module produces an exportable output.

---

## 4. Writing the Lua transformation (`script.lua`)

File:  
`modules/Module_1/script.lua`

The input dataframe is injected into the Lua environment as:

```text
<module_name>_input
```

For the module `Module_1`, the input will be available as:

```text
Module_1_input
```

### Minimal working example (echo):

```lua
function main()
    local df = Module_1_input

    -- Returning the dataframe under the "output" key tells the engine to export it
    return {
        output = df
    }
end
```

The engine expects:

- A single dataframe (Lua `UserData`), or  
- A Lua table with one or more dataframes as values

The `output` key is used by the engine to decide which dataframe to export.

---

## 5. Building the module (`.ipe` package)

Create an output directory if needed:

```bash
mkdir dist
```

Build the module:

```bash
ipe build <project_name> -o dist -v 1.0.0
```

Expected result:

```text
dist/<project_name>_1.0.0.ipe
```

---

## 6. Installing the module

Install the built `.ipe` package:

```bash
ipe install dist/<file>.ipe
```

The module is extracted into the IPE tools directory and becomes available to `ipe run`, `ipe set`, and other commands.

---

## 7. Setting input and output paths (`file_path`)

After installing, you must configure the CSV input and output paths:

```bash
ipe set <tool_name> Module_1 -i input.csv -o output.csv
```

This updates the installed `manifest.json` and sets:

- `input_args["file_path"]` for the input adapter
- `output_args["file_path"]` for the output adapter

Example logical structure inside the manifest:

```json
"input_args": {
  "batch_read": true,
  "header_row": 1,
  "start_column": 1,
  "file_path": "path/to/input.csv"
},
"output_args": {
  "header_row": 1,
  "file_path": "path/to/output.csv"
}
```

---

## 8. Running the module

To execute the installed tool:

```bash
ipe run <tool_name>
```

Execution flow:

1. The CSV input adapter reads the file at `input_args["file_path"]`.
2. The engine injects the resulting dataframe into Lua as `<module_name>_input`.
3. The engine calls `main()` from `script.lua`.
4. The return value is interpreted as payload(s).
5. If `export_as = "csv"` and an `output` dataframe is present, the CSV output adapter writes to `output_args["file_path"]`.

---

## 9. Lua ↔ Engine contract

### Injected inputs in Lua

- `<module_name>_input`  
  e.g. `Module_1_input`

### Expected output from `main()`

The Lua `main()` **must** return either:

- A single dataframe (UserData), or  
- A table where at least one entry is named `"output"` and contains a dataframe.

Typical pattern:

```lua
function main()
    local df = Module_1_input

    -- Perform filtering, mapping, column operations here...

    return {
        output = df
    }
end
```

### Examples

#### Filtering rows:

```lua
function main()
    local df = Module_1_input

    local filtered = df:filter(function(row)
        return row["age"] ~= "" and tonumber(row["age"]) > 0
    end)

    return {
        output = filtered
    }
end
```

#### Transforming columns:

```lua
function main()
    local df = Module_1_input

    local transformed = df:map(function(row)
        if row["name"] ~= nil then
            row["name"] = string.upper(row["name"])
        end
        return row
    end)

    return {
        output = transformed
    }
end
```

---

## 10. Full CSV → CSV module flow (summary)

1. `ipe init <project_name> -n 1`  
2. Edit `modules/Module_1/config.json` (set `adapter`, `export_as`, args)  
3. Edit `modules/Module_1/script.lua` (implement `main()`)  
4. `ipe build <project_name> -o dist -v <version>`  
5. `ipe install dist/<project_name>_<version>.ipe`  
6. `ipe set <tool_name> Module_1 -i <input.csv> -o <output.csv>`  
7. `ipe run <tool_name>`  

At this point, the module will read from `input.csv`, process it through `script.lua`, and write the result to `output.csv`.

---

## 11. Development best practices

- Use a consistent key: **`file_path`** for all CSV adapters.
- Avoid panics — always validate type assertions on adapter arguments.
- Keep `config.json` free of environment-specific paths; use `ipe set` for that.
- Treat `manifest.json` as a generated artifact; do not edit it manually whenever possible.
- Keep Lua `main()` small and composable; extract repeated logic into helper functions when needed.
- When adding new module types or adapters, document their expected `input_args` and `output_args` here or under `docs/`.

---

This developer guide should give you everything you need to build and evolve CSV→CSV modules on top of the IPE SDK in a consistent, maintainable way.
