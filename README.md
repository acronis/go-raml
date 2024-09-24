# RAML 1.0 parser for Go

> [!WARNING]
> The parser is in active development. See supported features in the **Supported features of RAML 1.0 specification**
> section.

This is an implementation of RAML 1.0 parser for Go according
to [the official RAML 1.0 specification](https://github.com/raml-org/raml-spec/blob/master/versions/raml-10/raml-10.md/).

This package aims to achieve the following:

1. Provide a compliant RAML 1.0 parser.
1. Provide an easy-to-use interface for parsing and validating RAML/data type definitions.
1. Provide optimal performance and memory efficiency for API/data type definitions of any size.
1. Provide additional, but not limited to, features built on top of parser, such as:
    1. Middlewares for popular HTTP frameworks that validates the requests/responses according to API definition.
    1. An HTTP gateway that validates requests/responses according to API definition.
    1. A language server built according to Language Server Protocol (LSP).
    1. A linter that may provide additional validations or style enforcement.
    1. Conversion of parsed structure into JSON Schema and back to RAML.

## Supported features of RAML 1.0 specification

The following sections are currently implemented. See notes for each point:

- [ ] RAML API definitions
- [x] RAML Data Types
    - [x] Defining Types
    - [x] Type Declarations
    - [x] Built-in Types
        - [x] The "Any" Type
        - [x] Object Type
            - [x] Property Declarations (explicit and pattern properties)
            - [x] Additional Properties
            - [x] Object Type Specialization
            - [x] Using Discriminator
        - [x] Array Type
        - [x] Scalar Types
            - [x] String
            - [x] Number
            - [x] Integer
            - [x] Boolean
            - [x] Date
            - [x] File
            - [x] Nil Type
        - [x] Union Type (mostly supported, lacks enum support)
        - [x] JSON Schema types (supported, but validation is not implemented)
        - [x] Recursive types
    - [x] User-defined Facets
    - [x] Determine Default Types
    - [x] Type Expressions
        - [x] Inheritance
    - [ ] Multiple Inheritance (supported, but not fully compliant)
    - [x] Inline Type Declarations
    - [x] Defining Examples in RAML
        - [x] Multiple Examples
        - [x] Single Example
        - [x] Validation against defined data type
- [ ] Annotations
    - [x] Declaring Annotation Types
    - [ ] Applying Annotations
        - [ ] Annotating Scalar-valued Nodes
        - [ ] Annotation Targets
        - [x] Annotating types
- [ ] Modularization
    - [ ] Includes
        - [x] Library
            - [x] NamedExample
            - [x] DataType
            - [ ] AnnotationTypeDeclaration
            - [ ] DocumentationItem
            - [ ] ResourceType
            - [ ] Trait
            - [ ] Overlay
            - [ ] Extension
            - [ ] SecurityScheme
- [ ] Conversion
    - [x] Conversion to JSON Schema
    - [ ] Conversion to RAML
- [ ] CLI
    - [x] Validate
    - [ ] Convert to JSON Schema

## Comparison to existing libraries

### Table of libraries

| Parser                 | RAML 1.0 support      | Language         |
|------------------------|-----------------------|------------------|
| AML Modeling Framework | Yes (full support)    | Scala/TypeScript |
| raml-js-parser         | Yes                   | TypeScript       |
| ramlfications          | No                    | Python           |
| go-raml                | Yes (partial support) | Go               |

### Performance

Complex project (7124 types, 148 libraries)

| Project Type                | Time taken | RAM taken |
|-----------------------------|------------|-----------|
| go-raml                     | ~280ms     | ~48MB     |
| AML Modeling Framework (TS) | ~17s       | ~870MB    |

Simple project (<100 types, 1 library)

| Project Type                | Time taken | RAM taken |
|-----------------------------|------------|-----------|
| AML Modeling Framework (TS) | ~2s        | ~100MB    |
| go-raml                     | ~4ms       | ~12MB     |

## Installation

### Library

```
go get -u github.com/acronis/go-raml
```

### CLI

Go install
```
go install github.com/acronis/go-raml/cmd/raml@latest
```

Make install
```
make install
```

## Library usage examples

### Parser options

By default, parser outputs the resulting model as is. This means that information about all links and inheritance chains
is unmodified. Be aware
that the parser may generate recursive structures, depending on your definition, and you may need to implement recursion
detection with the model.

The parser currently provides two options:

* `raml.OptWithValidate()` - performs validation of the resulting model (types inheritance validation, types facet
  validations, annotation types and instances validation, examples, defaults, instances, etc.). Also performs unwrap if
  `raml.OptWithUnwrap()` was not specified, but leaves the original model untouched.

* `raml.OptWithUnwrap()` - performs an unwrap of the resulting model and replaces all definitions with unwrapped
  structures. Unwrap resolves the inheritance chain and links and compiles a complete type, with all properties of its
  parents/links.

### Parsing from string

The following code will parse a RAML string, output a library model and print the common information about the defined
type.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/acronis/go-raml"
)

