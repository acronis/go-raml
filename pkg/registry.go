package goraml

import "fmt"

// Not thread-safe
type Registry struct {
	// TODO: Implement interface for fragments.
	fragmentsCache map[string]any // Library, NamedExample, DataType
	fragmentShapes map[string]map[string]*Shape
	// TODO: Temporary shape buffers for resolution and counting.
	UnresolvedShapes []*Shape
	ResolvedShapes   []*Shape
}

func NewRegistry() *Registry {
	return &Registry{
		fragmentShapes: make(map[string]map[string]*Shape),
		fragmentsCache: make(map[string]any),
	}
}

func GetRegistry() *Registry {
	return defaultRegistry
}

func (c *Registry) GetAllShapes() []*Shape {
	var shapes []*Shape
	for _, loc := range c.fragmentShapes {
		for _, shape := range loc {
			shapes = append(shapes, shape)
		}
	}
	return shapes
}

func (c *Registry) GetFragmentShapes(location string) map[string]*Shape {
	return c.fragmentShapes[location]
}

func (c *Registry) GetFromFragment(location string, typeName string) (*Shape, error) {
	loc, ok := c.fragmentShapes[location]
	if !ok {
		return nil, fmt.Errorf("location %s not found", location)
	}
	return loc[typeName], nil
}

func (c *Registry) PutIntoFragment(name string, location string, shape *Shape) {
	loc, ok := c.fragmentShapes[location]
	if !ok {
		loc = make(map[string]*Shape)
		c.fragmentShapes[location] = loc
	}
	loc[name] = shape
}

func (c *Registry) GetFragment(location string) any {
	return c.fragmentsCache[location]
}

func (c *Registry) PutFragment(location string, fragment any) {
	if _, ok := c.fragmentsCache[location]; !ok {
		c.fragmentsCache[location] = fragment
	}
}

// TODO: Switch to map
func (c *Registry) AppendUnresolvedShape(shape *Shape) {
	c.UnresolvedShapes = append(c.UnresolvedShapes, shape)
}

var defaultRegistry = NewRegistry()
