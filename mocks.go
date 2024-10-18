package raml

import "gopkg.in/yaml.v3"

// MockShape is a mock implementation of the Shape interface
type MockShape struct {
	BaseShape              *BaseShape
	MockInherit            func(source Shape) (Shape, error)
	MockCheck              func() error
	MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
	MockValidate           func(v interface{}, ctxPath string) error
	MockUnmarshalYAMLNodes func(v []*yaml.Node) error
	MockString             func() string
	MockIsScalar           func() bool
}

func (u MockShape) inherit(source Shape) (Shape, error) {
	return u.MockInherit(source)
}

func (u MockShape) Base() *BaseShape {
	return u.BaseShape
}

func (u MockShape) check() error {
	return u.MockCheck()
}

func (u MockShape) alias(source Shape) (Shape, error) {
	return u, nil
}

func (u MockShape) cloneShallow(base *BaseShape) Shape {
	return u.clone(base, nil)
}

func (u MockShape) clone(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
	return u.MockClone(base, clonedMap)
}

func (u MockShape) validate(v interface{}, ctxPath string) error {
	return u.MockValidate(v, ctxPath)
}

func (u MockShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	return u.MockUnmarshalYAMLNodes(v)
}

func (u MockShape) String() string {
	return u.MockString()
}

func (u MockShape) IsScalar() bool {
	return u.MockIsScalar()
}