func main() {
	// Get current working directory that will serve as a base path
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Define RAML 1.0 Library in a string
	content := `#%RAML 1.0 Library
types:
  BasicType: string
  ChildType:
    type: BasicType
    minLength: 5
`

	// Parse with validation
	r, err := raml.ParseFromString(content, "library.raml", workDir, raml.OptWithValidate())
	if err != nil {
		log.Fatal(err)
	}
	// Cast type to Library since our fragment is RAML 1.0 Library
	lib, _ := r.EntryPoint().(*raml.Library)
	typPtr, _ := lib.Types.Get("ChildType")
	// Cast type to StringShape since child type inherits from a string type
	typ := (*typPtr).(*raml.StringShape)
	fmt.Printf(
		"Type name: %s, type: %s, minLength: %d, location: %s\n",
		typ.Base().Name, typ.Base().Type, *typ.MinLength, typ.Base().Location,
	)
	// Cast type to StringShape since parent type is string
	parentTyp := (*typ.Base().Inherits[0]).(*raml.StringShape)
	fmt.Printf("Inherits from:\n")
	fmt.Printf(
		"Type name: %s, type: %s, minLength: %d, location: %s\n",
		parentTyp.Base().Name, parentTyp.Base().Type, parentTyp.MinLength, parentTyp.Base().Location,
	)
}
```

The expected output is:

```
Type name: ChildType, type: string, minLength: 5, location: <absolute_path>/library.raml
Inherits from:
Type name: BasicType, type: string, minLength: 0, location: <absolute_path>/library.raml
```

### Parsing from file

The following code will parse a RAML file, output a library model and print the common information about the defined
type.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/acronis/go-raml"
)

func main() {
	filePath := "<path_to_your_file>"
	r, err := raml.ParseFromPath(content, "library.raml", workDir, raml.OptWithValidate())
	if err != nil {
		log.Fatal(err)
	}
	lib, _ := r.EntryPoint().(*raml.Library)
	typPtr, _ := lib.Types.Get("BasicType")
	typ := *typPtr
	fmt.Printf("Type name: %s, type: %s, location: %s", typ.Base().Name, typ.Base().Type, typ.Base().Location)
}
```

### Validating data against type

Similar to JSON Schema, RAML data types provide a powerful validation mechanism against the defined type.
The following simple example demonstrates how a defined type can be used to validate the values against the type.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/acronis/go-raml"
)

