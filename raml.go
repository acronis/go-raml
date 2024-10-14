package raml

import (
	"container/list"
	"context"
	"fmt"
	"reflect"
)

type HookKey string

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

type HookFunc func(ctx context.Context, r *RAML, params ...any) error

func (r *RAML) getHooks(key HookKey) []HookFunc {
	if r.ctx == nil {
		r.ctx = context.Background()
	}
	hooks, ok := r.ctx.Value(key).([]HookFunc)
	if !ok {
		return []HookFunc{}
	}
	if hooks == nil {
		return []HookFunc{}
	}
	return hooks
}

func (r *RAML) setHooks(key HookKey, hooks []HookFunc) {
	if r.ctx == nil {
		r.ctx = context.Background()
	}
	r.ctx = context.WithValue(r.ctx, key, hooks)
}

func (r *RAML) AppendHook(key HookKey, hook HookFunc) {
	hooks := r.getHooks(key)
	hooks = append(hooks, hook)
	r.setHooks(key, hooks)
}

func (r *RAML) PrependHook(key HookKey, hook HookFunc) {
	hooks := r.getHooks(key)
	hooks = append([]HookFunc{hook}, hooks...)
	r.setHooks(key, hooks)
}

func (r *RAML) RemoveHook(key HookKey, hook HookFunc) {
	hooks := r.getHooks(key)
	for i, h := range hooks {
		if reflect.ValueOf(h).Pointer() == reflect.ValueOf(hook).Pointer() {
			hooks = append(hooks[:i], hooks[i+1:]...)
			break
		}
	}
	r.setHooks(key, hooks)
}

func (r *RAML) ClearHooks(key HookKey) {
	r.setHooks(key, []HookFunc{})
}

func (r *RAML) callHooks(key HookKey, params ...any) error {
	hooks := r.getHooks(key)
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		err := hook(r.ctx, r, params...)
		if err != nil {
			return err
		}
	}
	return nil
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
	frag := r.GetFragment(location)
	if frag == nil {
		return nil, fmt.Errorf("fragment not found")
	}
	ref, err := frag.GetReferenceType(refName)
	if err != nil {
		return nil, fmt.Errorf("get reference type: %s: %w", refName, err)
	}
	return ref, nil
}

func (r *RAML) GetReferencedAnnotationType(refName string, location string) (*BaseShape, error) {
	frag := r.GetFragment(location)
	if frag == nil {
		return nil, fmt.Errorf("fragment not found")
	}
	ref, err := frag.GetReferenceAnnotationType(refName)
	if err != nil {
		return nil, fmt.Errorf("get reference annotation type: %s: %w", refName, err)
	}
	return ref, nil
}
