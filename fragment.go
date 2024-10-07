package raml

import (
	"fmt"
	"path/filepath"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-stacktrace"
)

type FragmentKind int

const (
	FragmentUnknown FragmentKind = iota - 1
	FragmentLibrary
	FragmentDataType
	FragmentNamedExample
)

// CutReferenceName cuts a reference name into two parts: before and after the dot.
func CutReferenceName(refName string) (string, string, bool) {
	// External ref - <fragment>.<identifier>
	// Local ref - <identifier>
	return strings.Cut(refName, ".")
}

type LocationGetter interface {
	GetLocation() string
}

type ReferenceTypeGetter interface {
	GetReferenceType(refName string) (*BaseShape, error)
}

type ReferenceAnnotationTypeGetter interface {
	GetReferenceAnnotationType(refName string) (*BaseShape, error)
}

type Fragment interface {
	LocationGetter
	ReferenceTypeGetter
	ReferenceAnnotationTypeGetter
}

// Library is the RAML 1.0 Library
type Library struct {
	ID              string
	Usage           string
	AnnotationTypes *orderedmap.OrderedMap[string, *BaseShape]
	// TODO: Specific to API fragments. Not supported yet.
	// ResourceTypes   map[string]interface{} `yaml:"resourceTypes"`
	Types *orderedmap.OrderedMap[string, *BaseShape]
	Uses  *orderedmap.OrderedMap[string, *LibraryLink]

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	raml     *RAML
}

// GetReferenceType returns a reference type by name, implementing the ReferenceTypeGetter interface
func (l *Library) GetReferenceType(refName string) (*BaseShape, error) {
	before, after, found := CutReferenceName(refName)

	var ref *BaseShape

	if !found {
		rr, ok := l.Types.Get(refName)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", refName)
		}
		ref = rr
	} else {
		lib, ok := l.Uses.Get(before)
		if !ok {
			return nil, fmt.Errorf("library \"%s\" not found", before)
		}
		rr, ok := lib.Link.Types.Get(after)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", after)
		}
		ref = rr
	}

	return ref, nil
}

// GetReferenceAnnotationType returns a reference annotation type by name,
// implementing the ReferenceAnnotationTypeGetter interface
func (l *Library) GetReferenceAnnotationType(refName string) (*BaseShape, error) {
	before, after, found := CutReferenceName(refName)

	var ref *BaseShape

	if !found {
		rr, ok := l.AnnotationTypes.Get(refName)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", refName)
		}
		ref = rr
	} else {
		lib, ok := l.Uses.Get(before)
		if !ok {
			return nil, fmt.Errorf("library \"%s\" not found", before)
		}
		rr, ok := lib.Link.AnnotationTypes.Get(after)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", after)
		}
		ref = rr
	}

	return ref, nil
}

func (l *Library) GetLocation() string {
	return l.Location
}

type LibraryLink struct {
	ID    string
	Value string

	Link *Library

	Location string
	stacktrace.Position
}

func (l *Library) unmarshalUses(valueNode *yaml.Node) {
	if valueNode.Tag == TagNull {
		return
	}

	l.Uses = orderedmap.New[string, *LibraryLink](len(valueNode.Content) / 2)
	// Map nodes come in pairs in order [key, value]
	for j := 0; j != len(valueNode.Content); j += 2 {
		name := valueNode.Content[j].Value
		path := valueNode.Content[j+1]
		l.Uses.Set(name, &LibraryLink{
			Value:    path.Value,
			Location: l.Location,
			Position: stacktrace.Position{Line: path.Line, Column: path.Column},
		})
	}
}

func (l *Library) unmarshalTypes(valueNode *yaml.Node) {
	if valueNode.Tag == TagNull {
		return
	}

	l.Types = orderedmap.New[string, *BaseShape](len(valueNode.Content) / 2)
	// Map nodes come in pairs in order [key, value]
	for j := 0; j != len(valueNode.Content); j += 2 {
		name := valueNode.Content[j].Value
		data := valueNode.Content[j+1]
		shape, err := l.raml.makeNewShapeYAML(data, name, l.Location)
		if err != nil {
			panic(StacktraceNewWrapped("parse types: make shape", err, l.Location, WithNodePosition(data)))
		}
		l.Types.Set(name, shape)
		l.raml.PutTypeIntoFragment(name, l.Location, shape)
	}
}

func (l *Library) unmarshalAnnotationTypes(valueNode *yaml.Node) error {
	if valueNode.Tag == TagNull {
		return nil
	}

	l.AnnotationTypes = orderedmap.New[string, *BaseShape](len(valueNode.Content) / 2)
	// Map nodes come in pairs in order [key, value]
	for j := 0; j != len(valueNode.Content); j += 2 {
		name := valueNode.Content[j].Value
		data := valueNode.Content[j+1]
		shape, err := l.raml.makeNewShapeYAML(data, name, l.Location)
		if err != nil {
			return StacktraceNewWrapped("parse annotation types: make shape", err, l.Location, WithNodePosition(data))
		}
		l.AnnotationTypes.Set(name, shape)
		l.raml.PutAnnotationTypeIntoFragment(name, l.Location, shape)
	}

	return nil
}

// UnmarshalYAML unmarshals a Library from a yaml.Node, implementing the yaml.Unmarshaler interface
func (l *Library) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return stacktrace.New("must be map", l.Location, WithNodePosition(value))
	}

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		switch node.Value {
		case "uses":
			l.unmarshalUses(valueNode)
		case "types":
			l.unmarshalTypes(valueNode)
		case "annotationTypes":
			if err := l.unmarshalAnnotationTypes(valueNode); err != nil {
				return fmt.Errorf("unmarshall annotation types: %w", err)
			}
		case "usage":
			if err := valueNode.Decode(&l.Usage); err != nil {
				return StacktraceNewWrapped("parse usage: value node decode", err, l.Location, WithNodePosition(valueNode))
			}
		default:
			if IsCustomDomainExtensionNode(node.Value) {
				name, de, err := l.raml.unmarshalCustomDomainExtension(l.Location, node, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, l.Location, WithNodePosition(valueNode))
				}
				l.CustomDomainProperties.Set(name, de)
			}
		}
	}

	return nil
}

