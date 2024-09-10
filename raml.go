package raml

import (
	"container/list"
	"context"
	"fmt"
)

// RAML is a store for all fragments and shapes.
// WARNING: Not thread-safe
type RAML struct {
	fragmentsCache          map[string]Fragment // Library, NamedExample, DataType
	fragmentTypes           map[string]map[string]*Shape
	fragmentAnnotationTypes map[string]map[string]*Shape
	shapes                  []*Shape
	// entryPoint is a Library, NamedExample or DataType fragment that is used as an entry point for the resolution.
	entryPoint Fragment
	// basePath   string

	// May be reused for both validation and resolution.
	domainExtensions []*DomainExtension
	// Temporary storage for unresolved shapes.
	unresolvedShapes list.List

	// ctx is a context of the RAML, for future use.
	ctx context.Context
}

// EntryPoint returns the entry point of the RAML.
func (r *RAML) EntryPoint() Fragment {
	return r.entryPoint
}

// SetEntryPoint sets the entry point of the RAML.
func (r *RAML) SetEntryPoint(entryPoint Fragment) *RAML {
	r.entryPoint = entryPoint
	return r
}

// GetLocation returns the location of the RAML.
func (r *RAML) GetLocation() string {
	if r.entryPoint == nil {
		return ""
	}
	return r.entryPoint.GetLocation()
}

// GetAllAnnotationsPtr returns all annotations as pointers.
func (r *RAML) GetAllAnnotationsPtr() []*DomainExtension {
	var annotations []*DomainExtension
	for _, de := range r.domainExtensions {
		annotations = append(annotations, de)
	}
	return annotations
}

// GetAllAnnotations returns all annotations.
func (r *RAML) GetAllAnnotations() []DomainExtension {
	var annotations []DomainExtension
	for _, de := range r.domainExtensions {
		annotations = append(annotations, *de)
	}
	return annotations
}

// New creates a new RAML.
func New(ctx context.Context) *RAML {
	return &RAML{
		fragmentTypes:           make(map[string]map[string]*Shape),
		fragmentAnnotationTypes: make(map[string]map[string]*Shape),
		fragmentsCache:          make(map[string]Fragment),
		domainExtensions:        make([]*DomainExtension, 0),
		ctx:                     ctx,
	}
}

// GetShapes returns all shapes.
func (r *RAML) GetShapes() []Shape {
	var shapes []Shape
	for _, shape := range r.shapes {
		shapes = append(shapes, *shape)
	}
	return shapes
}

// GetShapePtrs returns all shapes as pointers.
func (r *RAML) GetShapePtrs() []*Shape {
	return r.shapes
}

func (r *RAML) PutShapePtr(shape *Shape) {
	r.shapes = append(r.shapes, shape)
}

// GetFragmentTypePtrs returns fragment shapes as pointers.
func (r *RAML) GetFragmentTypePtrs(location string) map[string]*Shape {
	return r.fragmentTypes[location]
}

// GetFragmentShapes returns fragment shapes.
func (r *RAML) GetFragmentShapes(location string) map[string]Shape {
	shapes := r.fragmentTypes[location]
	res := make(map[string]Shape, len(shapes))
	for k, v := range shapes {
		res[k] = *v
	}
	return res
}

// GetTypeFromFragmentPtr returns a shape from a fragment as a pointer.
func (r *RAML) GetTypeFromFragmentPtr(location string, typeName string) (*Shape, error) {
	loc, ok := r.fragmentTypes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return loc[typeName], nil
}

// GetTypeFromFragment returns a shape from a fragment.
func (r *RAML) GetTypeFromFragment(location string, typeName string) (Shape, error) {
	loc, ok := r.fragmentTypes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return *loc[typeName], nil
}

// PutTypeIntoFragment puts a shape into a fragment.
func (r *RAML) PutTypeIntoFragment(name string, location string, shape *Shape) {
	loc, ok := r.fragmentTypes[location]
	if !ok {
		loc = make(map[string]*Shape)
		r.fragmentTypes[location] = loc
	}
	loc[name] = shape
}

// GetTypeFromFragmentPtr returns a shape from a fragment.
func (r *RAML) GetAnnotationTypeFromFragmentPtr(location string, typeName string) (*Shape, error) {
	loc, ok := r.fragmentAnnotationTypes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return loc[typeName], nil
}

// GetTypeFromFragment returns a shape from a fragment.
func (r *RAML) GetAnnotationTypeFromFragment(location string, typeName string) (Shape, error) {
	loc, ok := r.fragmentAnnotationTypes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return *loc[typeName], nil
}

// PutTypeIntoFragment puts a shape into a fragment.
func (r *RAML) PutAnnotationTypeIntoFragment(name string, location string, shape *Shape) {
	loc, ok := r.fragmentAnnotationTypes[location]
	if !ok {
		loc = make(map[string]*Shape)
		r.fragmentAnnotationTypes[location] = loc
	}
	loc[name] = shape
}

// GetFragment returns a fragment.
func (r *RAML) GetFragment(location string) Fragment {
	return r.fragmentsCache[location]
}

// PutFragment puts a fragment.
func (r *RAML) PutFragment(location string, fragment Fragment) {
	if _, ok := r.fragmentsCache[location]; !ok {
		r.fragmentsCache[location] = fragment
	}
}
