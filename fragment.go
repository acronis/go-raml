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
	FragmentAPI
	FragmentDocumentationItem
	FragmentResourceType
	FragmentTrait
	FragmentAnnotationTypeDeclaration
	FragmentOverlay
	FragmentExtension
	FragmentSecurityScheme
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

type TraitDefinitionGetter interface {
	GetTraitDefinition(refName string) (*TraitDefinition, error)
}

type SecuritySchemeDefinitionGetter interface {
	GetSecuritySchemeDefinition(refName string) (*SecuritySchemeDefinition, error)
}

type ReferenceTypeGetter interface {
	GetReferenceType(refName string) (*BaseShape, error)
}

type ReferenceAnnotationTypeGetter interface {
	GetReferenceAnnotationType(refName string) (*BaseShape, error)
}

func (r *RAML) unmarshalUses(valueNode *yaml.Node, location string) (*orderedmap.OrderedMap[string, *LibraryLink], error) {
	if valueNode.Tag == TagNull {
		return orderedmap.New[string, *LibraryLink](), nil
	} else if valueNode.Kind != yaml.MappingNode {
		return nil, StacktraceNew("uses must be a map", location, WithNodePosition(valueNode))
	}

	uses := orderedmap.New[string, *LibraryLink](len(valueNode.Content) / 2)
	for j := 0; j != len(valueNode.Content); j += 2 {
		name := valueNode.Content[j].Value
		path := valueNode.Content[j+1]
		uses.Set(name, &LibraryLink{
			Value:    path.Value,
			Location: location,
			Position: stacktrace.Position{Line: path.Line, Column: path.Column},
		})
	}
	return uses, nil
}

func (r *RAML) unmarshalTypes(valueNode *yaml.Node, location string, isAnnotationType bool) (*orderedmap.OrderedMap[string, *BaseShape], error) {
	if valueNode.Tag == TagNull {
		return orderedmap.New[string, *BaseShape](), nil
	} else if valueNode.Kind != yaml.MappingNode {
		return nil, StacktraceNew("types must be a map", location, WithNodePosition(valueNode))
	}

	types := orderedmap.New[string, *BaseShape](len(valueNode.Content) / 2)
	for j := 0; j != len(valueNode.Content); j += 2 {
		name := valueNode.Content[j].Value
		data := valueNode.Content[j+1]
		shape, err := r.makeNewShapeYAML(data, name, location)
		if err != nil {
			return nil, StacktraceNewWrapped("unmarshal types: make shape", err, location, WithNodePosition(data))
		}
		types.Set(name, shape)
		if isAnnotationType {
			r.PutAnnotationTypeIntoFragment(name, location, shape)
		} else {
			r.PutTypeIntoFragment(name, location, shape)
		}
		r.PutTypeDefinitionIntoFragment(location, shape)
	}
	return types, nil
}

type Fragment interface {
	LocationGetter
	SecuritySchemeDefinitionGetter
	TraitDefinitionGetter
	ReferenceTypeGetter
	ReferenceAnnotationTypeGetter
}

