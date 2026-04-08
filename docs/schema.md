# Schema JSON Format

This document describes the structure used by the JSON files in the [`schemas/`](schemas/) directory.

These files are not JSON Schema documents. They are Siremis metadata files that describe:

- which database table a resource uses
- how fields are named and typed
- which operations are enabled in the UI
- how values are shown or entered

## Overview

Each file defines one resource. In practice, the filename usually matches the schema `Name`, for example:

- `schemas/dispatcher.json` -> `Name: "dispatcher"`
- `schemas/topos_d.json` -> `Name: "topos_d"`
- `schemas/uacreg.json` -> `Name: "uacreg"`

At the top level, a schema file typically contains:

- `Name`
- `Title`
- `Table`
- `MenuGroup` (optional)
- `Query`
- `InactiveActions` (optional)
- `Fields`

## Top-Level Structure

Typical shape:

```json
{
  "Name": "example",
  "Title": "Example",
  "Table": "example",
  "MenuGroup": "",
  "Query": {
    "IdField": "id",
    "Limit": 0
  },
  "InactiveActions": {
    "Insert": true,
    "Edit": true,
    "Delete": true
  },
  "Fields": [
    {
      "Title": "ID",
      "Column": "id",
      "Name": "id",
      "Type": "int",
      "Enable": {
        "List": true,
        "Insert": false,
        "Edit": false,
        "Show": true,
        "Search": true
      }
    }
  ]
}
```

## Top-Level Keys

### `Name`

Internal resource name.

- Type: string
- Usually matches the filename without `.json`
- Often also matches the database table name

Example:

```json
"Name": "dispatcher"
```

### `Title`

Human-readable resource label used in the UI.

- Type: string

Example:

```json
"Title": "Dispatcher"
```

### `Table`

Database table name used by the schema.

- Type: string

Example:

```json
"Table": "dispatcher"
```

### `MenuGroup`

Optional menu grouping hint.

- Type: string
- May be omitted
- Used to associate a schema with a menu section or feature group

Examples:

```json
"MenuGroup": "permissions"
```

### `Query`

Query-level metadata for the resource.

- Type: object
- Required in all existing schema files inspected

Supported keys:

- `IdField`: primary identifier column name
- `Limit`: default query limit, commonly `0`

Example:

```json
"Query": {
  "IdField": "id",
  "Limit": 0
}
```

### `InactiveActions`

Optional flags used to disable actions in the UI.

- Type: object
- Common for runtime-managed or read-only tables such as `location`, `acc`, `cdrs`,
  `dialog`, `topos_d`, and `topos_t`

Supported keys seen in the repository:

- `Insert`
- `Edit`
- `Delete`

Example:

```json
"InactiveActions": {
  "Insert": true,
  "Edit": true,
  "Delete": true
}
```

### `Fields`

List of field definitions for the resource.

- Type: array
- Each entry describes one column or visible field

## Field Structure

Each item in `Fields` usually contains:

- `Title`
- `Column`
- `Name`
- `Type`
- `Enable`

It can also contain optional behavior such as:

- `InputForm`
- `ValueShow`
- `ValueInsert`
- `ValueEdit`

Typical field:

```json
{
  "Title": "Username",
  "Column": "username",
  "Name": "username",
  "Type": "str",
  "Enable": {
    "List": true,
    "Insert": true,
    "Edit": true,
    "Show": true,
    "Search": true
  }
}
```

## Field Keys

### `Title`

Human-readable label for the field.

- Type: string

Examples:

- `Id`
- `Call ID`
- `Destination SIP URI`
- `Last Modified`

### `Column`

Database column name.

- Type: string
- Usually matches the SQL column exactly

Example:

```json
"Column": "last_modified"
```

### `Name`

Internal field name.

- Type: string
- In almost all existing files it matches `Column`

Example:

```json
"Name": "last_modified"
```

### `Type`

Logical field type used by the UI.

- Type: string

Common values seen in `schemas/`:

- `int`
- `bigint`
- `float`
- `str`
- `char`
- `datetime`

Typical SQL to schema mapping used in this repository:

- `INT` -> `int`
- `BIGINT` -> `bigint`
- `FLOAT` or numeric decimal-style values -> `float`
- `VARCHAR`, `TEXT`, `MEDIUMTEXT` -> `str`
- character-oriented short values may use `char`
- `DATETIME` -> `datetime`

### `Enable`

Per-action visibility and availability flags.

- Type: object
- Present on every field

Supported keys:

- `List`: show in list view
- `Insert`: allow in create form
- `Edit`: allow in edit form
- `Show`: show in detail view
- `Search`: allow in search form

Example:

```json
"Enable": {
  "List": true,
  "Insert": true,
  "Edit": true,
  "Show": true,
  "Search": false
}
```

Common pattern for auto-increment primary keys:

```json
"Enable": {
  "List": true,
  "Insert": false,
  "Edit": false,
  "Show": true,
  "Search": true
}
```

## Optional Field Extensions

### `InputForm`

Customizes how a field is rendered in forms.

- Type: object

Keys seen in the repository:

- `Type`
- `OptionValues`

Common `InputForm.Type` values seen:

- `number`
- `radio`
- `dataset`

Examples:

Numeric input:

