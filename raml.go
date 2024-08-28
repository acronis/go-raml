package raml

import (
	"context"
	"fmt"
)

// RAML is a store for all fragments and shapes.
// WARNING: Not thread-safe
type RAML struct {
	// TODO: Implement interface for fragments.
	fragmentsCache map[string]Fragment // Library, NamedExample, DataType
	fragmentShapes map[string]map[string]*Shape
	// entryPoint is a Library, NamedExample or DataType fragment that is used as an entry point for the resolution.
	entryPoint Fragment

	// May be reused for both validation and resolution.
	domainExtensions []*DomainExtension

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
		fragmentShapes:   make(map[string]map[string]*Shape),
		fragmentsCache:   make(map[string]Fragment),
		domainExtensions: make([]*DomainExtension, 0),
		ctx:              ctx,
	}
}

// GetAllShapes returns all shapes.
func (r *RAML) GetAllShapes() []Shape {
	var shapes []Shape
	for _, loc := range r.fragmentShapes {
		for _, shape := range loc {
			shapes = append(shapes, *shape)
		}
	}
	return shapes
}

// GetAllShapesPtr returns all shapes as pointers.
func (r *RAML) GetAllShapesPtr() []*Shape {
	var shapes []*Shape
	for _, loc := range r.fragmentShapes {
		for _, shape := range loc {
			shapes = append(shapes, shape)
		}
	}
	return shapes
}

// GetFragmentShapesPtr returns fragment shapes as pointers.
func (r *RAML) GetFragmentShapesPtr(location string) map[string]*Shape {
	return r.fragmentShapes[location]
}

// GetFragmentShapes returns fragment shapes.
func (r *RAML) GetFragmentShapes(location string) map[string]Shape {
	shapes := r.fragmentShapes[location]
	res := make(map[string]Shape)
	for k, v := range shapes {
		res[k] = *v
	}
	return res
}

// GetFromFragmentPtr returns a shape from a fragment as a pointer.
func (r *RAML) GetFromFragmentPtr(location string, typeName string) (*Shape, error) {
	loc, ok := r.fragmentShapes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return loc[typeName], nil
}

// GetFromFragment returns a shape from a fragment.
func (r *RAML) GetFromFragment(location string, typeName string) (Shape, error) {
	loc, ok := r.fragmentShapes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return *loc[typeName], nil
}

// PutIntoFragment puts a shape into a fragment.
func (r *RAML) PutIntoFragment(name string, location string, shape *Shape) {
	loc, ok := r.fragmentShapes[location]
	if !ok {
		loc = make(map[string]*Shape)
		r.fragmentShapes[location] = loc
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