// Library is the RAML 1.0 Library
type Library struct {
	ID    string
	Usage string

	AnnotationTypes *orderedmap.OrderedMap[string, *BaseShape]
	ResourceTypes   *orderedmap.OrderedMap[string, *ResourceTypeDefinition]
	Types           *orderedmap.OrderedMap[string, *BaseShape]
	Uses            *orderedmap.OrderedMap[string, *LibraryLink]
	Traits          *orderedmap.OrderedMap[string, *TraitDefinition]
	SecuritySchemes *orderedmap.OrderedMap[string, *SecuritySchemeDefinition]

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

func (l *Library) GetSecuritySchemeDefinition(refName string) (*SecuritySchemeDefinition, error) {
	before, after, found := CutReferenceName(refName)

	var ref *SecuritySchemeDefinition

	if !found {
		rr, ok := l.SecuritySchemes.Get(refName)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", refName)
		}
		ref = rr
	} else {
		lib, ok := l.Uses.Get(before)
		if !ok {
			return nil, fmt.Errorf("library \"%s\" not found", before)
		}
		rr, ok := lib.Link.SecuritySchemes.Get(after)
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

func (l *Library) GetTraitDefinition(refName string) (*TraitDefinition, error) {
	before, after, found := CutReferenceName(refName)

	var ref *TraitDefinition

	if !found {
		rr, ok := l.Traits.Get(refName)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", refName)
		}
		ref = rr
	} else {
		lib, ok := l.Uses.Get(before)
		if !ok {
			return nil, fmt.Errorf("library \"%s\" not found", before)
		}
		rr, ok := lib.Link.Traits.Get(after)
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

// UnmarshalYAML unmarshals a Library from a yaml.Node, implementing the yaml.Unmarshaler interface
func (l *Library) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return StacktraceNew("must be map", l.Location, WithNodePosition(value))
	}

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		switch node.Value {
		case "uses":
			uses, err := l.raml.unmarshalUses(valueNode, l.Location)
			if err != nil {
				return StacktraceNewWrapped("parse uses", err, l.Location, WithNodePosition(valueNode))
			}
			l.Uses = uses
		case "types":
			types, err := l.raml.unmarshalTypes(valueNode, l.Location, false)
			if err != nil {
				return StacktraceNewWrapped("parse types", err, l.Location, WithNodePosition(valueNode))
			}
			l.Types = types
		case "annotationTypes":
			types, err := l.raml.unmarshalTypes(valueNode, l.Location, true)
			if err != nil {
				return StacktraceNewWrapped("parse annotation types", err, l.Location, WithNodePosition(valueNode))
			}
			l.AnnotationTypes = types
		case "securitySchemes":
		case "resourceTypes":
		case "traits":
			traitDefs, err := l.raml.unmarshalTraitDefinitions(valueNode, l.Location)
			if err != nil {
				return StacktraceNewWrapped("unmarshal trait definitions", err, l.Location, WithNodePosition(valueNode))
			}
			l.Traits = traitDefs
		case "usage":
			if err := valueNode.Decode(&l.Usage); err != nil {
				return StacktraceNewWrapped("parse usage: value node decode", err, l.Location,
					WithNodePosition(valueNode))
			}
		default:
			if IsCustomDomainExtensionNode(node.Value) {
				name, de, err := l.raml.unmarshalCustomDomainExtension(l.Location, node, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, l.Location,
						WithNodePosition(valueNode))
				}
				l.CustomDomainProperties.Set(name, de)
			} else {
				return StacktraceNew("unknown field", l.Location, stacktrace.WithInfo("field", node.Value))
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

// DataTypeFragment is the RAML 1.0 DataType
type DataTypeFragment struct {
	ID string

	Uses *orderedmap.OrderedMap[string, *LibraryLink]

	Shape *BaseShape

	Location string
	raml     *RAML
}

// GetReferenceType returns a reference type by name, implementing the ReferenceTypeGetter interface
func (dt *DataTypeFragment) GetReferenceType(refName string) (*BaseShape, error) {
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
func (dt *DataTypeFragment) GetReferenceAnnotationType(refName string) (*BaseShape, error) {
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

func (dt *DataTypeFragment) GetSecuritySchemeDefinition(_ string) (*SecuritySchemeDefinition, error) {
	return nil, fmt.Errorf("data type does not define security schemes")
}

func (dt *DataTypeFragment) GetTraitDefinition(refName string) (*TraitDefinition, error) {
	return nil, fmt.Errorf("data type does not define traits")
}

func (dt *DataTypeFragment) GetLocation() string {
	return dt.Location
}

func (dt *DataTypeFragment) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return StacktraceNew("must be map", dt.Location, WithNodePosition(value))
	}

	shapeValue := &yaml.Node{
		Kind: yaml.MappingNode,
	}
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		switch node.Value {
		case "uses":
			uses, err := dt.raml.unmarshalUses(valueNode, dt.Location)
			if err != nil {
				return StacktraceNewWrapped("parse uses", err, dt.Location, WithNodePosition(valueNode))
			}
			dt.Uses = uses
		default:
			shapeValue.Content = append(shapeValue.Content, node, valueNode)
		}
	}
	shape, err := dt.raml.makeNewShapeYAML(shapeValue, filepath.Base(dt.Location), dt.Location)
	if err != nil {
		return StacktraceNewWrapped("parse types: make shape", err, dt.Location, WithNodePosition(shapeValue))
	}
	dt.Shape = shape
	dt.raml.PutTypeDefinitionIntoFragment(dt.Location, shape)
	return nil
}

func (r *RAML) MakeDataTypeFragment(path string) *DataTypeFragment {
	return &DataTypeFragment{
		Uses: orderedmap.New[string, *LibraryLink](0),

		Location: path,
		raml:     r,
	}
}

func (r *RAML) MakeJSONDataType(value []byte, path string) (*DataTypeFragment, error) {
	dt := r.MakeDataTypeFragment(path)
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
	// FIXME: NamedExampleFragment should follow the same pattern as other generic fragments.
	ID string

	Uses *orderedmap.OrderedMap[string, *LibraryLink]

	Map *orderedmap.OrderedMap[string, *Example]

	Location string
	raml     *RAML
}

// GetReferenceAnnotationType returns a reference annotation type by name,
// implementing the ReferenceAnnotationTypeGetter interface
func (ne *NamedExample) GetReferenceAnnotationType(_ string) (*BaseShape, error) {
	return nil, fmt.Errorf("named example does not define references")
}

func (ne *NamedExample) GetSecuritySchemeDefinition(_ string) (*SecuritySchemeDefinition, error) {
	return nil, fmt.Errorf("named example does not define security schemes")
}

func (ne *NamedExample) GetTraitDefinition(_ string) (*TraitDefinition, error) {
	return nil, fmt.Errorf("named example does not define traits")
}

// GetReferenceType returns a reference type by name, implementing the ReferenceTypeGetter interface
func (ne *NamedExample) GetReferenceType(_ string) (*BaseShape, error) {
	return nil, fmt.Errorf("named example does not define references")
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
		return StacktraceNew("must be map", ne.Location, WithNodePosition(value))
	}
	examples := orderedmap.New[string, *Example](len(value.Content) / 2)
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		switch node.Value {
		case "uses":
			uses, err := ne.raml.unmarshalUses(valueNode, ne.Location)
			if err != nil {
				return StacktraceNewWrapped("parse uses", err, ne.Location, WithNodePosition(valueNode))
			}
			ne.Uses = uses
		default:
			example, err := ne.raml.makeExample(valueNode, node.Value, ne.Location)
			if err != nil {
				return StacktraceNewWrapped("make example", err, ne.Location, WithNodePosition(valueNode))
			}
			examples.Set(node.Value, example)
		}
	}
	ne.Map = examples

	return nil
}

