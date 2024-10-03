package raml

import (
	"container/list"
	"context"
	"math/big"
	"reflect"
	"regexp"
	"testing"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

func TestRAML_MakeEnum(t *testing.T) {
	type fields struct {
		fragmentsCache          map[string]Fragment
		fragmentTypes           map[string]map[string]*Shape
		fragmentAnnotationTypes map[string]map[string]*Shape
		shapes                  []*Shape
		entryPoint              Fragment
		domainExtensions        []*DomainExtension
		unresolvedShapes        list.List
		ctx                     context.Context
	}
	type args struct {
		v        *yaml.Node
		location string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Nodes
		wantErr bool
	}{
		{
			name: "valid enum",
			fields: fields{
				fragmentsCache: make(map[string]Fragment),
				fragmentTypes:  make(map[string]map[string]*Shape),
			},
			args: args{
				v: &yaml.Node{
					Kind: yaml.SequenceNode,
					Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "value1"},
						{Kind: yaml.ScalarNode, Value: "value2"},
					},
				},
				location: "test_location",
			},
			want: Nodes{
				{Value: "value1"},
				{Value: "value2"},
			},
			wantErr: false,
		},
		{
			name: "invalid enum kind",
			fields: fields{
				fragmentsCache: make(map[string]Fragment),
				fragmentTypes:  make(map[string]map[string]*Shape),
			},
			args: args{
				v: &yaml.Node{
					Kind: yaml.MappingNode,
				},
				location: "test_location",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid node",
			fields: fields{
				fragmentsCache: make(map[string]Fragment),
				fragmentTypes:  make(map[string]map[string]*Shape),
			},
			args: args{
				v: &yaml.Node{
					Kind: yaml.SequenceNode,
					Content: []*yaml.Node{
						{Kind: yaml.SequenceNode, Value: "{"},
					},
				},
				location: "test_location",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RAML{
				fragmentsCache:          tt.fields.fragmentsCache,
				fragmentTypes:           tt.fields.fragmentTypes,
				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
				shapes:                  tt.fields.shapes,
				entryPoint:              tt.fields.entryPoint,
				domainExtensions:        tt.fields.domainExtensions,
				unresolvedShapes:        tt.fields.unresolvedShapes,
				ctx:                     tt.fields.ctx,
			}
			got, err := r.MakeEnum(tt.args.v, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeEnum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i := range got {
				if got[i].Value != tt.want[i].Value {
					t.Errorf("MakeEnum() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_isCompatibleEnum(t *testing.T) {
	type args struct {
		source Nodes
		target Nodes
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "target is subset of source",
			args: args{
				source: Nodes{{Value: "a"}, {Value: "b"}, {Value: "c"}},
				target: Nodes{{Value: "a"}, {Value: "b"}},
			},
			want: true,
		},
		{
			name: "target is not subset of source",
			args: args{
				source: Nodes{{Value: "a"}, {Value: "b"}},
				target: Nodes{{Value: "a"}, {Value: "d"}},
			},
			want: false,
		},
		{
			name: "target is empty",
			args: args{
				source: Nodes{{Value: "a"}, {Value: "b"}},
				target: Nodes{},
			},
			want: true,
		},
		{
			name: "source is empty",
			args: args{
				source: Nodes{},
				target: Nodes{{Value: "a"}},
			},
			want: false,
		},
		{
			name: "both source and target are empty",
			args: args{
				source: Nodes{},
				target: Nodes{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCompatibleEnum(tt.args.source, tt.args.target); got != tt.want {
				t.Errorf("isCompatibleEnum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntegerShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape     BaseShape
		EnumFacets    EnumFacets
		FormatFacets  FormatFacets
		IntegerFacets IntegerFacets
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid integer",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   "not an integer",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "nil value",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   nil,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "negative integer",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   -123,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "zero value",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   0,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "large integer",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   9223372036854775807,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "integer as float",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   123.0,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "integer as uint",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   uint(123),
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "string integer",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   "123",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "boolean value",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   true,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "empty string",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   "",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "validate minimum value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				IntegerFacets: IntegerFacets{Minimum: func() *big.Int {
					i, _ := new(big.Int).SetString("100", 10)
					return i
				}(),
				},
			},
			args: args{
				v:   5,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "validate maximum value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				IntegerFacets: IntegerFacets{Maximum: func() *big.Int {
					i, _ := new(big.Int).SetString("100", 10)
					return i
				}(),
				},
			},
			args: args{
				v:   150,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "validate enum negative",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{Enum: []*Node{{Value: 1}, {Value: uint64(2)}}},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   3,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "validate enum positive",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{Enum: []*Node{{Value: 1}, {Value: uint64(2)}}},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v:   1,
				in1: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &IntegerShape{
				BaseShape:     tt.fields.BaseShape,
				EnumFacets:    tt.fields.EnumFacets,
				FormatFacets:  tt.fields.FormatFacets,
				IntegerFacets: tt.fields.IntegerFacets,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIntegerShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape     BaseShape
		EnumFacets    EnumFacets
		FormatFacets  FormatFacets
		IntegerFacets IntegerFacets
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Shape
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				source: &IntegerShape{
					BaseShape:     BaseShape{},
					EnumFacets:    EnumFacets{},
					FormatFacets:  FormatFacets{},
					IntegerFacets: IntegerFacets{},
				},
			},
			want: &IntegerShape{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				source: &StringShape{
					BaseShape:    BaseShape{},
					EnumFacets:   EnumFacets{},
					StringFacets: StringFacets{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "minimum constraint violation",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				IntegerFacets: IntegerFacets{Minimum: func() *big.Int {
					i, _ := new(big.Int).SetString("100", 10)
					return i
				}()},
			},
			args: args{
				source: &IntegerShape{
					BaseShape:    BaseShape{},
					EnumFacets:   EnumFacets{},
					FormatFacets: FormatFacets{},
					IntegerFacets: IntegerFacets{Minimum: func() *big.Int {
						i, _ := new(big.Int).SetString("120", 10)
						return i
					}()},
				},
			},
			wantErr: true,
		},
		{
			name: "maximum constraint violation",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				IntegerFacets: IntegerFacets{Maximum: func() *big.Int {
					i, _ := new(big.Int).SetString("100", 10)
					return i
				}(),
				},
			},
			args: args{
				source: &IntegerShape{
					BaseShape:    BaseShape{},
					EnumFacets:   EnumFacets{},
					FormatFacets: FormatFacets{},
					IntegerFacets: IntegerFacets{Maximum: func() *big.Int {
						i, _ := new(big.Int).SetString("80", 10)
						return i
					}()},
				},
			},

			wantErr: true,
		},
		{
			name: "enum constraint violation",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{Enum: []*Node{{Value: 1}, {Value: uint64(2)}}},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				source: &IntegerShape{
					BaseShape:     BaseShape{},
					EnumFacets:    EnumFacets{Enum: []*Node{{Value: 3}}},
					FormatFacets:  FormatFacets{},
					IntegerFacets: IntegerFacets{},
				},
			},
			wantErr: true,
		},
		{
			name: "format constraint violation",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				FormatFacets: FormatFacets{Format: func() *string {
					s := "int32"
					return &s
				}()},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				source: &IntegerShape{
					BaseShape:  BaseShape{},
					EnumFacets: EnumFacets{},
					FormatFacets: FormatFacets{Format: func() *string {
						s := "int64"
						return &s
					}()},
					IntegerFacets: IntegerFacets{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &IntegerShape{
				BaseShape:     tt.fields.BaseShape,
				EnumFacets:    tt.fields.EnumFacets,
				FormatFacets:  tt.fields.FormatFacets,
				IntegerFacets: tt.fields.IntegerFacets,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Inherit() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntegerShape_Check(t *testing.T) {
	type fields struct {
		BaseShape     BaseShape
		EnumFacets    EnumFacets
		FormatFacets  FormatFacets
		IntegerFacets IntegerFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid shape",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			wantErr: false,
		},
		{
			name: "invalid minimum and maximum",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				IntegerFacets: IntegerFacets{
					Minimum: func() *big.Int {
						i, _ := new(big.Int).SetString("100", 10)
						return i
					}(),
					Maximum: func() *big.Int {
						i, _ := new(big.Int).SetString("50", 10)
						return i
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "valid enum",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{Enum: []*Node{{Value: 1}, {Value: 2}}},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			wantErr: false,
		},
		{
			name: "invalid enum",
			fields: fields{
				BaseShape:     BaseShape{},
				EnumFacets:    EnumFacets{Enum: []*Node{{Value: "a"}, {Value: "b"}}},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			wantErr: true,
		},
		{
			name: "valid format",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				FormatFacets: FormatFacets{Format: func() *string {
					s := "int32"
					return &s
				}()},
				IntegerFacets: IntegerFacets{},
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				FormatFacets: FormatFacets{Format: func() *string {
					s := "invalid"
					return &s
				}()},
				IntegerFacets: IntegerFacets{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &IntegerShape{
				BaseShape:     tt.fields.BaseShape,
				EnumFacets:    tt.fields.EnumFacets,
				FormatFacets:  tt.fields.FormatFacets,
				IntegerFacets: tt.fields.IntegerFacets,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIntegerShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape     BaseShape
		EnumFacets    EnumFacets
		FormatFacets  FormatFacets
		IntegerFacets IntegerFacets
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid YAML nodes",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "key1"},
					{Kind: yaml.ScalarNode, Value: "value1"},
					{Kind: yaml.ScalarNode, Value: "key2"},
					{Kind: yaml.ScalarNode, Value: "value2"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid YAML nodes",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.MappingNode, Value: "{"},
					{Kind: yaml.MappingNode, Value: "{"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty YAML nodes",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{},
			},
			wantErr: false,
		},
		{
			name: "odd number of YAML nodes",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "key1"},
					{Kind: yaml.ScalarNode, Value: "value1"},
					{Kind: yaml.ScalarNode, Value: "empty value"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet minimum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "minimum"},
					{Kind: yaml.ScalarNode, Value: "10", Tag: "!!int"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet minimum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "minimum"},
					{Kind: yaml.ScalarNode, Value: "invalid", Tag: "!!int"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet maximum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "maximum"},
					{Kind: yaml.ScalarNode, Value: "100", Tag: "!!int"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet maximum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "maximum"},
					{Kind: yaml.ScalarNode, Value: "invalid", Tag: "!!int"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet multipleOf",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "multipleOf"},
					{Kind: yaml.ScalarNode, Value: "5"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet multipleOf",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "multipleOf"},
					{Kind: yaml.ScalarNode, Value: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet format",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "format"},
					{Kind: yaml.ScalarNode, Value: "int32"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet format",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "format"},
					{Kind: yaml.ScalarNode, Value: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet enum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "enum"},
					{Kind: yaml.SequenceNode, Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "1"},
						{Kind: yaml.ScalarNode, Value: "2"},
					}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet enum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:    EnumFacets{},
				FormatFacets:  FormatFacets{},
				IntegerFacets: IntegerFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "enum"},
					{Kind: yaml.ScalarNode, Value: "invalid"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &IntegerShape{
				BaseShape:     tt.fields.BaseShape,
				EnumFacets:    tt.fields.EnumFacets,
				FormatFacets:  tt.fields.FormatFacets,
				IntegerFacets: tt.fields.IntegerFacets,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		EnumFacets   EnumFacets
		FormatFacets FormatFacets
		NumberFacets NumberFacets
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid number",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   123.45,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "nil value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   nil,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "negative number",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   -123.45,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "zero value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   uint(0),
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "large number",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   1e10,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "number as integer",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "boolean value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   true,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "empty string",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   "",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "validate minimum value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{Minimum: func() *float64 {
					f := 100.0
					return &f
				}()},
			},
			args: args{
				v:   5.0,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "validate maximum value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{Maximum: func() *float64 {
					f := 100.0
					return &f
				}()},
			},
			args: args{
				v:   150.0,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "validate enum negative",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{Enum: []*Node{{Value: 1.0}, {Value: 2.0}}},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   3.0,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "validate enum positive",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{Enum: []*Node{{Value: 1.0}, {Value: 2.0}}},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v:   1.0,
				in1: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NumberShape{
				BaseShape:    tt.fields.BaseShape,
				EnumFacets:   tt.fields.EnumFacets,
				FormatFacets: tt.fields.FormatFacets,
				NumberFacets: tt.fields.NumberFacets,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		EnumFacets   EnumFacets
		FormatFacets FormatFacets
		NumberFacets NumberFacets
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Shape
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				source: &NumberShape{
					BaseShape:    BaseShape{},
					EnumFacets:   EnumFacets{},
					FormatFacets: FormatFacets{},
					NumberFacets: NumberFacets{},
				},
			},
			want: &NumberShape{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				source: &IntegerShape{
					BaseShape:     BaseShape{},
					EnumFacets:    EnumFacets{},
					FormatFacets:  FormatFacets{},
					IntegerFacets: IntegerFacets{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "minimum constraint violation",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{Minimum: func() *float64 {
					f := 5.0
					return &f
				}()},
			},
			args: args{
				source: &NumberShape{
					BaseShape:    BaseShape{},
					EnumFacets:   EnumFacets{},
					FormatFacets: FormatFacets{},
					NumberFacets: NumberFacets{Minimum: func() *float64 {
						f := 10.0
						return &f
					}()},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "maximum constraint violation",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{Maximum: func() *float64 {
					f := 150.0
					return &f
				}()},
			},
			args: args{
				source: &NumberShape{
					BaseShape:    BaseShape{},
					EnumFacets:   EnumFacets{},
					FormatFacets: FormatFacets{},
					NumberFacets: NumberFacets{Maximum: func() *float64 {
						f := 100.0
						return &f
					}()},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "enum constraint violation",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{Enum: []*Node{{Value: 1.0}, {Value: 2.0}}},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				source: &NumberShape{
					BaseShape:    BaseShape{},
					EnumFacets:   EnumFacets{Enum: []*Node{{Value: 3.0}}},
					FormatFacets: FormatFacets{},
					NumberFacets: NumberFacets{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "format constraint violation",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				FormatFacets: FormatFacets{Format: func() *string {
					s := "float"
					return &s
				}()},
				NumberFacets: NumberFacets{},
			},
			args: args{
				source: &NumberShape{
					BaseShape:  BaseShape{},
					EnumFacets: EnumFacets{},
					FormatFacets: FormatFacets{Format: func() *string {
						s := "double"
						return &s
					}()},
					NumberFacets: NumberFacets{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "multipleOf validation",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{MultipleOf: func() *float64 {
					f := 2.0
					return &f
				}()},
			},
			args: args{
				source: &NumberShape{
					BaseShape:    BaseShape{},
					EnumFacets:   EnumFacets{},
					FormatFacets: FormatFacets{},
					NumberFacets: NumberFacets{MultipleOf: func() *float64 {
						f := 2.0
						return &f
					}()},
				},
			},
			want: &NumberShape{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{MultipleOf: func() *float64 {
					f := 2.0
					return &f
				}()},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NumberShape{
				BaseShape:    tt.fields.BaseShape,
				EnumFacets:   tt.fields.EnumFacets,
				FormatFacets: tt.fields.FormatFacets,
				NumberFacets: tt.fields.NumberFacets,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Inherit() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNumberShape_Check(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		EnumFacets   EnumFacets
		FormatFacets FormatFacets
		NumberFacets NumberFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid minimum and maximum",
			fields: fields{
				NumberFacets: NumberFacets{
					Minimum: func() *float64 {
						f := 1.0
						return &f
					}(),
					Maximum: func() *float64 {
						f := 1.0
						return &f
					}(),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid minimum and maximum",
			fields: fields{
				NumberFacets: NumberFacets{
					Minimum: func() *float64 {
						f := 2.0
						return &f
					}(),
					Maximum: func() *float64 {
						f := 1.0
						return &f
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "valid enum values",
			fields: fields{
				EnumFacets: EnumFacets{
					Enum: Nodes{
						{Value: 1.0},
						{Value: 2.0},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid enum values",
			fields: fields{
				EnumFacets: EnumFacets{
					Enum: Nodes{
						{Value: "invalid"},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NumberShape{
				BaseShape:    tt.fields.BaseShape,
				EnumFacets:   tt.fields.EnumFacets,
				FormatFacets: tt.fields.FormatFacets,
				NumberFacets: tt.fields.NumberFacets,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumberShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		EnumFacets   EnumFacets
		FormatFacets FormatFacets
		NumberFacets NumberFacets
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid YAML nodes",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "key1"},
					{Kind: yaml.ScalarNode, Value: "value1"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid YAML nodes",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "key1"},
					{Kind: yaml.MappingNode, Value: "{"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty YAML nodes",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{},
			},
			wantErr: false,
		},
		{
			name: "odd number of YAML nodes",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "key1"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet minimum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "minimum"},
					{Kind: yaml.ScalarNode, Value: "1.0"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet minimum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "minimum"},
					{Kind: yaml.ScalarNode, Value: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet maximum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "maximum"},
					{Kind: yaml.ScalarNode, Value: "10.0"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet maximum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "maximum"},
					{Kind: yaml.ScalarNode, Value: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet multipleOf",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "multipleOf"},
					{Kind: yaml.ScalarNode, Value: "2.0"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet multipleOf",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "multipleOf"},
					{Kind: yaml.ScalarNode, Value: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet format",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "format"},
					{Kind: yaml.ScalarNode, Value: "double"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet format",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "format"},
					{Kind: yaml.ScalarNode, Value: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet enum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "enum"},
					{Kind: yaml.SequenceNode, Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "1.0"},
						{Kind: yaml.ScalarNode, Value: "2.0"},
					}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet enum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				FormatFacets: FormatFacets{},
				NumberFacets: NumberFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "enum"},
					{Kind: yaml.ScalarNode, Value: "invalid"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NumberShape{
				BaseShape:    tt.fields.BaseShape,
				EnumFacets:   tt.fields.EnumFacets,
				FormatFacets: tt.fields.FormatFacets,
				NumberFacets: tt.fields.NumberFacets,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		EnumFacets   EnumFacets
		StringFacets StringFacets
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid string",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v:   "valid",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "valid enum value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{Enum: []*Node{{Value: "valid"}}},
				StringFacets: StringFacets{},
			},
			args: args{
				v:   "valid",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid enum value",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{Enum: []*Node{{Value: "valid"}}},
				StringFacets: StringFacets{},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "valid pattern",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{Pattern: func() *regexp.Regexp {
					p, err := regexp.Compile("^[a-z]+$")
					if err != nil {
						t.Error(err)
					}
					return p
				}()},
			},
			args: args{
				v:   "valid",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid pattern",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{Pattern: func() *regexp.Regexp {
					p, err := regexp.Compile("^[a-z]+$")
					if err != nil {
						t.Error(err)
					}
					return p
				}()},
			},
			args: args{
				v:   "INVALID",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "valid min length",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{
					LengthFacets: LengthFacets{
						MinLength: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			args: args{
				v:   "valid",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid min length",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{
					LengthFacets: LengthFacets{
						MinLength: func() *uint64 {
							i := uint64(6)
							return &i
						}(),
					},
				},
			},
			args: args{
				v:   "short",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "valid max length",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{
					LengthFacets: LengthFacets{
						MaxLength: func() *uint64 {
							i := uint64(5)
							return &i
						}(),
					},
				},
			},
			args: args{
				v:   "valid",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid max length",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{
					LengthFacets: LengthFacets{
						MaxLength: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			args: args{
				v:   "too long",
				in1: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StringShape{
				BaseShape:    tt.fields.BaseShape,
				EnumFacets:   tt.fields.EnumFacets,
				StringFacets: tt.fields.StringFacets,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		EnumFacets   EnumFacets
		StringFacets StringFacets
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Shape
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape:    BaseShape{Type: "string"},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				source: &StringShape{
					BaseShape:    BaseShape{Type: "string"},
					EnumFacets:   EnumFacets{},
					StringFacets: StringFacets{},
				},
			},
			want: &StringShape{
				BaseShape:    BaseShape{Type: "string"},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape:    BaseShape{Type: "string"},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				source: &NumberShape{
					BaseShape:    BaseShape{Type: "number"},
					EnumFacets:   EnumFacets{},
					FormatFacets: FormatFacets{},
					NumberFacets: NumberFacets{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "min length constraint violation",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{
					LengthFacets: LengthFacets{
						MinLength: func() *uint64 {
							i := uint64(2)
							return &i
						}(),
					},
				},
			},
			args: args{
				source: &StringShape{
					BaseShape:  BaseShape{Type: "string"},
					EnumFacets: EnumFacets{},
					StringFacets: StringFacets{
						LengthFacets: LengthFacets{
							MinLength: func() *uint64 {
								i := uint64(4)
								return &i
							}(),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "max length constraint violation",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{
					LengthFacets: LengthFacets{
						MaxLength: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			args: args{
				source: &StringShape{
					BaseShape:  BaseShape{Type: "string"},
					EnumFacets: EnumFacets{},
					StringFacets: StringFacets{
						LengthFacets: LengthFacets{
							MaxLength: func() *uint64 {
								i := uint64(2)
								return &i
							}(),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "enum constraint violation",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{Enum: []*Node{{Value: "valid"}}},
				StringFacets: StringFacets{},
			},
			args: args{
				source: &StringShape{
					BaseShape:    BaseShape{Type: "string"},
					EnumFacets:   EnumFacets{Enum: []*Node{{Value: "invalid"}}},
					StringFacets: StringFacets{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StringShape{
				BaseShape:    tt.fields.BaseShape,
				EnumFacets:   tt.fields.EnumFacets,
				StringFacets: tt.fields.StringFacets,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Inherit() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringShape_Check(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		EnumFacets   EnumFacets
		StringFacets StringFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid shape",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			wantErr: false,
		},
		{
			name: "min/max length constraint violation",
			fields: fields{
				BaseShape:  BaseShape{},
				EnumFacets: EnumFacets{},
				StringFacets: StringFacets{
					LengthFacets: LengthFacets{
						MinLength: func() *uint64 {
							i := uint64(5)
							return &i
						}(),
						MaxLength: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "enum constraint violation",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{Enum: []*Node{{Value: 1}}},
				StringFacets: StringFacets{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StringShape{
				BaseShape:    tt.fields.BaseShape,
				EnumFacets:   tt.fields.EnumFacets,
				StringFacets: tt.fields.StringFacets,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		EnumFacets   EnumFacets
		StringFacets StringFacets
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "empty YAML nodes",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{},
			},
			wantErr: false,
		},
		{
			name: "odd number of YAML nodes",
			fields: fields{
				BaseShape:    BaseShape{},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "minLength"}, {Value: "1", Kind: yaml.ScalarNode, Tag: "!!float"},
					{Value: "maxLength"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet minLength",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "minLength"}, {Value: "1", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet minLength",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "minLength"}, {Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet maxLength",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "maxLength"}, {Value: "1", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet maxLength",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "maxLength"}, {Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid facet pattern",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "pattern"}, {Value: "[a-z]", Kind: yaml.ScalarNode, Tag: "!!str"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid facet pattern",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "pattern"}, {Value: "?$", Kind: yaml.ScalarNode, Tag: "!!str"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid facet pattern tag",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "pattern"}, {Value: "?$", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid enum",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "enum"}, {Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: true,
		},
		{
			name: "unknown facet",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
				EnumFacets:   EnumFacets{},
				StringFacets: StringFacets{},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "unknown"}, {Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StringShape{
				BaseShape:    tt.fields.BaseShape,
				EnumFacets:   tt.fields.EnumFacets,
				StringFacets: tt.fields.StringFacets,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		LengthFacets LengthFacets
		FileFacets   FileFacets
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid file shape",
			fields: fields{
				BaseShape:    BaseShape{},
				LengthFacets: LengthFacets{},
				FileFacets:   FileFacets{},
			},
			args: args{
				v:   "valid_file",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape:    BaseShape{},
				LengthFacets: LengthFacets{},
				FileFacets:   FileFacets{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "valid length facets",
			fields: fields{
				BaseShape: BaseShape{},
				LengthFacets: LengthFacets{
					MinLength: func() *uint64 {
						i := uint64(10)
						return &i
					}(),
					MaxLength: func() *uint64 {
						i := uint64(10)
						return &i
					}(),
				},
				FileFacets: FileFacets{},
			},
			args: args{
				v:   "valid_file",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid length facets maxLength",
			fields: fields{
				BaseShape: BaseShape{},
				LengthFacets: LengthFacets{
					MaxLength: func() *uint64 {
						i := uint64(5)
						return &i
					}(),
				},
				FileFacets: FileFacets{},
			},
			args: args{
				v:   "valid_file",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid length facets minLength",
			fields: fields{
				BaseShape: BaseShape{},
				LengthFacets: LengthFacets{
					MinLength: func() *uint64 {
						i := uint64(5)
						return &i
					}(),
				},
				FileFacets: FileFacets{},
			},
			args: args{
				v:   "v",
				in1: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FileShape{
				BaseShape:    tt.fields.BaseShape,
				LengthFacets: tt.fields.LengthFacets,
				FileFacets:   tt.fields.FileFacets,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		LengthFacets LengthFacets
		FileFacets   FileFacets
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Shape
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape:    BaseShape{Type: "file"},
				LengthFacets: LengthFacets{},
				FileFacets:   FileFacets{},
			},
			args: args{
				source: &FileShape{
					BaseShape:    BaseShape{Type: "file"},
					LengthFacets: LengthFacets{},
					FileFacets:   FileFacets{},
				},
			},
			want: &FileShape{
				BaseShape:    BaseShape{Type: "file"},
				LengthFacets: LengthFacets{},
				FileFacets:   FileFacets{},
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape:    BaseShape{Type: "file"},
				LengthFacets: LengthFacets{},
				FileFacets:   FileFacets{},
			},
			args: args{
				source: &StringShape{
					BaseShape:    BaseShape{Type: "string"},
					EnumFacets:   EnumFacets{},
					StringFacets: StringFacets{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid length facets",
			fields: fields{
				BaseShape: BaseShape{Type: "file"},
				LengthFacets: LengthFacets{
					MinLength: func() *uint64 {
						i := uint64(5)
						return &i
					}(),
					MaxLength: func() *uint64 {
						i := uint64(10)
						return &i
					}(),
				},
				FileFacets: FileFacets{},
			},
			args: args{
				source: &FileShape{
					BaseShape: BaseShape{Type: "file"},
					LengthFacets: LengthFacets{
						MinLength: func() *uint64 {
							i := uint64(5)
							return &i
						}(),
						MaxLength: func() *uint64 {
							i := uint64(10)
							return &i
						}(),
					},
					FileFacets: FileFacets{},
				},
			},
			want: &FileShape{
				BaseShape: BaseShape{Type: "file"},
				LengthFacets: LengthFacets{
					MinLength: func() *uint64 {
						i := uint64(5)
						return &i
					}(),
					MaxLength: func() *uint64 {
						i := uint64(10)
						return &i
					}(),
				},
				FileFacets: FileFacets{},
			},
			wantErr: false,
		},
		{
			name: "min length constraint violation",
			fields: fields{
				BaseShape: BaseShape{Type: "file"},
				LengthFacets: LengthFacets{
					MinLength: func() *uint64 {
						i := uint64(5)
						return &i
					}(),
				},
				FileFacets: FileFacets{},
			},
			args: args{
				source: &FileShape{
					BaseShape: BaseShape{Type: "file"},
					LengthFacets: LengthFacets{
						MinLength: func() *uint64 {
							i := uint64(6)
							return &i
						}(),
					},
					FileFacets: FileFacets{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "max length constraint violation",
			fields: fields{
				BaseShape: BaseShape{Type: "file"},
				LengthFacets: LengthFacets{
					MaxLength: func() *uint64 {
						i := uint64(5)
						return &i
					}(),
				},
				FileFacets: FileFacets{},
			},
			args: args{
				source: &FileShape{
					BaseShape: BaseShape{Type: "file"},
					LengthFacets: LengthFacets{
						MaxLength: func() *uint64 {
							i := uint64(3)
							return &i
						}(),
					},
					FileFacets: FileFacets{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "file types are incompatible",
			fields: fields{
				BaseShape:    BaseShape{Type: "file"},
				LengthFacets: LengthFacets{},
				FileFacets: FileFacets{
					FileTypes: Nodes{
						{Value: "image/png"},
					},
				},
			},
			args: args{
				source: &FileShape{
					BaseShape:    BaseShape{Type: "file"},
					LengthFacets: LengthFacets{},
					FileFacets: FileFacets{
						FileTypes: Nodes{
							{Value: "image/jpeg"},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FileShape{
				BaseShape:    tt.fields.BaseShape,
				LengthFacets: tt.fields.LengthFacets,
				FileFacets:   tt.fields.FileFacets,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Inherit() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileShape_Check(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		LengthFacets LengthFacets
		FileFacets   FileFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid file shape",
			fields: fields{
				BaseShape: BaseShape{},
				LengthFacets: LengthFacets{
					MinLength: func() *uint64 {
						i := uint64(5)
						return &i
					}(),
					MaxLength: func() *uint64 {
						i := uint64(6)
						return &i
					}(),
				},
				FileFacets: FileFacets{
					FileTypes: Nodes{
						{Value: "image/png"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minLength must be less than or equal to maxLength",
			fields: fields{
				BaseShape: BaseShape{},
				LengthFacets: LengthFacets{
					MinLength: func() *uint64 {
						i := uint64(5)
						return &i
					}(),
					MaxLength: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "file type must be string",
			fields: fields{
				BaseShape:    BaseShape{},
				LengthFacets: LengthFacets{},
				FileFacets: FileFacets{
					FileTypes: Nodes{
						{Value: 1},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FileShape{
				BaseShape:    tt.fields.BaseShape,
				LengthFacets: tt.fields.LengthFacets,
				FileFacets:   tt.fields.FileFacets,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		LengthFacets LengthFacets
		FileFacets   FileFacets
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case with all facets",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "minLength"},
					{Value: "1", Kind: yaml.ScalarNode, Tag: "!!float"},
					{Value: "maxLength"},
					{Value: "10", Kind: yaml.ScalarNode, Tag: "!!float"},
					{Value: "fileTypes"},
					{
						Kind: yaml.SequenceNode, Content: []*yaml.Node{
							{Value: "image/png", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
					{
						Value: "custom", Kind: yaml.ScalarNode, Tag: "!!str",
					},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "key", Kind: yaml.ScalarNode, Tag: "!!str"},
							{Value: "value", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "decode minLength error",
			fields: fields{},
			args: args{
				v: []*yaml.Node{
					{Value: "minLength"},
					{Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: true,
		},
		{
			name:   "decode maxLength error",
			fields: fields{},
			args: args{
				v: []*yaml.Node{
					{Value: "maxLength"},
					{Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: true,
		},
		{
			name: "fileTypes must be sequence node",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "fileTypes"},
					{Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!str"},
				},
			},
			wantErr: true,
		},
		{
			name: "member of fileTypes must be string",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "fileTypes"},
					{
						Kind: yaml.SequenceNode, Content: []*yaml.Node{
							{Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!float"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "make node fileTypes error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "fileTypes"},
					{
						Kind: yaml.SequenceNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "odd number of YAML nodes error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "unknown"},
				},
			},
			wantErr: true,
		},
		{
			name: "unknown facet error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "unknown"}, {Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!float"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FileShape{
				BaseShape:    tt.fields.BaseShape,
				LengthFacets: tt.fields.LengthFacets,
				FileFacets:   tt.fields.FileFacets,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBooleanShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape  BaseShape
		EnumFacets EnumFacets
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid boolean shape",
			fields: fields{
				BaseShape: BaseShape{},
				EnumFacets: EnumFacets{
					Enum: []*Node{
						{Value: true},
					},
				},
			},
			args: args{
				v:   true,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid enum",
			fields: fields{
				BaseShape: BaseShape{},
				EnumFacets: EnumFacets{
					Enum: []*Node{
						{Value: "true"},
					},
				},
			},
			args: args{
				v: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BooleanShape{
				BaseShape:  tt.fields.BaseShape,
				EnumFacets: tt.fields.EnumFacets,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBooleanShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape  BaseShape
		EnumFacets EnumFacets
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got Shape) (string, bool)
		wantErr bool
	}{
		{
			name: "positive case with all facets",
			fields: fields{
				BaseShape: BaseShape{
					Type: "boolean",
				},
				EnumFacets: EnumFacets{
					Enum: []*Node{
						{Value: true},
					},
				},
			},
			args: args{
				source: &BooleanShape{
					BaseShape: BaseShape{
						Type: "boolean",
					},
					EnumFacets: EnumFacets{
						Enum: []*Node{
							{Value: true},
							{Value: false},
						},
					},
				},
			},
			want: func(got Shape) (string, bool) {
				gotBool, ok := got.(*BooleanShape)
				if !ok {
					return "unexpected type", false
				}
				if len(gotBool.EnumFacets.Enum) != 1 {
					return "unexpected enum length", false
				}
				if gotBool.EnumFacets.Enum[0].Value != true {
					return "unexpected enum value", false
				}
				return "", true
			},
			wantErr: false,
		},
		{
			name: "positive case with nil enum",
			fields: fields{
				BaseShape: BaseShape{
					Type: "boolean",
				},
			},
			args: args{
				source: &BooleanShape{
					BaseShape: BaseShape{
						Type: "boolean",
					},
					EnumFacets: EnumFacets{
						Enum: []*Node{
							{Value: true},
							{Value: false},
						},
					},
				},
			},
			want: func(got Shape) (string, bool) {
				gotBool, ok := got.(*BooleanShape)
				if !ok {
					return "unexpected type", false
				}
				if len(gotBool.EnumFacets.Enum) != 2 {
					return "unexpected enum length", false
				}
				if gotBool.EnumFacets.Enum[0].Value != true {
					return "unexpected enum value", false
				}
				if gotBool.EnumFacets.Enum[1].Value != false {
					return "unexpected enum value", false
				}
				return "", true
			},
			wantErr: false,
		},
		{
			name: "cannot inherit from different type",
			fields: fields{
				BaseShape: BaseShape{
					Type: "boolean",
				},
			},
			args: args{
				source: &StringShape{
					BaseShape: BaseShape{
						Type: "string",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "enum constraint violation",
			fields: fields{
				BaseShape: BaseShape{
					Type: "boolean",
				},
				EnumFacets: EnumFacets{
					Enum: []*Node{
						{Value: true},
					},
				},
			},
			args: args{
				source: &BooleanShape{
					BaseShape: BaseShape{
						Type: "boolean",
					},
					EnumFacets: EnumFacets{
						Enum: []*Node{
							{Value: false},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BooleanShape{
				BaseShape:  tt.fields.BaseShape,
				EnumFacets: tt.fields.EnumFacets,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) == tt.wantErr {
				return
			}
			if message, passed := tt.want(got); !passed {
				t.Errorf("Case hasn't been passed: %s", message)
			}
		})
	}
}

func TestBooleanShape_Check(t *testing.T) {
	type fields struct {
		BaseShape  BaseShape
		EnumFacets EnumFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid shape",
			fields: fields{
				BaseShape: BaseShape{},
				EnumFacets: EnumFacets{
					Enum: []*Node{
						{Value: true},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "enum value must be boolean",
			fields: fields{
				BaseShape: BaseShape{},
				EnumFacets: EnumFacets{
					Enum: []*Node{
						{Value: 1},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BooleanShape{
				BaseShape:  tt.fields.BaseShape,
				EnumFacets: tt.fields.EnumFacets,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBooleanShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape  BaseShape
		EnumFacets EnumFacets
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case with all facets",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "enum"},
					{
						Kind: yaml.SequenceNode, Content: []*yaml.Node{
							{Value: "true", Kind: yaml.ScalarNode, Tag: "!!str"},
							{Value: "false", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
					{
						Value: "custom", Kind: yaml.ScalarNode, Tag: "!!str",
					},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "key", Kind: yaml.ScalarNode, Tag: "!!str"},
							{Value: "value", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "make enum error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "enum"},
					{
						Kind: yaml.SequenceNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "make node error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "custom"},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BooleanShape{
				BaseShape:  tt.fields.BaseShape,
				EnumFacets: tt.fields.EnumFacets,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateTimeShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		FormatFacets FormatFacets
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case with rfc3339",
			fields: fields{
				BaseShape: BaseShape{},
				FormatFacets: FormatFacets{
					Format: func() *string {
						s := "rfc3339"
						return &s
					}(),
				},
			},
			args: args{
				v:   "2021-01-01T00:00:00Z",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "positive case with rfc2616",
			fields: fields{
				BaseShape: BaseShape{},
				FormatFacets: FormatFacets{
					Format: func() *string {
						s := "rfc2616"
						return &s
					}(),
				},
			},
			args: args{
				v:   "Sun, 06 Nov 1994 08:49:37 GMT",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "positive case without format",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "2021-01-01T00:00:00Z",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid value",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid case with rfc3339",
			fields: fields{
				BaseShape: BaseShape{},
				FormatFacets: FormatFacets{
					Format: func() *string {
						s := "rfc3339"
						return &s
					}(),
				},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid case with rfc2616",
			fields: fields{
				BaseShape: BaseShape{},
				FormatFacets: FormatFacets{
					Format: func() *string {
						s := "rfc2616"
						return &s
					}(),
				},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateTimeShape{
				BaseShape:    tt.fields.BaseShape,
				FormatFacets: tt.fields.FormatFacets,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateTimeShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		FormatFacets FormatFacets
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got Shape) (string, bool)
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape: BaseShape{Type: "datetime"},
			},
			args: args{
				source: &DateTimeShape{
					BaseShape: BaseShape{Type: "datetime"},
					FormatFacets: FormatFacets{
						Format: func() *string {
							s := "rfc3339"
							return &s
						}(),
					},
				},
			},
			want: func(got Shape) (string, bool) {
				gotDateTime, ok := got.(*DateTimeShape)
				if !ok {
					return "unexpected type", false
				}
				if gotDateTime.FormatFacets.Format == nil {
					return "format is nil", false
				}
				if *gotDateTime.FormatFacets.Format != "rfc3339" {
					return "unexpected format", false
				}
				return "", true
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape: BaseShape{Type: "datetime"},
			},
			args: args{
				source: &StringShape{
					BaseShape: BaseShape{Type: "string"},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "format contraint violation",
			fields: fields{
				BaseShape: BaseShape{Type: "datetime"},
				FormatFacets: FormatFacets{
					Format: func() *string {
						s := "rfc3339"
						return &s
					}(),
				},
			},
			args: args{
				source: &DateTimeShape{
					BaseShape: BaseShape{Type: "datetime"},
					FormatFacets: FormatFacets{
						Format: func() *string {
							s := "rfc2616"
							return &s
						}(),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateTimeShape{
				BaseShape:    tt.fields.BaseShape,
				FormatFacets: tt.fields.FormatFacets,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) == tt.wantErr {
				return
			}
			if message, passed := tt.want(got); !passed {
				t.Errorf("Case hasn't been passed: %s", message)
			}
		})
	}
}

func TestDateTimeShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		FormatFacets FormatFacets
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case with all facets",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "format"},
					{Value: "rfc3339", Kind: yaml.ScalarNode, Tag: "!!str"},
					{Value: "custom"},
					{Value: "value", Kind: yaml.ScalarNode, Tag: "!!str"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "format"},
					{Value: "invalid", Kind: yaml.ScalarNode, Tag: "!!str"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing value",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "format"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid decode value",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "format"},
					{Value: "rfc3339", Kind: yaml.ScalarNode, Tag: "!!int"},
				},
			},
			wantErr: true,
		},
		{
			name: "make node error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "custom"},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateTimeShape{
				BaseShape:    tt.fields.BaseShape,
				FormatFacets: tt.fields.FormatFacets,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateTimeShape_Check(t *testing.T) {
	type fields struct {
		BaseShape    BaseShape
		FormatFacets FormatFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid shape",
			fields: fields{
				BaseShape: BaseShape{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateTimeShape{
				BaseShape:    tt.fields.BaseShape,
				FormatFacets: tt.fields.FormatFacets,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateTimeOnlyShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "2021-01-01T00:00:00",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid value",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateTimeOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateTimeOnlyShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got Shape) (string, bool)
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape: BaseShape{Type: "datetime-only"},
			},
			args: args{
				source: &DateTimeOnlyShape{
					BaseShape: BaseShape{Type: "datetime-only"},
				},
			},
			want: func(got Shape) (string, bool) {
				gotDateTime, ok := got.(*DateTimeOnlyShape)
				if !ok {
					return "unexpected type", false
				}
				if gotDateTime.Type != "datetime-only" {
					return "unexpected type", false

				}
				return "", true
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape: BaseShape{Type: "datetime-only"},
			},
			args: args{
				source: &StringShape{
					BaseShape: BaseShape{Type: "string"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateTimeOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) == tt.wantErr {
				return
			}
			if message, passed := tt.want(got); !passed {
				t.Errorf("Case hasn't been passed: %s", message)
			}
		})
	}
}

func TestDateTimeOnlyShape_Check(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid shape",
			fields: fields{
				BaseShape: BaseShape{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateTimeOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateTimeOnlyShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case with all facets",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "custom", Kind: yaml.ScalarNode, Tag: "!!str",
					},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "key", Kind: yaml.ScalarNode, Tag: "!!str"},
							{Value: "value", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "odd number of YAML nodes error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "unknown"},
				},
			},
			wantErr: true,
		},
		{
			name: "make node error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "custom"},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateTimeOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateOnlyShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "2021-01-01",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid value",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateOnlyShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got Shape) (string, bool)
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape: BaseShape{Type: "date-only"},
			},
			args: args{
				source: &DateOnlyShape{
					BaseShape: BaseShape{Type: "date-only"},
				},
			},
			want: func(got Shape) (string, bool) {
				gotDate, ok := got.(*DateOnlyShape)
				if !ok {
					return "unexpected type", false
				}
				if gotDate.Type != "date-only" {
					return "unexpected type", false
				}
				return "", true
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape: BaseShape{Type: "date-only"},
			},
			args: args{
				source: &StringShape{
					BaseShape: BaseShape{Type: "string"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) == tt.wantErr {
				return
			}
			if message, passed := tt.want(got); !passed {
				t.Errorf("Case hasn't been passed: %s", message)
			}
		})
	}
}

func TestDateOnlyShape_Check(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid shape",
			fields: fields{
				BaseShape: BaseShape{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateOnlyShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case with all facets",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "custom", Kind: yaml.ScalarNode, Tag: "!!str",
					},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "key", Kind: yaml.ScalarNode, Tag: "!!str"},
							{Value: "value", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "make node error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "custom"},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DateOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimeOnlyShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "00:00:00",
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   123,
				in1: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid value",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TimeOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimeOnlyShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got Shape) (string, bool)
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape: BaseShape{Type: "time-only"},
			},
			args: args{
				source: &TimeOnlyShape{
					BaseShape: BaseShape{Type: "time-only"},
				},
			},
			want: func(got Shape) (string, bool) {
				gotTime, ok := got.(*TimeOnlyShape)
				if !ok {
					return "unexpected type", false
				}
				if gotTime.Type != "time-only" {
					return "unexpected type", false
				}
				return "", true
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape: BaseShape{Type: "time-only"},
			},
			args: args{
				source: &StringShape{
					BaseShape: BaseShape{Type: "string"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TimeOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) == tt.wantErr {
				return
			}
			if message, passed := tt.want(got); !passed {
				t.Errorf("Case hasn't been passed: %s", message)
			}
		})
	}
}

func TestTimeOnlyShape_Check(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid shape",
			fields: fields{
				BaseShape: BaseShape{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TimeOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimeOnlyShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case with all facets",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "custom", Kind: yaml.ScalarNode, Tag: "!!str",
					},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "key", Kind: yaml.ScalarNode, Tag: "!!str"},
							{Value: "value", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "odd number of YAML nodes error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "unknown"},
				},
			},
			wantErr: true,
		},
		{
			name: "make node error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "custom"},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TimeOnlyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAnyShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		in0 interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				in0: "test",
				in1: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AnyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Validate(tt.args.in0, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAnyShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got Shape) (string, bool)
		wantErr bool
	}{
		{
			name: "inherit from same type",
			fields: fields{
				BaseShape: BaseShape{Type: "any"},
			},
			args: args{
				source: &AnyShape{
					BaseShape: BaseShape{Type: "any"},
				},
			},
			want: func(got Shape) (string, bool) {
				gotAny, ok := got.(*AnyShape)
				if !ok {
					return "unexpected type", false
				}
				if gotAny.Type != "any" {
					return "unexpected type", false
				}
				return "", true
			},
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape: BaseShape{Type: "any"},
			},
			args: args{
				source: &StringShape{
					BaseShape: BaseShape{Type: "string"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AnyShape{
				BaseShape: tt.fields.BaseShape,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) == tt.wantErr {
				return
			}
			if tt.want != nil {
				if message, passed := tt.want(got); !passed {
					t.Errorf("Case hasn't been passed: %s", message)
				}
			}
		})
	}
}

func TestAnyShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case with all facets",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "custom", Kind: yaml.ScalarNode, Tag: "!!str",
					},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "key", Kind: yaml.ScalarNode, Tag: "!!str"},
							{Value: "value", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "odd number of YAML nodes error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "unknown"},
				},
			},
			wantErr: true,
		},
		{
			name: "make node error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "custom"},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AnyShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNilShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v   interface{}
		in1 string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   nil,
				in1: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid value",
			fields: fields{
				BaseShape: BaseShape{},
			},
			args: args{
				v:   "invalid",
				in1: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NilShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Validate(tt.args.v, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNilShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		source Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got Shape) (string, bool)
		wantErr bool
	}{
		{
			name: "inherit from nil",
			fields: fields{
				BaseShape: BaseShape{Type: "nil"},
			},
			args: args{
				source: &NilShape{
					BaseShape: BaseShape{Type: "nil"},
				},
			},
			want: func(got Shape) (string, bool) {
				gotNil, ok := got.(*NilShape)
				if !ok {
					return "unexpected type", false
				}
				if gotNil.Type != "nil" {
					return "unexpected type", false
				}
				return "", true
			},
			wantErr: false,
		},
		{
			name: "inherit from different type",
			fields: fields{
				BaseShape: BaseShape{Type: "nil"},
			},
			args: args{
				source: &StringShape{
					BaseShape: BaseShape{Type: "string"},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NilShape{
				BaseShape: tt.fields.BaseShape,
			}
			got, err := s.Inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if message, passed := tt.want(got); !passed {
					t.Errorf("Case hasn't been passed: %s", message)
				}
			}
		})
	}
}

func TestNilShape_Check(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid shape",
			fields: fields{
				BaseShape: BaseShape{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NilShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNilShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape BaseShape
	}
	type args struct {
		v []*yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "custom", Kind: yaml.ScalarNode, Tag: "!!str",
					},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "key", Kind: yaml.ScalarNode, Tag: "!!str"},
							{Value: "value", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "odd number of YAML nodes error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "unknown"},
				},
			},
			wantErr: true,
		},
		{
			name: "make node error",
			fields: fields{
				BaseShape: BaseShape{
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{Value: "custom"},
					{
						Kind: yaml.MappingNode, Content: []*yaml.Node{
							{Value: "{", Kind: yaml.ScalarNode, Tag: "!!str"},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NilShape{
				BaseShape: tt.fields.BaseShape,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