func (r *RAML) MakeLibrary(path string) *Library {
	return &Library{
		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
		Uses:                   orderedmap.New[string, *LibraryLink](0),
		Types:                  orderedmap.New[string, *BaseShape](0),
		AnnotationTypes:        orderedmap.New[string, *BaseShape](0),

		Location: path,
		raml:     r,
	}
}

// DataType is the RAML 1.0 DataType
type DataType struct {
	ID    string
	Usage string
	Uses  *orderedmap.OrderedMap[string, *LibraryLink]
	Shape *BaseShape

	Location string
	raml     *RAML
}

// GetReferenceType returns a reference type by name, implementing the ReferenceTypeGetter interface
func (dt *DataType) GetReferenceType(refName string) (*BaseShape, error) {
	before, after, found := CutReferenceName(refName)

	var ref *BaseShape

	if !found {
		return nil, fmt.Errorf("invalid reference %s", refName)
	}
	lib, ok := dt.Uses.Get(before)
	if !ok {
		return nil, fmt.Errorf("library \"%s\" not found", before)
	}
	ref, ok = lib.Link.Types.Get(after)
	if !ok {
		return nil, fmt.Errorf("reference \"%s\" not found", after)
	}

	return ref, nil
}

// GetReferenceAnnotationType returns a reference annotation type by name,
// implementing the ReferenceAnnotationTypeGetter interface
func (dt *DataType) GetReferenceAnnotationType(refName string) (*BaseShape, error) {
	before, after, found := CutReferenceName(refName)

	var ref *BaseShape

	if !found {
		return nil, fmt.Errorf("invalid reference %s", refName)
	}
	lib, ok := dt.Uses.Get(before)
	if !ok {
		return nil, fmt.Errorf("library \"%s\" not found", before)
	}
	ref, ok = lib.Link.AnnotationTypes.Get(after)
	if !ok {
		return nil, fmt.Errorf("reference \"%s\" not found", after)
	}

	return ref, nil
}

func (dt *DataType) GetLocation() string {
	return dt.Location
}

func (dt *DataType) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return stacktrace.New("must be map", dt.Location, WithNodePosition(value))
	}

	shapeValue := &yaml.Node{
		Kind: yaml.MappingNode,
	}
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		switch node.Value {
		case "usage":
			dt.Usage = valueNode.Value
		case "uses":
			if valueNode.Tag == TagNull {
				break
			}

			dt.Uses = orderedmap.New[string, *LibraryLink](len(valueNode.Content) / 2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				path := valueNode.Content[j+1]
				dt.Uses.Set(name, &LibraryLink{
					Value:    path.Value,
					Location: dt.Location,
					Position: stacktrace.Position{Line: path.Line, Column: path.Column},
				})
			}
		default:
			shapeValue.Content = append(shapeValue.Content, node, valueNode)
		}
	}
	shape, err := dt.raml.makeNewShapeYAML(shapeValue, filepath.Base(dt.Location), dt.Location)
	if err != nil {
		return StacktraceNewWrapped("parse types: make shape", err, dt.Location, WithNodePosition(shapeValue))
	}
	dt.Shape = shape
	return nil
}

func (r *RAML) MakeDataType(path string) *DataType {
	return &DataType{
		Uses: orderedmap.New[string, *LibraryLink](0),

		Location: path,
		raml:     r,
	}
}

func (r *RAML) MakeJSONDataType(value []byte, path string) (*DataType, error) {
	dt := r.MakeDataType(path)
	// Convert to yaml node to reuse the same data node creation interface
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Value: "type",
				Tag:   "!!str",
			},
			{
				Kind:  yaml.ScalarNode,
				Value: string(value),
				Tag:   "!!str",
			},
		},
	}
	if err := node.Decode(&dt); err != nil {
		return nil, StacktraceNewWrapped("decode fragment", err, path)
	}
	return dt, nil
}

// NamedExample is the RAML 1.0 NamedExample
type NamedExample struct {
	ID  string
	Map *orderedmap.OrderedMap[string, *Example]

	Location string
	raml     *RAML
}

// GetReferenceAnnotationType returns a reference annotation type by name,
// implementing the ReferenceAnnotationTypeGetter interface
func (ne *NamedExample) GetReferenceAnnotationType(_ string) (*BaseShape, error) {
	return nil, fmt.Errorf("named example does not have references")
}

// GetReferenceType returns a reference type by name, implementing the ReferenceTypeGetter interface
func (ne *NamedExample) GetReferenceType(_ string) (*BaseShape, error) {
	return nil, fmt.Errorf("named example does not have references")
}

func (ne *NamedExample) GetLocation() string {
	return ne.Location
}

func (r *RAML) MakeNamedExample(path string) *NamedExample {
	return &NamedExample{
		Location: path,
		raml:     r,
	}
}

func (ne *NamedExample) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return stacktrace.New("must be map", ne.Location, WithNodePosition(value))
	}
	examples := orderedmap.New[string, *Example](len(value.Content) / 2)
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		example, err := ne.raml.makeExample(valueNode, node.Value, ne.Location)
		if err != nil {
			return StacktraceNewWrapped("make example", err, ne.Location, WithNodePosition(valueNode))
		}
		examples.Set(node.Value, example)
	}
	ne.Map = examples

	return nil
}