///////////////////////////////////////////

type ResourceTypeFragment struct {
	ID int64

	Uses *orderedmap.OrderedMap[string, *LibraryLink]

	ResourceType *ResourceTypeDefinition

	Location string
	raml     *RAML
}

func (r *RAML) MakeResourceTypeFragment(path string) *ResourceTypeFragment {
	return &ResourceTypeFragment{
		Uses: orderedmap.New[string, *LibraryLink](0),

		Location: path,
		raml:     r,
	}
}

type TraitFragment struct {
	ID int64

	Uses *orderedmap.OrderedMap[string, *LibraryLink]

	Trait *TraitDefinition

	Location string
	raml     *RAML
}

// GetReferenceAnnotationType returns a reference annotation type by name,
// implementing the ReferenceAnnotationTypeGetter interface
func (t *TraitFragment) GetReferenceAnnotationType(_ string) (*BaseShape, error) {
	return nil, fmt.Errorf("trait does not define references")
}

func (t *TraitFragment) GetTraitDefinition(_ string) (*TraitDefinition, error) {
	return nil, fmt.Errorf("trait does not define traits")
}

func (t *TraitFragment) GetSecuritySchemeDefinition(_ string) (*SecuritySchemeDefinition, error) {
	return nil, fmt.Errorf("trait does not define security schemes")
}

