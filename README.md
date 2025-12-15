[![Go Report Card](https://goreportcard.com/badge/gojp/goreportcard)](https://goreportcard.com/report/gojp/goreportcard)
<!---[![Go Coverage](https://github.com/acronis/go-raml/wiki/coverage.svg)](https://raw.githack.com/wiki/acronis/go-raml/coverage.html)--->

# RAML 1.0 parser for Go

> [!WARNING]
> The parser is in active development. See supported features in the **Supported features of RAML 1.0 specification**
> section.

This is an implementation of RAML parser for Go according to 
[the official RAML 1.0 specification](https://github.com/raml-org/raml-spec/blob/master/versions/raml-10/raml-10.md/).

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

## Why RAML?

RAML is a powerful modelling language that encourages a design-first approach to API definitions by offering
modularity and flexibility. As well as the powerful type system with inheritance support that it offers,
it also allows developers to easily define complex data type schemas by leveraging
both an inheritance mechanism (by referencing a parent type with common properties) and modularity (by referencing common data type fragments).

RAML uses YAML as a markup language, which also contributes to readability, especially in complex definitions.

### RAML vs OpenAPI

While both specifications provide a way to define endpoints, request and responses, and security schemes,
RAML additionally provides ways to reuse their content. For example:

* By specifying traits to apply individual reusable properties to specific endpoints or collections.
* By specifying resource types to apply a common set of available actions or collections.
* By utilizing type inheritance when building complex types.

OpenAPI also allows breaking down the specification into individual and reusable components and include them
in the document. However, these components only deduplicate the content and cannot be modified. In RAML, on the other hand,
it is possible to reuse a component AND modify it, thus reducing the boilerplate.
The following simple example demonstrates how common resource types can be used to build a uniform CRUD API:

```yaml
#%RAML 1.0

title: Test API
version: v1
baseUri: /api/{version}
mediaType: application/json

resourceTypes:
  BatchCollection:
    get:
      responses:
        200:
          body:
            type: <<entityType>>[]
    post:
      body:
        type: <<entityType>>
      responses:
        201:
          body:
            type: <<entityType>>
    delete:
      responses:
        204:
  ItemCollection:
    get:
      responses:
        200:
          body:
            type: <<entityType>>
    put:
      body:
        type: <<entityType>>
      responses:
        204:
    delete:
      responses:
        204:

types:
  User:
    properties:
      id: integer
      name: string
      email: string
      age: integer
  
  Organization:
    properties:
      id: integer
      name: string
      address: string

/users:
  type: 
    BatchCollection:
      entityType: User
  /{user_id}:
    type: 
      ItemCollection:
        entityType: User

/organizations:
  type: 
    BatchCollection:
      entityType: Organization
  /{organization_id}:
    type: 
      ItemCollection:
        entityType: Organization
```

Additionally, developers may use typed annotations to introduce custom behavior, contract validation, assist code generation,
auto-tests, etc. Compared to OpenAPI where extensions are untyped, this allows the developers to describe the extension
right in the document.

### RAML data type vs JSON Schema

For comparison, an example of a simple data type schema in RAML would be the following:

```yaml
#%RAML 1.0 Library

types:
  CommonType:
    additionalProperties: false
    properties:
      id: string
      content: object
    example:
      id: sample.message
      content: {}

  DerivedType:
    type: CommonType # inherits from "CommonType"
    properties:
      content:
        properties:
          value: integer
    example:
      id: "metrics.message"
      content:
        value: 0
```

Where `DerivedType` would translate into the following JSON Schema using Draft-7 specification due to the lack
of the inheritance support:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$ref": "#/definitions/DerivedType",
  "definitions": {
    "DerivedType": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "id": {
          "type": "string"
        },
        "content": {
          "type": "object",
          "properties": {
            "value": {
              "type": "integer"
            }
          },
          "required": ["value"]
        }
      },
      "required": ["id", "content"]
    }
  }
}
```

## What you can do with it

### Create API definitions

RAML's primary objective is to provide a specification for maintainable, design-driven API definitions.

By using RAML, you can make a well-structured API definition that is easy to read, maintain
and use.

### Validate data

Given a powerful type definition system, you can declaratively define your data schema and validate
data instances against that schema.

### Define static configuration

On top of the type definition system, RAML allows defining metadata using custom typed annotations.
One of the ways in which a custom annotation can be used by processors is creating a static configuration
that is validated against a specific schema. For example:

```yaml
#%RAML 1.0 Library

types:
  # Define a data type for configuration
  Config:
    additionalProperties: false
    properties:
      host: string
      port: integer

annotationTypes:
  # Define a custom annotation that would be used to create an instance of a type
  ConfigInstance: Config

# Create an instance of a type.
(ConfigInstance):
  host: example.com
  port: 8080
```

### And more

With all these features, more creative cases can be covered than the use cases mentioned here.

## Supported features of RAML 1.0 specification

The following sections are currently implemented. See notes for each point:

- [ ] RAML API definitions
  - [ ] Resource Types
  - [ ] Traits
  - [x] Endpoint definitions
  - [ ] Security schemes
  - [x] Documentation
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
    - [ ] JSON Schema types (supported, but validation is not implemented)
    - [x] Recursive types
  - [x] User-defined Facets
  - [x] Determine Default Types
  - [x] Type Expressions
  - [x] Type Inheritance
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
    - [x] AnnotationTypeDeclaration
    - [x] DocumentationItem
    - [x] ResourceType
    - [x] Trait
    - [ ] Overlay
    - [ ] Extension
    - [x] SecurityScheme
- [ ] Conversion
  - [x] Conversion to JSON Schema
  - [ ] Conversion to RAML
- [ ] CLI
  - [x] Validate RAML
  - [ ] Convert RAML data type to JSON Schema

## Comparison with existing libraries

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
| go-raml                     | ~4ms       | ~12MB     |
| AML Modeling Framework (TS) | ~2s        | ~100MB    |

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

By default, the parser outputs the resulting model as is and without validation. This means that information about all links and inheritance chains
is unmodified. Be aware
that the parser may generate recursive structures, depending on your definition, and you may need to implement recursion
detection when traversing the model.

The parser currently provides two options:

* `raml.OptWithValidate()` - performs validation of the resulting model (types inheritance validation, types facet
  validations, annotation types and instances validation, examples, defaults, instances, etc.). Also performs unwrap if
  `raml.OptWithUnwrap()` was not specified, but leaves the original types untouched.

* `raml.OptWithUnwrap()` - performs an unwrap of the resulting model and replaces all definitions with unwrapped
  structures. Unwrap resolves the inheritance chain and links and compiles a complete type, with all properties of its
  parents/links.

> [!NOTE]
> In most cases, the use of both flags is advised. If you need to access unmodified types, use only `OptWithValidate()`. Memory consumption may be higher and processing time may be longer since `OptWithValidate()` performs a dedicated copy and unwrap for each type.

### Parsing from string

The following code will parse a RAML string, output a library model, and print the common information about the defined
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
	// Here we omit OptWithUnwrap to show the difference between child and parent.
	r, err := raml.ParseFromString(content, "library.raml", workDir, raml.OptWithValidate())
	if err != nil {
		log.Fatal(err)
	}
	// Cast type to Library since our fragment is RAML 1.0 Library
	lib, _ := r.EntryPoint().(*raml.Library)
	base, _ := lib.Types.Get("ChildType")
	// Cast type to StringShape since child type inherits from a string type
	typ := base.Shape.(*raml.StringShape)
	fmt.Printf(
		"Type name: %s, type: %s, minLength: %d, location: %s\n",
		typ.Name, typ.Type, *typ.MinLength, typ.Location,
	)
	// Cast type to StringShape since parent type is string
	parentTyp := base.Inherits[0].Shape.(*raml.StringShape)
	fmt.Printf("Inherits from:\n")
	fmt.Printf(
		"Type name: %s, type: %s, minLength: %d, location: %s\n",
		parentTyp.Name, parentTyp.Type, parentTyp.MinLength, parentTyp.Location,
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

The following code will parse a RAML file, output a library model, and print the common information about the defined
type.

```go
package main

import (
	"fmt"
	"log"

	"github.com/acronis/go-raml"
)

func main() {
	filePath := "<path_to_your_file>"
	r, err := raml.ParseFromPath(filePath, raml.OptWithValidate(), raml.OptWithUnwrap())
	if err != nil {
		log.Fatal(err)
	}
	// Assuming that the parsed fragment is a RAML Library
	lib, _ := r.EntryPoint().(*raml.Library)
	base, _ := lib.Types.Get("BasicType")
	fmt.Printf("Type name: %s, type: %s, location: %s", base.Name, base.Type, base.Location)
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
	base, _ := lib.Types.Get("StringType")
	// Cast type to StringShape since defined type is a string
	typ := base.Shape.(*raml.StringShape)
	fmt.Printf(
		"Type name: %s, type: %s, minLength: %d, location: %s\n",
		typ.Name, typ.Type, *typ.MinLength, typ.Location,
	)
	fmt.Printf("Empty string: %v\n", base.Validate(""))
	fmt.Printf("Less than 5 characters: %v\n", base.Validate("abc"))
	fmt.Printf("More than 5 characters: %v\n", base.Validate("more than 5 chars"))
	fmt.Printf("Not a string: %v\n", base.Validate(123))
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
