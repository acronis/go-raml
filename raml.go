package raml

import (
	"container/list"
	"context"
	"fmt"
	"strings"
)

// RAML is a store for all fragments and shapes.
// WARNING: Not thread-safe
type RAML struct {
	fragmentsCache          map[string]Fragment // Library, NamedExample, DataType
	fragmentTypes           map[string]map[string]*BaseShape
	fragmentAnnotationTypes map[string]map[string]*BaseShape
	// entryPoint is a Library, NamedExample or DataType fragment that is used as an entry point for the resolution.
	entryPoint Fragment
	// basePath   string

	// May be reused for both validation and resolution.
	domainExtensions []*DomainExtension

	shapes []*BaseShape
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
	return append(annotations, r.domainExtensions...)
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
		fragmentTypes:           make(map[string]map[string]*BaseShape),
		fragmentAnnotationTypes: make(map[string]map[string]*BaseShape),
		fragmentsCache:          make(map[string]Fragment),
		domainExtensions:        make([]*DomainExtension, 0),
		ctx:                     ctx,
	}
}

// Shapes returns all shapes.
func (r *RAML) GetShapes() []*BaseShape {
	return r.shapes
}

func (r *RAML) PutShape(shape *BaseShape) {
	r.shapes = append(r.shapes, shape)
}

// GetFragmentTypePtrs returns fragment shapes as pointers.
func (r *RAML) GetFragmentTypePtrs(location string) map[string]*BaseShape {
	return r.fragmentTypes[location]
}

// GetTypeFromFragmentPtr returns a shape from a fragment as a pointer.
func (r *RAML) GetTypeFromFragmentPtr(location string, typeName string) (*BaseShape, error) {
	loc, ok := r.fragmentTypes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return loc[typeName], nil
}

// PutTypeIntoFragment puts a shape into a fragment.
func (r *RAML) PutTypeIntoFragment(name string, location string, shape *BaseShape) {
	loc, ok := r.fragmentTypes[location]
	if !ok {
		loc = make(map[string]*BaseShape)
		r.fragmentTypes[location] = loc
	}
	loc[name] = shape
}

// GetTypeFromFragmentPtr returns a shape from a fragment.
func (r *RAML) GetAnnotationTypeFromFragmentPtr(location string, typeName string) (*BaseShape, error) {
	loc, ok := r.fragmentAnnotationTypes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return loc[typeName], nil
}

// PutTypeIntoFragment puts a shape into a fragment.
func (r *RAML) PutAnnotationTypeIntoFragment(name string, location string, shape *BaseShape) {
	loc, ok := r.fragmentAnnotationTypes[location]
	if !ok {
		loc = make(map[string]*BaseShape)
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

func (r *RAML) GetReferencedType(refName string, location string) (*BaseShape, error) {
	// External ref - <fragment>.<identifier>
	// Local ref - <identifier>
	before, after, found := strings.Cut(refName, ".")

	var ref *BaseShape
	// TODO: Rework using Fragment interface.
	switch frag := r.GetFragment(location).(type) {
	case *Library:
		if !found {
			r, ok := frag.Types.Get(refName)
			if !ok {
				return nil, fmt.Errorf("reference \"%s\" not found", refName)
			}
			ref = r
		} else {
			lib, ok := frag.Uses.Get(before)
			if !ok {
				return nil, fmt.Errorf("library \"%s\" not found", before)
			}
			r, ok := lib.Link.Types.Get(after)
			if !ok {
				return nil, fmt.Errorf("reference \"%s\" not found", after)
			}
			ref = r
		}
	case *DataType:
		// DataType cannot have local reference to annotation type.
		if !found {
			return nil, fmt.Errorf("invalid reference %s", refName)
		}
		lib, ok := frag.Uses.Get(before)
		if !ok {
			return nil, fmt.Errorf("library \"%s\" not found", before)
		}
		ref, ok = lib.Link.Types.Get(after)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", after)
		}
	default:
		return nil, fmt.Errorf("invalid fragment type: %T", frag)
	}
	return ref, nil
}

func (r *RAML) GetReferencedAnnotationType(refName string, location string) (*BaseShape, error) {
	// TODO: Maybe can be merged with GetReferencedType.
	// External ref - <fragment>.<identifier>
	// Local ref - <identifier>
	before, after, found := strings.Cut(refName, ".")

	var ref *BaseShape
	// TODO: Rework using Fragment interface.
	switch frag := r.GetFragment(location).(type) {
	case *Library:
		if !found {
			r, ok := frag.AnnotationTypes.Get(refName)
			if !ok {
				return nil, fmt.Errorf("reference \"%s\" not found", refName)
			}
			ref = r
		} else {
			lib, ok := frag.Uses.Get(before)
			if !ok {
				return nil, fmt.Errorf("library \"%s\" not found", before)
			}
			r, ok := lib.Link.AnnotationTypes.Get(after)
			if !ok {
				return nil, fmt.Errorf("reference \"%s\" not found", after)
			}
			ref = r
		}
	case *DataType:
		// DataType cannot have local reference to annotation type.
		if !found {
			return nil, fmt.Errorf("invalid reference %s", refName)
		}
		lib, ok := frag.Uses.Get(before)
		if !ok {
			return nil, fmt.Errorf("library \"%s\" not found", before)
		}
		ref, ok = lib.Link.AnnotationTypes.Get(after)
		if !ok {
			return nil, fmt.Errorf("reference \"%s\" not found", after)
		}
	default:
		return nil, fmt.Errorf("invalid fragment type: %T", frag)
	}
	return ref, nil
}