// GetReferenceType returns a reference type by name, implementing the ReferenceTypeGetter interface
func (t *TraitFragment) GetReferenceType(_ string) (*BaseShape, error) {
	return nil, fmt.Errorf("trait does not define references")
}

func (t *TraitFragment) GetLocation() string {
	return t.Location
}

func (t *TraitFragment) UnmarshalYAML(node *yaml.Node) error {
	traitDef, err := t.raml.makeTraitDefinition(node, t.Location)
	if err != nil {
		return StacktraceNewWrapped("make trait definition", err, t.Location, WithNodePosition(node))
	}
	t.Trait = traitDef
	return nil
}

func (r *RAML) MakeTraitFragment(path string) *TraitFragment {
	return &TraitFragment{
		Uses: orderedmap.New[string, *LibraryLink](0),

		Location: path,
		raml:     r,
	}
}

type SecuritySchemeFragment struct {
	ID int64

	Uses *orderedmap.OrderedMap[string, *LibraryLink]

	SecurityScheme *SecuritySchemeDefinition

	Location string
	raml     *RAML
}

func (r *RAML) MakeSecuritySchemeFragment(path string) *SecuritySchemeFragment {
	return &SecuritySchemeFragment{
		Uses: orderedmap.New[string, *LibraryLink](0),

		Location: path,
		raml:     r,
	}
}

// GetReferenceAnnotationType returns a reference annotation type by name,
// implementing the ReferenceAnnotationTypeGetter interface
func (t *SecuritySchemeFragment) GetReferenceAnnotationType(_ string) (*BaseShape, error) {
	return nil, fmt.Errorf("trait does not define references")
}

func (t *SecuritySchemeFragment) GetTraitDefinition(_ string) (*TraitDefinition, error) {
	return nil, fmt.Errorf("trait does not define traits")
}

func (t *SecuritySchemeFragment) GetSecuritySchemeDefinition(_ string) (*SecuritySchemeDefinition, error) {
	return nil, fmt.Errorf("trait does not define security schemes")
}

// GetReferenceType returns a reference type by name, implementing the ReferenceTypeGetter interface
func (t *SecuritySchemeFragment) GetReferenceType(_ string) (*BaseShape, error) {
	return nil, fmt.Errorf("trait does not define references")
}

func (t *SecuritySchemeFragment) GetLocation() string {
	return t.Location
}

func (t *SecuritySchemeFragment) UnmarshalYAML(node *yaml.Node) error {
	securitySchemeDef, err := t.raml.makeSecuritySchemeDefinition(node, t.Location)
	if err != nil {
		return StacktraceNewWrapped("make security scheme definition", err, t.Location, WithNodePosition(node))
	}
	t.SecurityScheme = securitySchemeDef
	return nil
}

type DocumentationItemFragment struct {
	ID int64

	Uses *orderedmap.OrderedMap[string, *LibraryLink]

	DocumentationItem *DocumentationItem

	Location string
	raml     *RAML
}

func (r *RAML) MakeDocumentationItemFragment(path string) *DocumentationItemFragment {
	return &DocumentationItemFragment{
		Uses: orderedmap.New[string, *LibraryLink](0),

		Location: path,
		raml:     r,
	}
}

type APIFragment struct {
	ID string

	Title       string
	Description string
	Version     string

	Documentation []*DocumentationItem

	// NOTE: We might want to keep forward compatibility with OpenAPI and define multiple servers
	BaseURI           string
	BaseURIParameters *orderedmap.OrderedMap[string, *BaseShape]
	Protocols         []string

	MediaType []string          // Global media types
	SecuredBy []*SecurityScheme // Global security schemes

	AnnotationTypes *orderedmap.OrderedMap[string, *BaseShape]
	ResourceTypes   *orderedmap.OrderedMap[string, *ResourceTypeDefinition]
	Types           *orderedmap.OrderedMap[string, *BaseShape]
	Uses            *orderedmap.OrderedMap[string, *LibraryLink]
	Traits          *orderedmap.OrderedMap[string, *TraitDefinition]
	SecuritySchemes *orderedmap.OrderedMap[string, *SecuritySchemeDefinition]

	EndPoints *orderedmap.OrderedMap[string, *EndPoint]

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	raml     *RAML
}

