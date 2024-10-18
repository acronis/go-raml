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
		return nil, StacktraceNew("enum must be sequence node", location, WithNodePosition(v))
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

type scalarShape struct{}

func (scalarShape) IsScalar() bool {
	return true
}

type IntegerShape struct {
	scalarShape
	*BaseShape

	EnumFacets
	FormatFacets
	IntegerFacets
}

func (s *IntegerShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *IntegerShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *IntegerShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *IntegerShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*IntegerShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.Minimum = ss.Minimum
	s.Maximum = ss.Maximum
	s.MultipleOf = ss.MultipleOf
	s.Format = ss.Format
	return s, nil
}

func (s *IntegerShape) validate(v interface{}, _ string) error {
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

func (s *IntegerShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*IntegerShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Minimum == nil {
		s.Minimum = ss.Minimum
	} else if ss.Minimum != nil && s.Minimum.Cmp(ss.Minimum) < 0 {
		return nil, StacktraceNew("minimum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Minimum),
			stacktrace.WithInfo("target", *s.Minimum))
	}
	if s.Maximum == nil {
		s.Maximum = ss.Maximum
	} else if ss.Maximum != nil && s.Maximum.Cmp(ss.Maximum) > 0 {
		return nil, StacktraceNew("maximum constraint violation", s.Location,
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
		return nil, StacktraceNew("enum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.Enum.String()),
			stacktrace.WithInfo("target", s.Enum.String()))
	}
	if s.Format == nil {
		s.Format = ss.Format
	} else if ss.Format != nil && SetOfIntegerFormats[*s.Format] != SetOfIntegerFormats[*ss.Format] {
		return nil, StacktraceNew("format constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Format),
			stacktrace.WithInfo("target", *s.Format))
	}
	return s, nil
}

func (s *IntegerShape) check() error {
	if s.Minimum != nil && s.Maximum != nil && s.Minimum.Cmp(s.Maximum) > 0 {
		return StacktraceNew("minimum must be less than or equal to maximum", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	if s.Enum != nil {
		for _, e := range s.Enum {
			switch e.Value.(type) {
			case int, uint:
			default:
				return StacktraceNew("enum value must be int or uint", s.Location,
					stacktrace.WithPosition(&e.Position))
			}
		}
	}
	// invalid format, found by copilot =)
	if s.Format != nil {
		if _, ok := SetOfIntegerFormats[*s.Format]; !ok {
			return StacktraceNew("invalid format", s.Location, stacktrace.WithPosition(&s.Position))
		}
	}
	return nil
}

func (s *IntegerShape) unmarshalYAMLNode(node, valueNode *yaml.Node) error {
	switch node.Value {
	case FacetMinimum:
		if valueNode.Tag != TagInt {
			return StacktraceNew("minimum must be integer", s.Location, WithNodePosition(valueNode))
		}
		num, ok := big.NewInt(0).SetString(valueNode.Value, 10)
		if !ok {
			return StacktraceNew("invalid minimum value", s.Location, WithNodePosition(valueNode))
		}
		s.Minimum = num
	case FacetMaximum:
		if valueNode.Tag != TagInt {
			return StacktraceNew("maximum must be integer", s.Location, WithNodePosition(valueNode))
		}
		num, ok := big.NewInt(0).SetString(valueNode.Value, 10)
		if !ok {
			return StacktraceNew("invalid maximum value", s.Location, WithNodePosition(valueNode))
		}
		s.Maximum = num
	case FacetMultipleOf:
		if err := valueNode.Decode(&s.MultipleOf); err != nil {
			return StacktraceNewWrapped("decode multipleOf", err, s.Location, WithNodePosition(valueNode))
		}
	case FacetFormat:
		if _, ok := SetOfIntegerFormats[valueNode.Value]; !ok {
			return StacktraceNew("invalid format", s.Location, WithNodePosition(valueNode),
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
	return nil
}

func (s *IntegerShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location)
	}
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]
		if err := s.unmarshalYAMLNode(node, valueNode); err != nil {
			return fmt.Errorf("unmarshal %v: %v: %w", node, valueNode, err)
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
	scalarShape
	*BaseShape

	EnumFacets
	FormatFacets
	NumberFacets
}

func (s *NumberShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *NumberShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *NumberShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *NumberShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*NumberShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.Minimum = ss.Minimum
	s.Maximum = ss.Maximum
	s.MultipleOf = ss.MultipleOf
	s.Format = ss.Format
	return s, nil
}

func (s *NumberShape) validate(v interface{}, _ string) error {
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

func (s *NumberShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*NumberShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Minimum == nil {
		s.Minimum = ss.Minimum
	} else if ss.Minimum != nil && *s.Minimum < *ss.Minimum {
		return nil, StacktraceNew("minimum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Minimum),
			stacktrace.WithInfo("target", *s.Minimum))
	}
	if s.Maximum == nil {
		s.Maximum = ss.Maximum
	} else if ss.Maximum != nil && *s.Maximum > *ss.Maximum {
		return nil, StacktraceNew("maximum constraint violation", s.Location,
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
		return nil, StacktraceNew("enum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.Enum.String()),
			stacktrace.WithInfo("target", s.Enum.String()))
	}
	if s.Format == nil {
		s.Format = ss.Format
	} else if ss.Format != nil && *s.Format != *ss.Format {
		return nil, StacktraceNew("format constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Format),
			stacktrace.WithInfo("target", *s.Format))
	}
	return s, nil
}

func (s *NumberShape) check() error {
	if s.Minimum != nil && s.Maximum != nil && *s.Minimum > *s.Maximum {
		return StacktraceNew("minimum must be less than or equal to maximum", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	if s.Enum != nil {
		for _, e := range s.Enum {
			switch e.Value.(type) {
			case int, uint, float64:
			default:
				return StacktraceNew("enum value must be int, uint, float64", s.Location,
					stacktrace.WithPosition(&e.Position))
			}
		}
	}
	return nil
}

func (s *NumberShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location)
	}
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
				return StacktraceNew("invalid format", s.Location, WithNodePosition(valueNode),
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
	scalarShape
	*BaseShape

	EnumFacets
	StringFacets
}

func (s *StringShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *StringShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *StringShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *StringShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*StringShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.MinLength = ss.MinLength
	s.MaxLength = ss.MaxLength
	s.Pattern = ss.Pattern
	return s, nil
}

func (s *StringShape) validate(v interface{}, _ string) error {
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

func (s *StringShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*StringShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.MinLength == nil {
		s.MinLength = ss.MinLength
	} else if ss.MinLength != nil && *s.MinLength < *ss.MinLength {
		return nil, StacktraceNew("minLength constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.MinLength),
			stacktrace.WithInfo("target", *s.MinLength))
	}
	if s.MaxLength == nil {
		s.MaxLength = ss.MaxLength
	} else if ss.MaxLength != nil && *s.MaxLength > *ss.MaxLength {
		return nil, StacktraceNew("maxLength constraint violation", s.Location,
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
		return nil, StacktraceNew("enum constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.Enum.String()),
			stacktrace.WithInfo("target", s.Enum.String()))
	}
	return s, nil
}

func (s *StringShape) check() error {
	if s.MinLength != nil && s.MaxLength != nil && *s.MinLength > *s.MaxLength {
		return StacktraceNew("minLength must be less than or equal to maxLength",
			s.Location, stacktrace.WithPosition(&s.Position))
	}
	if s.Enum != nil {
		for _, e := range s.Enum {
			if _, ok := e.Value.(string); !ok {
				return StacktraceNew("enum value must be string",
					s.Location, stacktrace.WithPosition(&e.Position))
			}
		}
	}
	return nil
}

func (s *StringShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location)
	}
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
				return StacktraceNew("pattern must be string", s.Location, WithNodePosition(valueNode))
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
	scalarShape
	*BaseShape

	LengthFacets
	FileFacets
}

func (s *FileShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *FileShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *FileShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *FileShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*FileShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.MinLength = ss.MinLength
	s.MaxLength = ss.MaxLength
	s.FileTypes = ss.FileTypes
	return s, nil
}

func (s *FileShape) validate(v interface{}, _ string) error {
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

func (s *FileShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*FileShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.MinLength == nil {
		s.MinLength = ss.MinLength
	} else if ss.MinLength != nil && *s.MinLength < *ss.MinLength {
		return nil, StacktraceNew("minLength constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.MinLength),
			stacktrace.WithInfo("target", *s.MinLength))
	}
	if s.MaxLength == nil {
		s.MaxLength = ss.MaxLength
	} else if ss.MaxLength != nil && *s.MaxLength > *ss.MaxLength {
		return nil, StacktraceNew("maxLength constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.MaxLength),
			stacktrace.WithInfo("target", *s.MaxLength))
	}
	if s.FileTypes == nil {
		s.FileTypes = ss.FileTypes
	} else if ss.FileTypes != nil && !isCompatibleEnum(ss.FileTypes, s.FileTypes) {
		return nil, StacktraceNew("file types are incompatible", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.FileTypes.String()),
			stacktrace.WithInfo("target", s.FileTypes.String()))
	}
	return s, nil
}

func (s *FileShape) check() error {
	if s.MinLength != nil && s.MaxLength != nil && *s.MinLength > *s.MaxLength {
		return StacktraceNew("minLength must be less than or equal to maxLength", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	if s.FileTypes != nil {
		for _, e := range s.FileTypes {
			if _, ok := e.Value.(string); !ok {
				return StacktraceNew("file type must be string", s.Location,
					stacktrace.WithPosition(&s.Position))
			}
		}
	}
	return nil
}

func (s *FileShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location)
	}
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
				return StacktraceNew("fileTypes must be sequence node", s.Location, WithNodePosition(valueNode))
			}
			fileTypes := make(Nodes, len(valueNode.Content))
			for i, v := range valueNode.Content {
				if v.Tag != "!!str" {
					return StacktraceNew("member of fileTypes must be string", s.Location, WithNodePosition(v))
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
	scalarShape
	*BaseShape

	EnumFacets
}

func (s *BooleanShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *BooleanShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *BooleanShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *BooleanShape) alias(source Shape) (Shape, error) {
	_, ok := source.(*BooleanShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *BooleanShape) validate(v interface{}, _ string) error {
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

func (s *BooleanShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*BooleanShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Enum == nil {
		s.Enum = ss.Enum
	} else if ss.Enum != nil && !isCompatibleEnum(ss.Enum, s.Enum) {
		return nil, StacktraceNew("enum constraint violation", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", ss.Enum.String()), stacktrace.WithInfo("target", s.Enum.String()))
	}
	return s, nil
}

func (s *BooleanShape) check() error {
	if s.Enum != nil {
		for _, e := range s.Enum {
			if _, ok := e.Value.(bool); !ok {
				return StacktraceNew("enum value must be boolean", s.Location, stacktrace.WithPosition(&e.Position))
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
	scalarShape
	*BaseShape

	FormatFacets
}

func (s *DateTimeShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *DateTimeShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *DateTimeShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *DateTimeShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*DateTimeShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.Format = ss.Format
	return s, nil
}

func (s *DateTimeShape) validate(v interface{}, _ string) error {
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
		case DateTimeFormatRFC3339:
			if _, err := time.Parse(time.RFC3339, i); err != nil {
				return fmt.Errorf("value must match format %s", time.RFC3339)
			}
		// TODO: https://www.rfc-editor.org/rfc/rfc7231#section-7.1.1.1
		case DateTimeFormatRFC2616:
			if _, err := time.Parse(RFC2616, i); err != nil {
				return fmt.Errorf("value must match format %s", RFC2616)
			}
		}
	}

	return nil
}

func (s *DateTimeShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*DateTimeShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Format == nil {
		s.Format = ss.Format
	} else if ss.Format != nil && *s.Format != *ss.Format {
		return nil, StacktraceNew("format constraint violation", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *ss.Format), stacktrace.WithInfo("target", *s.Format))
	}
	return s, nil
}

func (s *DateTimeShape) check() error {
	return nil
}

func (s *DateTimeShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		if i+1 >= len(v) {
			return StacktraceNew("missing value", s.Location)
		}
		node := v[i]
		valueNode := v[i+1]
		if node.Value == "format" {
			if _, ok := SetOfDateTimeFormats[valueNode.Value]; !ok {
				return StacktraceNew("invalid format", s.Location, WithNodePosition(valueNode),
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
	scalarShape
	*BaseShape
}

func (s *DateTimeOnlyShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *DateTimeOnlyShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *DateTimeOnlyShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *DateTimeOnlyShape) alias(source Shape) (Shape, error) {
	_, ok := source.(*DateTimeOnlyShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *DateTimeOnlyShape) validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	if _, err := time.Parse(DateTime, i); err != nil {
		return fmt.Errorf("value must match format %s", DateTime)
	}

	return nil
}

func (s *DateTimeOnlyShape) inherit(source Shape) (Shape, error) {
	_, ok := source.(*DateTimeOnlyShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *DateTimeOnlyShape) check() error {
	return nil
}

func (s *DateTimeOnlyShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location)
	}
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
	scalarShape
	*BaseShape
}

func (s *DateOnlyShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *DateOnlyShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *DateOnlyShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *DateOnlyShape) alias(source Shape) (Shape, error) {
	_, ok := source.(*DateOnlyShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *DateOnlyShape) validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	if _, err := time.Parse(time.DateOnly, i); err != nil {
		return fmt.Errorf("value must match format %s", time.DateOnly)
	}

	return nil
}

func (s *DateOnlyShape) inherit(source Shape) (Shape, error) {
	_, ok := source.(*DateOnlyShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *DateOnlyShape) check() error {
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
	scalarShape
	*BaseShape
}

func (s *TimeOnlyShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *TimeOnlyShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *TimeOnlyShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *TimeOnlyShape) alias(source Shape) (Shape, error) {
	_, ok := source.(*TimeOnlyShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *TimeOnlyShape) validate(v interface{}, _ string) error {
	i, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected string", v)
	}

	if _, err := time.Parse(time.TimeOnly, i); err != nil {
		return fmt.Errorf("value must match format %s", time.TimeOnly)
	}

	return nil
}

func (s *TimeOnlyShape) inherit(source Shape) (Shape, error) {
	_, ok := source.(*TimeOnlyShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *TimeOnlyShape) check() error {
	return nil
}

func (s *TimeOnlyShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location)
	}
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
	scalarShape
	*BaseShape
}

func (s *AnyShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *AnyShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *AnyShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *AnyShape) alias(source Shape) (Shape, error) {
	_, ok := source.(*AnyShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

// Validate checks if the value is nil, implements Shape interface
func (s *AnyShape) validate(_ interface{}, _ string) error {
	return nil
}

func (s *AnyShape) inherit(source Shape) (Shape, error) {
	_, ok := source.(*AnyShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *AnyShape) check() error {
	return nil
}

func (s *AnyShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location)
	}
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
	scalarShape
	*BaseShape
}

func (s *NilShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *NilShape) cloneShallow(base *BaseShape) Shape {
	return s.clone(base, nil)
}

func (s *NilShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *NilShape) alias(source Shape) (Shape, error) {
	_, ok := source.(*NilShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

// Validate checks if the value is nil, implements Shape interface
func (s *NilShape) validate(v interface{}, _ string) error {
	if v != nil {
		return fmt.Errorf("invalid type, got %T, expected nil", v)
	}
	return nil
}

func (s *NilShape) inherit(source Shape) (Shape, error) {
	_, ok := source.(*NilShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *NilShape) check() error {
	return nil
}

func (s *NilShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location)
	}
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