func main() {
	// Get current working directory that will serve as a base path
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Define RAML 1.0 Library in a string
	content := `#%RAML 1.0 Library
types:
  StringType:
    type: string
    minLength: 5
`

	// Parse with validation
	r, err := raml.ParseFromString(content, "library.raml", workDir, raml.OptWithValidate())
	if err != nil {
		log.Fatal(err)
	}
	// Cast type to Library since our fragment is RAML 1.0 Library
	lib, _ := r.EntryPoint().(*raml.Library)
	typPtr, _ := lib.Types.Get("StringType")
	// Cast type to StringShape since defined type is a string
	typ := (*typPtr).(*raml.StringShape)
	fmt.Printf(
		"Type name: %s, type: %s, minLength: %d, location: %s\n",
		typ.Base().Name, typ.Base().Type, *typ.MinLength, typ.Base().Location,
	)
	fmt.Printf("Empty string: %v\n", typ.Validate("", "$"))
	fmt.Printf("Less than 5 characters: %v\n", typ.Validate("abc", "$"))
	fmt.Printf("More than 5 characters: %v\n", typ.Validate("more than 5 chars", "$"))
	fmt.Printf("Not a string: %v\n", typ.Validate(123, "$"))
}
```

The expected output is:

```
Empty string: length must be greater than 5
Less than 5 characters: length must be greater than 5
More than 5 characters: <nil>
Not a string: invalid type, got int, expected string
```

## CLI usage examples

Flags:
* `-v` `--verbosity count` - increase verbosity level, one flag for each level, e.g. `-v` for DEBUG
* `-d` `--ensure-duplicates` - ensure that there are no duplicates in tracebacks

### Validate

The `validate` command validates the RAML file against the RAML 1.0 specification.

The following commands will validate the RAML files and output the validation errors.

One file
```bash
raml validate <path_to_your_file>.raml
```

Multiple files
```bash
raml validate <path_to_your_file1>.raml <path_to_your_file2>.raml <path_to_your_file3>.raml
```

Output example
```
% raml validate library.raml
[11:46:40.053] INFO: Validating RAML... {
  "path": "library.raml"
}
[11:46:40.060] ERROR: RAML is invalid {
  "tracebacks": {
    "traces": {
      "0": {
        "stack": {
          "0": {
            "message": "unwrap shapes",
            "position": "/tmp/library.raml:1",
            "severity": "error",
            "type": "parsing"
          },
          "1": {
            "message": "unwrap shape",
            "position": "/tmp/common.raml:15:5",
            "severity": "error",
            "type": "unwrapping"
          },
          "2": {
            "message": "merge shapes",
            "position": "/tmp/common.raml:15:5",
            "severity": "error",
            "type": "unwrapping"
          },
          "3": {
            "message": "merge shapes",
            "position": "/tmp/common.raml:15:5",
            "severity": "error",
            "type": "unwrapping"
          },
          "4": {
            "message": "inherit property: property: a",
            "position": "/tmp/common.raml:17:10",
            "severity": "error",
            "type": "unwrapping"
          },
          "5": {
            "message": "merge shapes",
            "position": "/tmp/common.raml:17:10",
            "severity": "error",
            "type": "unwrapping"
          },
          "6": {
            "message": "cannot inherit from different type: source: string: target: integer",
            "position": "/tmp/common.raml:17:10",
            "severity": "error",
            "type": "unwrapping"
          }
        }
      },
      "1": {
        "stack": {
          "0": {
            "message": "validate shapes",
            "position": "/tmp/library.raml:1",
            "severity": "error",
            "type": "parsing"
          },
          "1": {
            "message": "check type",
            "position": "/tmp/common.raml:8:5",
            "severity": "error",
            "type": "validating"
          },
          "2": {
            "message": "minProperties must be less than or equal to maxProperties",
            "position": "/tmp/common.raml:8:5",
            "severity": "error",
            "type": "validating"
          },
          "3": {
            "message": "unwrap shape",
            "position": "/tmp/common.raml:15:5",
            "severity": "error",
            "type": "validating"
          },
          "4": {
            "message": "merge shapes",
            "position": "/tmp/common.raml:15:5",
            "severity": "error",
            "type": "unwrapping"
          },
          "5": {
            "message": "merge shapes",
            "position": "/tmp/common.raml:15:5",
            "severity": "error",
            "type": "unwrapping"
          },
          "6": {
            "message": "inherit property: property: a",
            "position": "/tmp/common.raml:17:10",
            "severity": "error",
            "type": "unwrapping"
          },
          "7": {
            "message": "merge shapes",
            "position": "/tmp/common.raml:17:10",
            "severity": "error",
            "type": "unwrapping"
          },
          "8": {
            "message": "cannot inherit from different type: source: string: target: integer",
            "position": "/tmp/common.raml:17:10",
            "severity": "error",
            "type": "unwrapping"
          }
        }
      }
    }
  }
}
[11:46:40.092] ERROR: Command failed {
  "error": "errors have been found in the RAML files"
}
```