func (r *RAML) MakeAPIFragment(path string) *APIFragment {
	return &APIFragment{
		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
		Uses:                   orderedmap.New[string, *LibraryLink](0),
		Types:                  orderedmap.New[string, *BaseShape](0),
		ResourceTypes:          orderedmap.New[string, *ResourceTypeDefinition](0),
		Traits:                 orderedmap.New[string, *TraitDefinition](0),
		SecuritySchemes:        orderedmap.New[string, *SecuritySchemeDefinition](0),
		AnnotationTypes:        orderedmap.New[string, *BaseShape](0),
		EndPoints:              orderedmap.New[string, *EndPoint](0),

		Location: path,
		raml:     r,
	}
}

func (api *APIFragment) GetLocation() string {
	return api.Location
}

// GetReferenceType returns a reference type by name, implementing the ReferenceTypeGetter interface
func (api *APIFragment) GetReferenceType(refName string) (*BaseShape, error) {
	before, after, found := CutReferenceName(refName)

	var ref *BaseShape

	if !found {
		rr, ok := api.Types.Get(refName)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", refName)
		}
		ref = rr
	} else {
		lib, ok := api.Uses.Get(before)
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
func (api *APIFragment) GetReferenceAnnotationType(refName string) (*BaseShape, error) {
	before, after, found := CutReferenceName(refName)

	var ref *BaseShape

	if !found {
		rr, ok := api.AnnotationTypes.Get(refName)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", refName)
		}
		ref = rr
	} else {
		lib, ok := api.Uses.Get(before)
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

func (api *APIFragment) GetTraitDefinition(refName string) (*TraitDefinition, error) {
	before, after, found := CutReferenceName(refName)

	var ref *TraitDefinition

	if !found {
		rr, ok := api.Traits.Get(refName)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", refName)
		}
		ref = rr
	} else {
		lib, ok := api.Uses.Get(before)
		if !ok {
			return nil, fmt.Errorf("library \"%s\" not found", before)
		}
		rr, ok := lib.Link.Traits.Get(after)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", after)
		}
		ref = rr
	}

	return ref, nil
}

func (api *APIFragment) GetSecuritySchemeDefinition(refName string) (*SecuritySchemeDefinition, error) {
	before, after, found := CutReferenceName(refName)

	var ref *SecuritySchemeDefinition

	if !found {
		rr, ok := api.SecuritySchemes.Get(refName)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", refName)
		}
		ref = rr
	} else {
		lib, ok := api.Uses.Get(before)
		if !ok {
			return nil, fmt.Errorf("library \"%s\" not found", before)
		}
		rr, ok := lib.Link.SecuritySchemes.Get(after)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", after)
		}
		ref = rr
	}

	return ref, nil
}

