package raml

import (
	"fmt"
	"math/big"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/acronis/go-stacktrace"
)

const (
	RFC2616 = "Mon, 02 Jan 2006 15:04:05 GMT"
	// NOTE: time.DateTime uses "2006-01-02 15:04:05" format which is different from date-time defined in RAML spec.
	DateTime = "2006-01-02T15:04:05"
)

type EnumFacets struct {
	Enum Nodes
}

func (r *RAML) MakeEnum(v *yaml.Node, location string) (Nodes, error) {
	if v.Kind != yaml.SequenceNode {
		return nil, stacktrace.New("enum must be sequence node", location, WithNodePosition(v))
	}
	enums := make(Nodes, len(v.Content))
	for i, v := range v.Content {
		n, err := r.makeRootNode(v, location)
		if err != nil {
			return nil, StacktraceNewWrapped("make node enum", err, location, WithNodePosition(v))
		}
		enums[i] = n
	}
	return enums, nil
}

func isCompatibleEnum(source Nodes, target Nodes) bool {
	// Target enum must be a subset of source enum
	for _, v := range target {
		found := false
		for _, e := range source {
			if v.Value == e.Value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type FormatFacets struct {
	Format *string
}

type IntegerFacets struct {
	Minimum    *big.Int
	Maximum    *big.Int
	MultipleOf *float64
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

func (s *IntegerShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *IntegerShape) Validate(v interface{}, _ string) error {
	var val big.Int
	switch v := v.(type) {
	case int:
		val.SetInt64(int64(v))
	case uint:
		val.SetUint64(uint64(v))
	// json unmarshals numbers as float64
	case float64:
		val.SetInt64(int64(v))
	default:
		return fmt.Errorf("invalid type, got %T, expected int, uint or float64", v)
	}

	if s.Minimum != nil && val.Cmp(s.Minimum) < 0 {
		return fmt.Errorf("value must be greater than %s", s.Minimum.String())
	}
	if s.Maximum != nil && val.Cmp(s.Maximum) > 0 {
		return fmt.Errorf("value must be less than %s", s.Maximum.String())
	}
	// TODO: Implement multipleOf validation
	// TODO: Implement format validation
	if s.Enum != nil {
		// TODO: Probably enum values should be stored as big.Int to simplify validation
		var num any
		if val.IsInt64() {
			num = int(val.Int64())
		} else if val.IsUint64() {
			num = uint(val.Uint64())
		}
		found := false
		for _, e := range s.Enum {
			if e.Value == num {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value must be one of (%s)", s.Enum.String())
		}
	}

	return nil
}

func (s *IntegerShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*IntegerShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Minimum == nil {
		s.Minimum = ss.Minimum
	} else if ss.Minimum != nil && s.Minimum.Cmp(ss.Minimum) < 0 {
		return nil, stacktrace.New("minimum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Minimum),
			stacktrace.WithInfo("target", *s.Minimum))
	}
	if s.Maximum == nil {
		s.Maximum = ss.Maximum
	} else if ss.Maximum != nil && s.Maximum.Cmp(ss.Maximum) > 0 {
		return nil, stacktrace.New("maximum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Maximum),
			stacktrace.WithInfo("target", *s.Maximum))
	}
	// TODO: multipleOf validation
	if s.MultipleOf == nil {
		// TODO: Disallow multipleOf 0 to avoid division by zero during validation
		s.MultipleOf = ss.MultipleOf
	}
	if s.Enum == nil {
		s.Enum = ss.Enum
	} else if ss.Enum != nil && !isCompatibleEnum(ss.Enum, s.Enum) {
		return nil, stacktrace.New("enum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.Enum.String()),
			stacktrace.WithInfo("target", s.Enum.String()))
	}
	if s.Format == nil {
		s.Format = ss.Format
	} else if ss.Format != nil && SetOfIntegerFormats[*s.Format] != SetOfIntegerFormats[*ss.Format] {
		return nil, stacktrace.New("format constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Format),
			stacktrace.WithInfo("target", *s.Format))
	}
	return s, nil
}

func (s *IntegerShape) Check() error {
	if s.Minimum != nil && s.Maximum != nil && s.Minimum.Cmp(s.Maximum) > 0 {
		return stacktrace.New("minimum must be less than or equal to maximum", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	if s.Enum != nil {
		for _, e := range s.Enum {
			switch e.Value.(type) {
			case int, uint:
			default:
				return stacktrace.New("enum value must be int or uint", s.Location,
					stacktrace.WithPosition(&e.Position))
			}
		}
	}
	return nil
}

func (s *IntegerShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]
		switch node.Value {
		case FacetMinimum:
			if valueNode.Tag != TagInt {
				return stacktrace.New("minimum must be integer", s.Location, WithNodePosition(valueNode))
			}
			num, ok := big.NewInt(0).SetString(valueNode.Value, 10)
			if !ok {
				return stacktrace.New("invalid minimum value", s.Location, WithNodePosition(valueNode))
			}
			s.Minimum = num
		case FacetMaximum:
			if valueNode.Tag != TagInt {
				return stacktrace.New("maximum must be integer", s.Location, WithNodePosition(valueNode))
			}
			num, ok := big.NewInt(0).SetString(valueNode.Value, 10)
			if !ok {
				return stacktrace.New("invalid maximum value", s.Location, WithNodePosition(valueNode))
			}
			s.Maximum = num
		case FacetMultipleOf:
			if err := valueNode.Decode(&s.MultipleOf); err != nil {
				return StacktraceNewWrapped("decode multipleOf", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetFormat:
			if _, ok := SetOfIntegerFormats[valueNode.Value]; !ok {
				return stacktrace.New("invalid format", s.Location, WithNodePosition(valueNode),
					stacktrace.WithInfo("allowed_formats", SetOfIntegerFormats))
			}
			if err := valueNode.Decode(&s.Format); err != nil {
				return StacktraceNewWrapped("decode format", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetEnum:
			enums, err := s.raml.MakeEnum(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make enum", err, s.Location, WithNodePosition(valueNode))
			}
			s.Enum = enums
		default:
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(node.Value, n)
		}
	}
	return nil
}

type NumberFacets struct {
	// Minimum and maximum are unset since there's no theoretical minimum and maximum for numbers by default
	Minimum    *float64
	Maximum    *float64
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

func (s *NumberShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *NumberShape) Validate(v interface{}, _ string) error {
	var val float64
	switch v := v.(type) {
	// go-yaml unmarshals integers as int
	case int:
		val = float64(v)
	case uint:
		val = float64(v)
	case float64:
		val = v
	default:
		return fmt.Errorf("invalid type, got %T, expected int, uint, float64", v)
	}

	if s.Minimum != nil && val < *s.Minimum {
		return fmt.Errorf("value must be greater than %f", *s.Minimum)
	}
	if s.Maximum != nil && val > *s.Maximum {
		return fmt.Errorf("value must be less than %f", *s.Maximum)
	}
	// TODO: Implement multipleOf validation
	// TODO: Implement format validation
	if s.Enum != nil {
		found := false
		for _, e := range s.Enum {
			if e.Value == val {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value must be one of (%s)", s.Enum.String())
		}
	}

	return nil
}

func (s *NumberShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*NumberShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Minimum == nil {
		s.Minimum = ss.Minimum
	} else if ss.Minimum != nil && *s.Minimum < *ss.Minimum {
		return nil, stacktrace.New("minimum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Minimum),
			stacktrace.WithInfo("target", *s.Minimum))
	}
	if s.Maximum == nil {
		s.Maximum = ss.Maximum
	} else if ss.Maximum != nil && *s.Maximum > *ss.Maximum {
		return nil, stacktrace.New("maximum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Maximum),
			stacktrace.WithInfo("target", *s.Maximum))
	}
	// TODO: multipleOf validation
	if ss.MultipleOf != nil {
		// TODO: Disallow multipleOf 0 to avoid division by zero during validation
		s.MultipleOf = ss.MultipleOf
	}
	if s.Enum == nil {
		s.Enum = ss.Enum
	} else if ss.Enum != nil && !isCompatibleEnum(ss.Enum, s.Enum) {
		return nil, stacktrace.New("enum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.Enum.String()),
			stacktrace.WithInfo("target", s.Enum.String()))
	}
	if s.Format == nil {
		s.Format = ss.Format
	} else if ss.Format != nil && *s.Format != *ss.Format {
		return nil, stacktrace.New("format constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Format),
			stacktrace.WithInfo("target", *s.Format))
	}
	return s, nil
}

func (s *NumberShape) Check() error {
	if s.Minimum != nil && s.Maximum != nil && *s.Minimum > *s.Maximum {
		return stacktrace.New("minimum must be less than or equal to maximum", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	if s.Enum != nil {
		for _, e := range s.Enum {
			switch e.Value.(type) {
			case int, uint, float64:
			default:
				return stacktrace.New("enum value must be int, uint, float64", s.Location,
					stacktrace.WithPosition(&e.Position))
			}
		}
	}
	return nil
}

func (s *NumberShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]
		switch node.Value {
		case FacetMinimum:
			if err := valueNode.Decode(&s.Minimum); err != nil {
				return StacktraceNewWrapped("decode minimum", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetMaximum:
			if err := valueNode.Decode(&s.Maximum); err != nil {
				return StacktraceNewWrapped("decode maximum", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetFormat:
			if _, ok := SetOfNumberFormats[valueNode.Value]; !ok {
				return stacktrace.New("invalid format", s.Location, WithNodePosition(valueNode),
					stacktrace.WithInfo("allowed_formats", SetOfNumberFormats))
			}
			if err := valueNode.Decode(&s.Format); err != nil {
				return StacktraceNewWrapped("decode format", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetEnum:
			enums, err := s.raml.MakeEnum(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make enum", err, s.Location, WithNodePosition(valueNode))
			}
			s.Enum = enums
		case FacetMultipleOf:
			if err := valueNode.Decode(&s.MultipleOf); err != nil {
				return StacktraceNewWrapped("decode multipleOf", err, s.Location, WithNodePosition(valueNode))
			}
		default:
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(node.Value, n)
		}
	}
	return nil
}

type LengthFacets struct {
	MaxLength *uint64
	MinLength *uint64
}

type StringFacets struct {
	LengthFacets
	Pattern *regexp.Regexp
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

func (s *StringShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *StringShape) Validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	strLen := uint64(len(i))
	if s.MinLength != nil && strLen < *s.MinLength {
		return fmt.Errorf("length must be greater than %d", *s.MinLength)
	}
	if s.MaxLength != nil && strLen > *s.MaxLength {
		return fmt.Errorf("length must be less than %d", *s.MaxLength)
	}
	if s.Pattern != nil && !s.Pattern.MatchString(i) {
		return fmt.Errorf("must match pattern %s", s.Pattern.String())
	}
	if s.Enum != nil {
		found := false
		for _, e := range s.Enum {
			if e.Value == i {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value must be one of (%s)", s.Enum.String())
		}
	}

	return nil
}

func (s *StringShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*StringShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.MinLength == nil {
		s.MinLength = ss.MinLength
	} else if ss.MinLength != nil && *s.MinLength < *ss.MinLength {
		return nil, stacktrace.New("minLength constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.MinLength),
			stacktrace.WithInfo("target", *s.MinLength))
	}
	if s.MaxLength == nil {
		s.MaxLength = ss.MaxLength
	} else if ss.MaxLength != nil && *s.MaxLength > *ss.MaxLength {
		return nil, stacktrace.New("maxLength constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.MaxLength),
			stacktrace.WithInfo("target", *s.MaxLength))
	}
	// FIXME: Patterns are merged unconditionally, but ideally they should be validated against intersection of their DFAs
	if s.Pattern == nil {
		s.Pattern = ss.Pattern
	}
	if s.Enum == nil {
		s.Enum = ss.Enum
	} else if ss.Enum != nil && !isCompatibleEnum(ss.Enum, s.Enum) {
		return nil, stacktrace.New("enum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.Enum.String()),
			stacktrace.WithInfo("target", s.Enum.String()))
	}
	return s, nil
}

func (s *StringShape) Check() error {
	if s.MinLength != nil && s.MaxLength != nil && *s.MinLength > *s.MaxLength {
		return stacktrace.New("minLength must be less than or equal to maxLength",
			s.Location, stacktrace.WithPosition(&s.Position))
	}
	if s.Enum != nil {
		for _, e := range s.Enum {
			if _, ok := e.Value.(string); !ok {
				return stacktrace.New("enum value must be string",
					s.Location, stacktrace.WithPosition(&e.Position))
			}
		}
	}
	return nil
}

func (s *StringShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		switch node.Value {
		case FacetMinLength:
			if err := valueNode.Decode(&s.MinLength); err != nil {
				return StacktraceNewWrapped("decode minLength", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetMaxLength:
			if err := valueNode.Decode(&s.MaxLength); err != nil {
				return StacktraceNewWrapped("decode maxLength", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetPattern:
			if valueNode.Tag != TagStr {
				return stacktrace.New("pattern must be string", s.Location, WithNodePosition(valueNode))
			}

			re, err := regexp.Compile(valueNode.Value)
			if err != nil {
				return StacktraceNewWrapped("decode pattern", err, s.Location, WithNodePosition(valueNode))
			}
			s.Pattern = re
		case FacetEnum:
			enums, err := s.raml.MakeEnum(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make enum", err, s.Location, WithNodePosition(valueNode))
			}
			s.Enum = enums
		default:
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(node.Value, n)
		}
	}
	return nil
}

type FileFacets struct {
	FileTypes Nodes
}

type FileShape struct {
	BaseShape

	LengthFacets
	FileFacets
}

func (s *FileShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *FileShape) Clone() Shape {
	c := *s
	return &c
}

func (s *FileShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *FileShape) Validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	// TODO: What is compared, byte size or base64 string size?
	strLen := uint64(len(i))
	if s.MinLength != nil && strLen < *s.MinLength {
		return fmt.Errorf("length must be greater than %d", *s.MinLength)
	}
	if s.MaxLength != nil && strLen > *s.MaxLength {
		return fmt.Errorf("length must be less than %d", *s.MaxLength)
	}
	// TODO: Validation against file types

	return nil
}

func (s *FileShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*FileShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.MinLength == nil {
		s.MinLength = ss.MinLength
	} else if ss.MinLength != nil && *s.MinLength < *ss.MinLength {
		return nil, stacktrace.New("minLength constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.MinLength),
			stacktrace.WithInfo("target", *s.MinLength))
	}
	if s.MaxLength == nil {
		s.MaxLength = ss.MaxLength
	} else if ss.MaxLength != nil && *s.MaxLength > *ss.MaxLength {
		return nil, stacktrace.New("maxLength constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.MaxLength),
			stacktrace.WithInfo("target", *s.MaxLength))
	}
	if s.FileTypes == nil {
		s.FileTypes = ss.FileTypes
	} else if ss.FileTypes != nil && !isCompatibleEnum(ss.FileTypes, s.FileTypes) {
		return nil, stacktrace.New("enum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.FileTypes.String()),
			stacktrace.WithInfo("target", s.FileTypes.String()))
	}
	return s, nil
}

func (s *FileShape) Check() error {
	if s.MinLength != nil && s.MaxLength != nil && *s.MinLength > *s.MaxLength {
		return stacktrace.New("minLength must be less than or equal to maxLength", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	if s.FileTypes != nil {
		for _, e := range s.FileTypes {
			if _, ok := e.Value.(string); !ok {
				return stacktrace.New("file type must be string", s.Location,
					stacktrace.WithPosition(&s.Position))
			}
		}
	}
	return nil
}

func (s *FileShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		switch node.Value {
		case FacetMinLength:
			if err := valueNode.Decode(&s.MinLength); err != nil {
				return StacktraceNewWrapped("decode minLength", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetMaxLength:
			if err := valueNode.Decode(&s.MaxLength); err != nil {
				return StacktraceNewWrapped("decode maxLength", err, s.Location, WithNodePosition(valueNode))
			}
		case FacetFileTypes:
			if valueNode.Kind != yaml.SequenceNode {
				return stacktrace.New("fileTypes must be sequence node", s.Location, WithNodePosition(valueNode))
			}
			fileTypes := make(Nodes, len(valueNode.Content))
			for i, v := range valueNode.Content {
				if v.Tag != "!!str" {
					return stacktrace.New("member of fileTypes must be string", s.Location, WithNodePosition(v))
				}
				n, err := s.raml.makeRootNode(v, s.Location)
				if err != nil {
					return StacktraceNewWrapped("make node fileTypes", err, s.Location, WithNodePosition(v))
				}
				fileTypes[i] = n
			}
			s.FileTypes = fileTypes
		default:
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(node.Value, n)
		}
	}
	return nil
}

type BooleanShape struct {
	BaseShape

	EnumFacets
}

func (s *BooleanShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *BooleanShape) Clone() Shape {
	c := *s
	return &c
}

func (s *BooleanShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *BooleanShape) Validate(v interface{}, _ string) error {
	i, ok := v.(bool)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected bool", v)
	}

	if s.Enum != nil {
		found := false
		for _, e := range s.Enum {
			if e.Value == i {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value must be one of (%s)", s.Enum.String())
		}
	}

	return nil
}

func (s *BooleanShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*BooleanShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Enum == nil {
		s.Enum = ss.Enum
	} else if ss.Enum != nil && !isCompatibleEnum(ss.Enum, s.Enum) {
		return nil, stacktrace.New("enum constraint violation", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.Enum.String()), stacktrace.WithInfo("target", s.Enum.String()))
	}
	return s, nil
}

func (s *BooleanShape) Check() error {
	if s.Enum != nil {
		for _, e := range s.Enum {
			if _, ok := e.Value.(bool); !ok {
				return stacktrace.New("enum value must be boolean", s.Location, stacktrace.WithPosition(&e.Position))
			}
		}
	}
	return nil
}

func (s *BooleanShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		if node.Value == "enum" {
			enums, err := s.raml.MakeEnum(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make enum", err, s.Location, WithNodePosition(valueNode))
			}
			s.Enum = enums
		} else {
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(node.Value, n)
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

func (s *DateTimeShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *DateTimeShape) Validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	if s.Format == nil {
		if _, err := time.Parse(time.RFC3339, i); err != nil {
			return fmt.Errorf("value must match format %s", time.RFC3339)
		}
	} else {
		switch *s.Format {
		case "rfc3339":
			if _, err := time.Parse(time.RFC3339, i); err != nil {
				return fmt.Errorf("value must match format %s", time.RFC3339)
			}
		// TODO: https://www.rfc-editor.org/rfc/rfc7231#section-7.1.1.1
		case "rfc2616":
			if _, err := time.Parse(RFC2616, i); err != nil {
				return fmt.Errorf("value must match format %s", RFC2616)
			}
		}
	}

	return nil
}

func (s *DateTimeShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*DateTimeShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Format == nil {
		s.Format = ss.Format
	} else if ss.Format != nil && *s.Format != *ss.Format {
		return nil, stacktrace.New("format constraint violation", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Format), stacktrace.WithInfo("target", *s.Format))
	}
	return s, nil
}

func (s *DateTimeShape) Check() error {
	return nil
}

func (s *DateTimeShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]
		if node.Value == "format" {
			if _, ok := SetOfDateTimeFormats[valueNode.Value]; !ok {
				return stacktrace.New("invalid format", s.Location, WithNodePosition(valueNode),
					stacktrace.WithInfo("allowed_formats", SetOfNumberFormats))
			}

			if err := valueNode.Decode(&s.Format); err != nil {
				return StacktraceNewWrapped("decode format", err, s.Location, WithNodePosition(valueNode))
			}
		} else {
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(node.Value, n)
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

func (s *DateTimeOnlyShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *DateTimeOnlyShape) Validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	if _, err := time.Parse(DateTime, i); err != nil {
		return fmt.Errorf("value must match format %s", DateTime)
	}

	return nil
}

func (s *DateTimeOnlyShape) Inherit(source Shape) (Shape, error) {
	_, ok := source.(*DateTimeOnlyShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *DateTimeOnlyShape) Check() error {
	return nil
}

func (s *DateTimeOnlyShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		n, err := s.raml.makeRootNode(valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
		}
		s.CustomShapeFacets.Set(node.Value, n)
	}
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

func (s *DateOnlyShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *DateOnlyShape) Validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	if _, err := time.Parse(time.DateOnly, i); err != nil {
		return fmt.Errorf("value must match format %s", time.DateOnly)
	}

	return nil
}

func (s *DateOnlyShape) Inherit(source Shape) (Shape, error) {
	_, ok := source.(*DateOnlyShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *DateOnlyShape) Check() error {
	return nil
}

func (s *DateOnlyShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		n, err := s.raml.makeRootNode(valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
		}
		s.CustomShapeFacets.Set(node.Value, n)
	}
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

func (s *TimeOnlyShape) clone(_ []Shape) Shape {
	return s.Clone()
}

func (s *TimeOnlyShape) Validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	if _, err := time.Parse(time.TimeOnly, i); err != nil {
		return fmt.Errorf("value must match format %s", time.TimeOnly)
	}

	return nil
}

func (s *TimeOnlyShape) Inherit(source Shape) (Shape, error) {
	_, ok := source.(*TimeOnlyShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *TimeOnlyShape) Check() error {
	return nil
}

func (s *TimeOnlyShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		n, err := s.raml.makeRootNode(valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
		}
		s.CustomShapeFacets.Set(node.Value, n)
	}
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

func (s *AnyShape) clone(_ []Shape) Shape {
	return s.Clone()
}

// Validate checks if the value is nil, implements Shape interface
func (s *AnyShape) Validate(_ interface{}, _ string) error {
	return nil
}

func (s *AnyShape) Inherit(source Shape) (Shape, error) {
	_, ok := source.(*AnyShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *AnyShape) Check() error {
	return nil
}

func (s *AnyShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		n, err := s.raml.makeRootNode(valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
		}
		s.CustomShapeFacets.Set(node.Value, n)
	}
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

// clone returns a copy of the shape
func (s *NilShape) clone(_ []Shape) Shape {
	return s.Clone()
}

// Validate checks if the value is nil, implements Shape interface
func (s *NilShape) Validate(v interface{}, _ string) error {
	if v != nil {
		return fmt.Errorf("invalid type, got %T, expected nil", v)
	}
	return nil
}

func (s *NilShape) Inherit(source Shape) (Shape, error) {
	_, ok := source.(*NilShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *NilShape) Check() error {
	return nil
}

func (s *NilShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		n, err := s.raml.makeRootNode(valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
		}
		s.CustomShapeFacets.Set(node.Value, n)
	}
	return nil
}