```json
"InputForm": {
  "Type": "number"
}
```

Radio input with generated options:

```json
"InputForm": {
  "Type": "radio",
  "OptionValues": {
    "Func": "ParamVN",
    "Params": [ "0", "simple", "1", "array" ]
  }
}
```

Dataset-backed selector:

```json
"InputForm": {
  "Type": "dataset",
  "OptionValues": {
    "Func": "DBColumnValues",
    "Params": [ "domain", "domain" ]
  }
}
```

### `ValueShow`

Transforms a stored value when displayed.

- Type: object

Keys seen in the repository:

- `Func`
- `Params`

Examples:

Format a float:

```json
"ValueShow": {
  "Func": "Float2D",
  "Params": []
}
```

Show integer flags as named bits:

```json
"ValueShow": {
  "Func": "ListBitFlags",
  "Params": [ "@fld:flags", "0:INACTIVE", "1:TRYING", "2:ADMIN-DISABLED" ]
}
```

### `ValueInsert`

Computes a value during record creation.

- Type: object

Example:

```json
"ValueInsert": {
  "Func": "DateTimeNow",
  "Params": []
}
```

### `ValueEdit`

Computes or refreshes a value during record update.

- Type: object

Example:

```json
"ValueEdit": {
  "Func": "DateTimeNow",
  "Params": []
}
```

## Function-Style Helper Objects

Several optional blocks use the same helper-object shape:

```json
{
  "Func": "FunctionName",
  "Params": [ "...", "..." ]
}
```

This pattern is used in:

- `InputForm.OptionValues`
- `ValueShow`
- `ValueInsert`
- `ValueEdit`

Examples found in the repository include:

- `DateTimeNow`
- `DBColumnValues`
- `ParamVN`
- `Float2D`
- `ListBitFlags`

## Conventions

The current files under `schemas/` follow a few strong conventions:

- The first field is usually the primary key.
- The primary key field usually has `Insert: false` and `Edit: false`.
- `Column` and `Name` are usually identical.
- `Title` is human-readable and can differ from the SQL column name.
- Runtime-generated tables often define `InactiveActions` to disable editing.
- Longer text fields are often hidden from list views with `List: false`.
- Search fields are explicitly controlled per field with `Enable.Search`.

## Minimal Example

```json
{
  "Name": "example_table",
  "Title": "Example Table",
  "Table": "example_table",
  "Query": {
    "IdField": "id",
    "Limit": 0
  },
  "Fields": [
    {
      "Title": "ID",
      "Column": "id",
      "Name": "id",
      "Type": "int",
      "Enable": {
        "List": true,
        "Insert": false,
        "Edit": false,
        "Show": true,
        "Search": true
      }
    },
    {
      "Title": "Name",
      "Column": "name",
      "Name": "name",
      "Type": "str",
      "Enable": {
        "List": true,
        "Insert": true,
        "Edit": true,
        "Show": true,
        "Search": true
      }
    }
  ]
}
```

## Extended Example

```json
{
  "Name": "example_table",
  "Title": "Example Table",
  "Table": "example_table",
  "MenuGroup": "",
  "Query": {
    "IdField": "id",
    "Limit": 0
  },
  "InactiveActions": {
    "Insert": true,
    "Edit": true,
    "Delete": true
  },
  "Fields": [
    {
      "Title": "ID",
      "Column": "id",
      "Name": "id",
      "Type": "int",
      "Enable": {
        "List": true,
        "Insert": false,
        "Edit": false,
        "Show": true,
        "Search": true
      }
    },
    {
      "Title": "Status",
      "Column": "status",
      "Name": "status",
      "Type": "int",
      "Enable": {
        "List": true,
        "Insert": true,
        "Edit": true,
        "Show": true,
        "Search": true
      },
      "InputForm": {
        "Type": "radio",
        "OptionValues": {
          "Func": "ParamVN",
          "Params": [ "0", "Disabled", "1", "Enabled" ]
        }
      },
      "ValueShow": {
        "Func": "ListBitFlags",
        "Params": [ "@fld:status", "0:Disabled", "1:Enabled" ]
      }
    },
    {
      "Title": "Updated At",
      "Column": "updated_at",
      "Name": "updated_at",
      "Type": "datetime",
      "Enable": {
        "List": false,
        "Insert": false,
        "Edit": false,
        "Show": true,
        "Search": true
      },
      "ValueInsert": {
        "Func": "DateTimeNow",
        "Params": []
      },
      "ValueEdit": {
        "Func": "DateTimeNow",
        "Params": []
      }
    }
  ]
}
```

## Notes For Adding New Schema Files

- Keep the filename aligned with `Name`.
- Set `Table` to the SQL table name.
- Set `Query.IdField` to the primary key column.
- Add one `Fields` entry for columns you want exposed in Siremis.
- Use `InactiveActions` for tables that should be browse-only.
- Prefer `Type: "str"` for `VARCHAR`, `TEXT`, and `MEDIUMTEXT`.
- Use `InputForm`, `ValueShow`, `ValueInsert`, and `ValueEdit` only when needed.

## Scope

This document is based on the structures currently used in the repository’s [`schemas/`](schemas/)
directory. If new keys are introduced later, this document should be updated to match them.