func (api *APIFragment) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return StacktraceNew("must be map", api.Location, WithNodePosition(node))
	}

	filtered, err := api.preProcess(node)
	if err != nil {
		return StacktraceNewWrapped("preprocess", err, api.Location)
	}

	for i := 0; i != len(filtered); i += 2 {
		keyNode := filtered[i]
		valueNode := filtered[i+1]
		switch keyNode.Value {
		case "title":
			if err := valueNode.Decode(&api.Title); err != nil {
				return StacktraceNewWrapped("parse title: value node decode", err, api.Location, WithNodePosition(valueNode))
			}
		case "description":
			if err := valueNode.Decode(&api.Description); err != nil {
				return StacktraceNewWrapped("parse description: value node decode", err, api.Location, WithNodePosition(valueNode))
			}
		case "version":
			if err := valueNode.Decode(&api.Version); err != nil {
				return StacktraceNewWrapped("parse version: value node decode", err, api.Location, WithNodePosition(valueNode))
			}
		case "baseUri":
			if err := valueNode.Decode(&api.BaseURI); err != nil {
				return StacktraceNewWrapped("parse base uri: value node decode", err, api.Location, WithNodePosition(valueNode))
			}
		case "baseUriParameters":
			if err := api.unmarshalBaseURIParameters(valueNode); err != nil {
				return StacktraceNewWrapped("unmarshal base uri parameters", err, api.Location, WithNodePosition(valueNode))
			}
		case "documentation":
			documentationItems, err := api.raml.unmarshalDocumentationItems(valueNode, api.Location)
			if err != nil {
				return StacktraceNewWrapped("parse documentation items", err, api.Location, WithNodePosition(valueNode))
			}
			api.Documentation = documentationItems
		case "types":
			types, err := api.raml.unmarshalTypes(valueNode, api.Location, false)
			if err != nil {
				return StacktraceNewWrapped("parse types", err, api.Location, WithNodePosition(valueNode))
			}
			api.Types = types
		case "annotationTypes":
			types, err := api.raml.unmarshalTypes(valueNode, api.Location, true)
			if err != nil {
				return StacktraceNewWrapped("parse annotation types", err, api.Location, WithNodePosition(valueNode))
			}
			api.AnnotationTypes = types
		case "securitySchemes":
			securitySchemeDefs, err := api.raml.unmarshalSecuritySchemes(valueNode, api.Location)
			if err != nil {
				return StacktraceNewWrapped("unmarshal security scheme definitions", err, api.Location, WithNodePosition(valueNode))
			}
			api.SecuritySchemes = securitySchemeDefs
		case FacetUses:
			uses, err := api.raml.unmarshalUses(valueNode, api.Location)
			if err != nil {
				return StacktraceNewWrapped("parse uses", err, api.Location, WithNodePosition(valueNode))
			}
			api.Uses = uses
		case "resourceTypes":
		case "traits":
			traitDefs, err := api.raml.unmarshalTraitDefinitions(valueNode, api.Location)
			if err != nil {
				return StacktraceNewWrapped("unmarshal trait definitions", err, api.Location, WithNodePosition(valueNode))
			}
			api.Traits = traitDefs
		default:
			switch {
			case IsCustomDomainExtensionNode(keyNode.Value):
				name, de, err := api.raml.unmarshalCustomDomainExtension(api.Location, keyNode, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, api.Location, WithNodePosition(valueNode))
				}
				api.CustomDomainProperties.Set(name, de)
			case IsEndPoint(keyNode.Value):
				endpoint, err := api.raml.makeEndpoint(valueNode, api.Location, keyNode.Value, "")
				if err != nil {
					return StacktraceNewWrapped("make endpoint", err, api.Location, WithNodePosition(valueNode))
				}
				api.EndPoints.Set(keyNode.Value, endpoint)
			default:
				return StacktraceNew("unknown field", api.Location, stacktrace.WithInfo("field", keyNode.Value))
			}
		}
	}

	return nil
}

func (api *APIFragment) preProcess(node *yaml.Node) ([]*yaml.Node, error) {
	// NOTE: Preprocessing step ensures that we have global metadata set for this API.
	// This metadata will be used when specifying media types, security and protocols for the requests and responses.
	// Filtered array stores YAML key-value pairs that are not global metadata.
	var filtered []*yaml.Node
	for i := 0; i != len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case FacetProtocols:
			if err := api.unmarshalProtocols(valueNode); err != nil {
				return nil, StacktraceNewWrapped("unmarshal protocols", err, api.Location, WithNodePosition(valueNode))
			}
			api.raml.globalProtocols = api.Protocols
		case "mediaType":
			if err := api.unmarshalMediaType(valueNode); err != nil {
				return nil, StacktraceNewWrapped("unmarshal media type", err, api.Location, WithNodePosition(valueNode))
			}
			api.raml.globalMediaType = api.MediaType
		case FacetSecuredBy:
			securitySchemes, err := api.raml.makeSecuritySchemes(valueNode, api.Location)
			if err != nil {
				return nil, StacktraceNewWrapped("make security schemes", err, api.Location, WithNodePosition(valueNode))
			}
			api.SecuredBy = securitySchemes
			api.raml.globalSecuredBy = securitySchemes
		default:
			filtered = append(filtered, keyNode, valueNode)
		}
	}
	return filtered, nil
}

