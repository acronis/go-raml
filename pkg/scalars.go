package goraml

import (
	"gopkg.in/yaml.v3"
)

type EnumFacets struct {
	Enum []Node
}

type FormatFacets struct {
	Format *string
}

type IntegerFacets struct {
	Minimum *any
	Maximum *any
}

type IntegerShape struct {
	BaseShape

	EnumFacets
	FormatFacets
	IntegerFacets
}

func (s *IntegerShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *IntegerShape) Clone() Shape {
	c := *s
	return &c
}

// func (s *IntegerShape) Validate(v interface{}) error {
// 	i, ok := v.(int64)
// 	if !ok {
// 		return fmt.Errorf("invalid value")
// 	}

// 	if t.Minimum != nil && *t.Minimum < i {
// 		return fmt.Errorf("value must be in range")
// 	}
// 	if t.Maximum != nil && i > *t.Maximum {
// 		return fmt.Errorf("value must be in range")
// 	}

// 	return nil
// }

func (s *IntegerShape) UnmarshalYAML(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]
		if node.Value == "minimum" {
			if err := valueNode.Decode(&s.Minimum); err != nil {
				return err
			}
		} else if node.Value == "maximum" {
			if err := valueNode.Decode(&s.Maximum); err != nil {
				return err
			}
		} else if node.Value == "format" {
			if err := valueNode.Decode(&s.Format); err != nil {
				return err
			}
		} else if node.Value == "enum" {
			if err := valueNode.Decode(&s.Enum); err != nil {
				return err
			}
		} else {
			dt, err := MakeNode(valueNode, s.Location)
			if err != nil {
				return err
			}
			s.CustomShapeFacets[node.Value] = dt
		}
	}
	return nil
}

type NumberFacets struct {
	IntegerFacets
	MultipleOf *float64
}

type NumberShape struct {
	BaseShape

	EnumFacets
	FormatFacets
	NumberFacets
}

func (s *NumberShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *NumberShape) Clone() Shape {
	c := *s
	return &c
}

func (s *NumberShape) UnmarshalYAML(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]
		if node.Value == "minimum" {
			if err := valueNode.Decode(&s.Minimum); err != nil {
				return err
			}
		} else if node.Value == "maximum" {
			if err := valueNode.Decode(&s.Maximum); err != nil {
				return err
			}
		} else if node.Value == "format" {
			if err := valueNode.Decode(&s.Format); err != nil {
				return err
			}
		} else if node.Value == "enum" {
			if err := valueNode.Decode(&s.Enum); err != nil {
				return err
			}
		} else if node.Value == "multipleOf" {
			if err := valueNode.Decode(&s.MultipleOf); err != nil {
				return err
			}
		} else {
			dt, err := MakeNode(valueNode, s.Location)
			if err != nil {
				return err
			}
			s.CustomShapeFacets[node.Value] = dt
		}
	}
	return nil
}

type ByteLengthFacets struct {
	MaxLength *int64
	MinLength *int64
}

type StringFacets struct {
	ByteLengthFacets
	Pattern *string
}

type StringShape struct {
	BaseShape

	EnumFacets
	StringFacets
}

func (s *StringShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *StringShape) Clone() Shape {
	c := *s
	return &c
}

func (s *StringShape) UnmarshalYAML(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		if node.Value == "minLength" {
			if err := valueNode.Decode(&s.MinLength); err != nil {
				return err
			}
		} else if node.Value == "maxLength" {
			if err := valueNode.Decode(&s.MaxLength); err != nil {
				return err
			}
		} else if node.Value == "pattern" {
			if err := valueNode.Decode(&s.Pattern); err != nil {
				return err
			}
		} else if node.Value == "enum" {
			if err := valueNode.Decode(&s.Enum); err != nil {
				return err
			}
		} else {
			dt, err := MakeNode(valueNode, s.Location)
			if err != nil {
				return err
			}
			s.CustomShapeFacets[node.Value] = dt
		}
	}
	return nil
}

type FileFacets struct {
	FileTypes []Node
}

type FileShape struct {
	BaseShape

	ByteLengthFacets
	FileFacets
}

func (s *FileShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *FileShape) Clone() Shape {
	c := *s
	return &c
}

func (s *FileShape) UnmarshalYAML(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		if node.Value == "minLength" {
			if err := valueNode.Decode(&s.MinLength); err != nil {
				return err
			}
		} else if node.Value == "maxLength" {
			if err := valueNode.Decode(&s.MaxLength); err != nil {
				return err
			}
		} else if node.Value == "fileTypes" {
			if err := valueNode.Decode(&s.FileTypes); err != nil {
				return err
			}
		} else {
			dt, err := MakeNode(valueNode, s.Location)
			if err != nil {
				return err
			}
			s.CustomShapeFacets[node.Value] = dt
		}
	}
	return nil
}

type BooleanShape struct {
	BaseShape

	EnumFacets
}

func (s *BooleanShape) Clone() Shape {
	c := *s
	return &c
}

func (s *BooleanShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *BooleanShape) UnmarshalYAML(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		if node.Value == "enum" {
			if err := valueNode.Decode(&s.Enum); err != nil {
				return err
			}
		} else {
			dt, err := MakeNode(valueNode, s.Location)
			if err != nil {
				return err
			}
			s.CustomShapeFacets[node.Value] = dt
		}
	}

	return nil
}

type DateTimeShape struct {
	BaseShape

	FormatFacets
}

func (s *DateTimeShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *DateTimeShape) Clone() Shape {
	c := *s
	return &c
}

func (s *DateTimeShape) UnmarshalYAML(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]
		if node.Value == "format" {
			if err := valueNode.Decode(&s.Format); err != nil {
				return err
			}
		} else {
			dt, err := MakeNode(valueNode, s.Location)
			if err != nil {
				return err
			}
			s.CustomShapeFacets[node.Value] = dt
		}
	}
	return nil
}

type DateTimeOnlyShape struct {
	BaseShape
}

func (s *DateTimeOnlyShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *DateTimeOnlyShape) Clone() Shape {
	c := *s
	return &c
}

func (s *DateTimeOnlyShape) UnmarshalYAML(v []*yaml.Node) error {
	return nil
}

type DateOnlyShape struct {
	BaseShape
}

func (s *DateOnlyShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *DateOnlyShape) Clone() Shape {
	c := *s
	return &c
}

func (s *DateOnlyShape) UnmarshalYAML(v []*yaml.Node) error {
	return nil
}

type TimeOnlyShape struct {
	BaseShape
}

func (s *TimeOnlyShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *TimeOnlyShape) Clone() Shape {
	c := *s
	return &c
}

func (s *TimeOnlyShape) UnmarshalYAML(v []*yaml.Node) error {
	return nil
}

type AnyShape struct {
	BaseShape
}

func (s *AnyShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *AnyShape) Clone() Shape {
	c := *s
	return &c
}

func (s *AnyShape) UnmarshalYAML(v []*yaml.Node) error {
	return nil
}

type NilShape struct {
	BaseShape
}

func (s *NilShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *NilShape) Clone() Shape {
	c := *s
	return &c
}

func (s *NilShape) UnmarshalYAML(v []*yaml.Node) error {
	return nil
}