func (r *RAML) unmarshalTraitDefinitions(node *yaml.Node, location string) (*orderedmap.OrderedMap[string, *TraitDefinition], error) {
	if node.Tag == TagNull {
		return nil, nil
	} else if node.Kind != yaml.MappingNode {
		return nil, StacktraceNew("traits must be a mapping node", location, WithNodePosition(node))
	}

	traitDefs := orderedmap.New[string, *TraitDefinition](len(node.Content) / 2)
	for j := 0; j != len(node.Content); j += 2 {
		nodeName := node.Content[j].Value
		data := node.Content[j+1]

		traitDef, err := r.makeTraitDefinition(data, location)
		if err != nil {
			return nil, StacktraceNewWrapped("make traits", err, location, WithNodePosition(data))
		}
		traitDefs.Set(nodeName, traitDef)
	}
	return traitDefs, nil
}

func (r *RAML) unmarshalSecuritySchemes(node *yaml.Node, location string) (*orderedmap.OrderedMap[string, *SecuritySchemeDefinition], error) {
	if node.Tag == TagNull {
		return nil, nil
	} else if node.Kind != yaml.MappingNode {
		return nil, StacktraceNew("endpoint must be a mapping node", location, WithNodePosition(node))
	}

	securitySchemeDefs := orderedmap.New[string, *SecuritySchemeDefinition](len(node.Content) / 2)
	for j := 0; j != len(node.Content); j += 2 {
		nodeName := node.Content[j].Value
		data := node.Content[j+1]

		securityScheme, err := r.makeSecuritySchemeDefinition(data, location)
		if err != nil {
			return nil, StacktraceNewWrapped("make security scheme definition", err, location, WithNodePosition(data))
		}
		securitySchemeDefs.Set(nodeName, securityScheme)
	}
	return securitySchemeDefs, nil
}

func (api *APIFragment) unmarshalMediaType(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		if node.Tag != TagStr {
			return StacktraceNew("media type must be a string", api.Location, WithNodePosition(node))
		}
		api.MediaType = []string{node.Value}
		return nil
	}
	if err := node.Decode(&api.MediaType); err != nil {
		return StacktraceNewWrapped("parse media type: value node decode", err, api.Location, WithNodePosition(node))
	}
	return nil
}

func (api *APIFragment) unmarshalProtocols(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		if node.Tag != TagStr {
			return StacktraceNew("protocols must be a string", api.Location, WithNodePosition(node))
		}
		api.Protocols = []string{node.Value}
		return nil
	}
	if err := node.Decode(&api.Protocols); err != nil {
		return StacktraceNewWrapped("parse protocols: value node decode", err, api.Location, WithNodePosition(node))
	}
	return nil
}

func (api *APIFragment) unmarshalBaseURIParameters(node *yaml.Node) error {
	if node.Tag == TagNull {
		return nil
	} else if node.Kind != yaml.MappingNode {
		return StacktraceNew("baseUriParameters must be a mapping node", api.Location, WithNodePosition(node))
	}

	api.BaseURIParameters = orderedmap.New[string, *BaseShape](len(node.Content) / 2)
	for j := 0; j != len(node.Content); j += 2 {
		nodeName := node.Content[j].Value
		data := node.Content[j+1]

		shape, err := api.raml.makeNewShapeYAML(data, nodeName, api.Location)
		if err != nil {
			return StacktraceNewWrapped("make new shape yaml", err, api.Location, WithNodePosition(node))
		}
		api.BaseURIParameters.Set(nodeName, shape)
		api.raml.PutTypeDefinitionIntoFragment(api.Location, shape)
	}
	return nil
}
