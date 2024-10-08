package raml

import (
	"container/list"
	"context"
	"regexp"
	"testing"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

func TestArrayShape_clone(t *testing.T) {
	type fields struct {
		BaseShape   *BaseShape
		ArrayFacets ArrayFacets
	}
	type args struct {
		base      *BaseShape
		clonedMap map[int64]*BaseShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(got Shape) (string, bool)
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape:   &BaseShape{},
				ArrayFacets: ArrayFacets{},
			},
			args: args{
				base:      &BaseShape{},
				clonedMap: make(map[int64]*BaseShape),
			},
			want: func(got Shape) (string, bool) {
				if _, ok := got.(*ArrayShape); !ok {
					return "expected to get *ArrayShape", false
				}
				return "", true
			},
		},
		{
			name: "positive case with history",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ArrayFacets: ArrayFacets{},
			},
			args: args{
				base: &BaseShape{
					ID: 1,
					Shape: &ArrayShape{
						BaseShape:   &BaseShape{ID: 1},
						ArrayFacets: ArrayFacets{},
					},
				},
				clonedMap: make(map[int64]*BaseShape),
			},
			want: func(got Shape) (string, bool) {
				if _, ok := got.(*ArrayShape); !ok {
					return "expected to get *ArrayShape", false
				}
				return "", true
			},
		},
		{
			name: "positive case with items",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					Items: &BaseShape{
						ID: 1,
						Shape: &StringShape{
							BaseShape: &BaseShape{ID: 2},
						},
					},
				},
			},
			args: args{
				clonedMap: make(map[int64]*BaseShape),
			},
			want: func(got Shape) (string, bool) {
				if _, ok := got.(*ArrayShape); !ok {
					return "expected to get *ArrayShape", false
				}
				return "", true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArrayShape{
				BaseShape:   tt.fields.BaseShape,
				ArrayFacets: tt.fields.ArrayFacets,
			}
			got := s.clone(tt.args.base, tt.args.clonedMap)
			if msg, ok := tt.want(got); !ok {
				t.Errorf("Case hasn't been passed: %s", msg)
			}
		})
	}
}

func TestArrayShape_Validate(t *testing.T) {
	type fields struct {
		BaseShape   *BaseShape
		ArrayFacets ArrayFacets
	}
	type args struct {
		v       interface{}
		ctxPath string
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
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					Items: &BaseShape{
						Shape: &StringShape{
							BaseShape: &BaseShape{ID: 2},
						},
					},
					UniqueItems: func() *bool {
						b := true
						return &b
					}(),
				},
			},
			args: args{
				v:       []interface{}{"test"},
				ctxPath: "",
			},
		},
		{
			name: "invalid type",
			fields: fields{
				BaseShape: &BaseShape{},
			},
			args: args{
				v: "test",
			},
			wantErr: true,
		},
		{
			name: "array must have at least two items",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					MinItems: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			args: args{
				v: []interface{}{"test"},
			},
			wantErr: true,
		},
		{
			name: "array must have no more than two items",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					MaxItems: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			args: args{
				v: []interface{}{"test", "test", "test"},
			},
			wantErr: true,
		},
		{
			name: "invalid array item",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					Items: &BaseShape{
						Shape: &StringShape{
							BaseShape: &BaseShape{ID: 2},
						},
					},
				},
			},
			args: args{
				v:       []interface{}{1},
				ctxPath: "",
			},
			wantErr: true,
		},
		{
			name: "array must have unique items",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					Items: &BaseShape{
						Shape: &StringShape{
							BaseShape: &BaseShape{ID: 2},
						},
					},
					UniqueItems: func() *bool {
						b := true
						return &b
					}(),
				},
			},
			args: args{
				v:       []interface{}{"test", "test"},
				ctxPath: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArrayShape{
				BaseShape:   tt.fields.BaseShape,
				ArrayFacets: tt.fields.ArrayFacets,
			}
			if err := s.validate(tt.args.v, tt.args.ctxPath); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArrayShape_Inherit(t *testing.T) {
	type fields struct {
		BaseShape   *BaseShape
		ArrayFacets ArrayFacets
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
			name: "positive case without facets",
			fields: fields{
				BaseShape:   &BaseShape{},
				ArrayFacets: ArrayFacets{},
			},
			args: args{
				source: &ArrayShape{
					BaseShape: &BaseShape{
						ID: 1,
					},
					ArrayFacets: ArrayFacets{
						Items: &BaseShape{
							ID: 1,
							Shape: &StringShape{
								BaseShape: &BaseShape{ID: 2},
							},
						},
						MinItems: func() *uint64 {
							i := uint64(2)
							return &i
						}(),
						MaxItems: func() *uint64 {
							i := uint64(2)
							return &i
						}(),
						UniqueItems: func() *bool {
							b := true
							return &b
						}(),
					},
				},
			},
			want: func(got Shape) (string, bool) {
				arr, ok := got.(*ArrayShape)
				if !ok {
					return "expected to get *ArrayShape", false
				}
				if arr.MinItems == nil || *arr.MinItems != 2 {
					return "MinItems hasn't been inherited", false
				}
				if arr.MaxItems == nil || *arr.MaxItems != 2 {
					return "MaxItems hasn't been inherited", false
				}
				if arr.UniqueItems == nil || *arr.UniqueItems != true {
					return "UniqueItems hasn't been inherited", false
				}
				if arr.Items == nil || arr.Items.ID != 1 {
					return "Items hasn't been inherited", false
				}
				return "", true
			},
		},
		{
			name: "positive case with facets",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					Items: &BaseShape{
						ID: 1,
						Shape: &StringShape{
							BaseShape: &BaseShape{ID: 2},
						},
					},
					MinItems: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
					MaxItems: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
					UniqueItems: func() *bool {
						b := true
						return &b
					}(),
				},
			},
			args: args{
				source: &ArrayShape{
					BaseShape: &BaseShape{
						ID: 1,
					},
					ArrayFacets: ArrayFacets{
						Items: &BaseShape{
							ID: 1,
							Shape: &StringShape{
								BaseShape: &BaseShape{ID: 2},
							},
						},
						MinItems: func() *uint64 {
							i := uint64(3)
							return &i
						}(),
						MaxItems: func() *uint64 {
							i := uint64(3)
							return &i
						}(),
						UniqueItems: func() *bool {
							b := false
							return &b
						}(),
					},
				},
			},
			want: func(got Shape) (string, bool) {
				arr, ok := got.(*ArrayShape)
				if !ok {
					return "expected to get *ArrayShape", false
				}
				if arr.MinItems == nil || *arr.MinItems != 2 {
					return "MinItems hasn't been inherited", false
				}
				if arr.MaxItems == nil || *arr.MaxItems != 4 {
					return "MaxItems hasn't been inherited", false
				}
				if arr.UniqueItems == nil || *arr.UniqueItems != true {
					return "UniqueItems hasn't been inherited", false
				}
				if arr.Items == nil || arr.Items.ID != 1 {
					return "Items hasn't been inherited", false
				}
				return "", true
			},
		},
		{
			name: "negative case different type",
			fields: fields{
				BaseShape:   &BaseShape{},
				ArrayFacets: ArrayFacets{},
			},
			args: args{
				source: &StringShape{
					BaseShape: &BaseShape{
						ID: 1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative case different items type",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					Items: &BaseShape{
						ID: 1,
						Shape: &StringShape{
							BaseShape: &BaseShape{ID: 2},
						},
					},
				},
			},
			args: args{
				source: &ArrayShape{
					BaseShape: &BaseShape{},
					ArrayFacets: ArrayFacets{
						Items: &BaseShape{
							ID: 1,
							Shape: &NumberShape{
								BaseShape: &BaseShape{ID: 2},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "minItems constraint violation",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					MinItems: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			args: args{
				source: &ArrayShape{
					BaseShape: &BaseShape{},
					ArrayFacets: ArrayFacets{
						MinItems: func() *uint64 {
							i := uint64(1)
							return &i
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "maxItems constraint violation",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					MaxItems: func() *uint64 {
						i := uint64(1)
						return &i
					}(),
				},
			},
			args: args{
				source: &ArrayShape{
					BaseShape: &BaseShape{},
					ArrayFacets: ArrayFacets{
						MaxItems: func() *uint64 {
							i := uint64(2)
							return &i
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "uniqueItems constraint violation",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					UniqueItems: func() *bool {
						b := false
						return &b
					}(),
				},
			},
			args: args{
				source: &ArrayShape{
					BaseShape: &BaseShape{},
					ArrayFacets: ArrayFacets{
						UniqueItems: func() *bool {
							b := true
							return &b
						}(),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArrayShape{
				BaseShape:   tt.fields.BaseShape,
				ArrayFacets: tt.fields.ArrayFacets,
			}
			got, err := s.inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestArrayShape_Check(t *testing.T) {
	type fields struct {
		BaseShape   *BaseShape
		ArrayFacets ArrayFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					Items: &BaseShape{
						ID: 1,
						Shape: &StringShape{
							BaseShape: &BaseShape{ID: 2},
						},
					},
					MinItems: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
					MaxItems: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
					UniqueItems: func() *bool {
						b := true
						return &b
					}(),
				},
			},
		},
		{
			name: "invalid min and max items",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					MinItems: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
					MaxItems: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid items",
			fields: fields{
				BaseShape: &BaseShape{},
				ArrayFacets: ArrayFacets{
					Items: &BaseShape{
						ID: 1,
						Shape: &StringShape{
							BaseShape: &BaseShape{ID: 2},
							StringFacets: StringFacets{
								LengthFacets: LengthFacets{
									MinLength: func() *uint64 {
										i := uint64(4)
										return &i
									}(),
									MaxLength: func() *uint64 {
										i := uint64(2)
										return &i
									}(),
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArrayShape{
				BaseShape:   tt.fields.BaseShape,
				ArrayFacets: tt.fields.ArrayFacets,
			}
			if err := s.check(); (err != nil) != tt.wantErr {
				t.Errorf("check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArrayShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape   *BaseShape
		ArrayFacets ArrayFacets
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
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "minItems",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "2",
						Tag:   "!!int",
					},
					{
						Value: "maxItems",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "4",
						Tag:   "!!int",
					},
					{
						Value: "uniqueItems",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "true",
						Tag:   "!!bool",
					},
					{
						Value: "items",
					},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{
								Kind:  yaml.ScalarNode,
								Value: "type",
								Tag:   "!!str",
							},
							{
								Kind:  yaml.ScalarNode,
								Value: "string",
								Tag:   "!!str",
							},
						},
					},
					{
						Value: "custom",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "value",
						Tag:   "!!str",
					},
				},
			},
		},
		{
			name: "invalid odd",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "minItems",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid minItems",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "minItems",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "string",
						Tag:   "!!str",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid maxItems",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "maxItems",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "string",
						Tag:   "!!str",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid uniqueItems",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "uniqueItems",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "string",
						Tag:   "!!str",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid items",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "items",
					},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{
								Kind:  yaml.ScalarNode,
								Value: "type",
								Tag:   "!!int",
							},
							{
								Kind:  yaml.ScalarNode,
								Value: "string",
								Tag:   "!!int",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid custom facet",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "custom",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "value",
						Tag:   "!!int",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArrayShape{
				BaseShape:   tt.fields.BaseShape,
				ArrayFacets: tt.fields.ArrayFacets,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectShape_unmarshalPatternProperties(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		nodeName            string
		propertyName        string
		data                *yaml.Node
		hasImplicitOptional bool
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
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				nodeName:     "patternProperties",
				propertyName: "pattern",
				data: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "type",
							Tag:   "!!str",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "string",
							Tag:   "!!str",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "make pattern property error",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				nodeName:     "patternProperties",
				propertyName: "pattern",
				data: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "required",
							Tag:   "!!str",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "true",
							Tag:   "!!bool",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.unmarshalPatternProperties(tt.args.nodeName, tt.args.propertyName, tt.args.data, tt.args.hasImplicitOptional); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalPatternProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectShape_unmarshalProperty(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		nodeName string
		data     *yaml.Node
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
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				nodeName: "properties",
				data: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "type",
							Tag:   "!!str",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "string",
							Tag:   "!!str",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "positive case: pattern properties",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				nodeName: "//",
				data: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "type",
							Tag:   "!!str",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "string",
							Tag:   "!!str",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "negative case: make property error: decode error",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				nodeName: "properties",
				data: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "required",
							Tag:   "!!bool",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "true",
							Tag:   "!!int",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.unmarshalProperty(tt.args.nodeName, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalProperty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
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
			name: "positive case with all possible facets",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "minProperties",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "2",
						Tag:   "!!int",
					},
					{
						Value: "maxProperties",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "4",
						Tag:   "!!int",
					},
					{
						Value: "additionalProperties",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "false",
						Tag:   "!!bool",
					},
					{
						Value: "discriminator",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "discriminator",
						Tag:   "!!str",
					},
					{
						Value: "discriminatorValue",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "discriminatorValue",
						Tag:   "!!str",
					},
					{
						Value: "properties",
					},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{
								Kind:  yaml.ScalarNode,
								Value: "type",
								Tag:   "!!str",
							},
							{
								Kind:  yaml.ScalarNode,
								Value: "string",
								Tag:   "!!str",
							},
						},
					},
					{
						Value: "custom",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "value",
						Tag:   "!!str",
					},
				},
			},
		},
		{
			name: "invalid odd",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "minProperties",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid minProperties",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "minProperties",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "string",
						Tag:   "!!str",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid maxProperties",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "maxProperties",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "string",
						Tag:   "!!str",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid additionalProperties",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "additionalProperties",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "string",
						Tag:   "!!str",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid discriminator",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "discriminator",
					},
					{
						Kind:  yaml.MappingNode,
						Value: "{",
						Tag:   "!!unknown",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid discriminatorValue",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "discriminatorValue",
					},
					{
						Kind:  100500,
						Value: "{",
						Tag:   "!!int",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid odd number of nodes in properties",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "properties",
					},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{
								Kind:  yaml.ScalarNode,
								Value: "type",
								Tag:   "!!str",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid properties",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "properties",
					},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{
								Kind:  yaml.ScalarNode,
								Value: "//",
								Tag:   "!!str",
							},
							{
								Kind: yaml.MappingNode,
								Content: []*yaml.Node{
									{
										Kind:  yaml.ScalarNode,
										Value: "required",
										Tag:   "!!str",
									},
									{
										Kind:  yaml.ScalarNode,
										Value: "true",
										Tag:   "!!bool",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid custom facet",
			fields: fields{
				BaseShape: &BaseShape{
					raml:              &RAML{},
					CustomShapeFacets: orderedmap.New[string, *Node](0),
				},
			},
			args: args{
				v: []*yaml.Node{
					{
						Value: "custom",
					},
					{
						Kind:  yaml.ScalarNode,
						Value: "value",
						Tag:   "!!int",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectShape_clone(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		base      *BaseShape
		clonedMap map[int64]*BaseShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(got Shape) (string, bool)
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("test", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("test", PatternProperty{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				base: &BaseShape{
					ID: 2,
				},
				clonedMap: map[int64]*BaseShape{},
			},
			want: func(got Shape) (string, bool) {
				obj, ok := got.(*ObjectShape)
				if !ok {
					return "expected to get *ObjectShape", false
				}
				if obj.ID == 1 {
					return "ID hasn't been cloned", false
				}
				return "", true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			got := s.clone(tt.args.base, tt.args.clonedMap)
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			} else {
				t.Errorf("No want function provided")
			}
		})
	}
}

func TestObjectShape_validateProperties(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		ctxPath string
		props   map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case without additional properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
					AdditionalProperties: func() *bool {
						b := false
						return &b
					}(),
				},
			},
			args: args{
				ctxPath: "test",
				props: map[string]interface{}{
					"property": "property",
				},
			},
		},
		{
			name: "positive case with additional property",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					AdditionalProperties: func() *bool {
						b := true
						return &b
					}(),
				},
			},
			args: args{
				ctxPath: "test",
				props: map[string]interface{}{
					"additional_property": "additional property",
				},
			},
		},
		{
			name: "negative case: validate property error",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				ctxPath: "test",
				props: map[string]interface{}{
					"property": 123,
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: validate pattern property error",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				ctxPath: "test",
				props: map[string]interface{}{
					"pattern": 123,
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: property is not present",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
					AdditionalProperties: func() *bool {
						b := false
						return &b
					}(),
				},
			},
			args: args{
				ctxPath: "test",
				props: map[string]interface{}{
					"additional_property": "additional property",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: additional property error",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					AdditionalProperties: func() *bool {
						b := false
						return &b
					}(),
				},
			},
			args: args{
				ctxPath: "test",
				props: map[string]interface{}{
					"additional_property": "additional property",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.validateProperties(tt.args.ctxPath, tt.args.props); (err != nil) != tt.wantErr {
				t.Errorf("validateProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectShape_validate(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		v       interface{}
		ctxPath string
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
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MinProperties: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
					MaxProperties: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
					AdditionalProperties: func() *bool {
						b := true
						return &b
					}(),
				},
			},
			args: args{
				v: map[string]interface{}{
					"property1": "property1",
					"property2": "property2",
					"property3": "property3",
					"property4": "property3",
				},
			},
		},
		{
			name: "negative case: invalid value type",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
			args: args{
				v: 123,
			},
			wantErr: true,
		},
		{
			name: "negative case: additional properties constraint violation",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					AdditionalProperties: func() *bool {
						b := false
						return &b
					}(),
				},
			},
			args: args{
				v: map[string]interface{}{
					"property": "property",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: max properties constraint violation",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MaxProperties: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			args: args{
				v: map[string]interface{}{
					"property1": "property1",
					"property2": "property2",
					"property3": "property3",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: min properties constraint violation",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MinProperties: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			args: args{
				v: map[string]interface{}{
					"property1": "property1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.validate(tt.args.v, tt.args.ctxPath); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectShape_inheritMinProperties(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		source *ObjectShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got *ObjectShape) (string, bool)
		wantErr bool
	}{
		{
			name: "positive case with min properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MinProperties: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MinProperties: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.MinProperties == nil {
					return "MinProperties hasn't been inherited", false
				}
				if *got.MinProperties != 2 {
					return "MinProperties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "positive case without min properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MinProperties: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.MinProperties == nil {
					return "MinProperties hasn't been inherited", false
				}
				if *got.MinProperties != 4 {
					return "MinProperties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "negative case: min properties inheritance error",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MinProperties: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MinProperties: func() *uint64 {
							i := uint64(2)
							return &i
						}(),
					},
				},
			},
			wantErr: true,
			want: func(got *ObjectShape) (string, bool) {
				if got.MinProperties == nil {
					return "MinProperties hasn't been inherited", false
				}
				if *got.MinProperties != 4 {
					return "MinProperties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.inheritMinProperties(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("inheritMinProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				if msg, ok := tt.want(s); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestObjectShape_inheritMaxProperties(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		source *ObjectShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got *ObjectShape) (string, bool)
		wantErr bool
	}{
		{
			name: "positive case with max properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MaxProperties: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MaxProperties: func() *uint64 {
							i := uint64(2)
							return &i
						}(),
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.MaxProperties == nil {
					return "MaxProperties hasn't been inherited", false
				}
				if *got.MaxProperties != 4 {
					return "MaxProperties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "positive case without max properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MaxProperties: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.MaxProperties == nil {
					return "MaxProperties hasn't been inherited", false
				}
				if *got.MaxProperties != 4 {
					return "MaxProperties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "negative case: max properties inheritance error",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MaxProperties: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MaxProperties: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			wantErr: true,
			want: func(got *ObjectShape) (string, bool) {
				if got.MaxProperties == nil {
					return "MaxProperties hasn't been inherited", false
				}
				if *got.MaxProperties != 2 {
					return "MaxProperties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.inheritMaxProperties(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("inheritMaxProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				if msg, ok := tt.want(s); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestObjectShape_inheritProperties(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		source *ObjectShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    func(got *ObjectShape) (string, bool)
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("property2", Property{
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							m.Set("property", Property{
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							return m
						}(),
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.Properties == nil {
					return "Properties hasn't been inherited", false
				}
				if got.Properties.Len() != 2 {
					return "Properties hasn't been inherited correctly", false
				}
				if _, ok := got.Properties.Get("property"); !ok {
					return "Properties hasn't been inherited correctly", false
				}
				if _, ok := got.Properties.Get("property2"); !ok {
					return "Properties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "positive case: nil properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("property2", Property{
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							return m
						}(),
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.Properties == nil {
					return "Properties hasn't been inherited", false
				}
				if got.Properties.Len() != 1 {
					return "Properties hasn't been inherited correctly", false
				}
				if _, ok := got.Properties.Get("property2"); !ok {
					return "Properties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "positive case: nil source properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.Properties == nil {
					return "Properties hasn't been inherited", false
				}
				if got.Properties.Len() != 1 {
					return "Properties hasn't been inherited correctly", false
				}
				if _, ok := got.Properties.Get("property"); !ok {
					return "Properties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "negative case: cannot make required property optional",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
							Required: false,
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("property", Property{
								Base: &BaseShape{
									Shape: &StringShape{},
								},
								Required: true,
							})
							return m
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative: cannot inherit properties with different types",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{
									BaseShape: &BaseShape{
										Position: stacktrace.Position{},
									},
								},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("property", Property{
								Base: &BaseShape{
									Shape: &NumberShape{
										BaseShape: &BaseShape{
											Type: "number",
										},
									},
								},
							})
							return m
						}(),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.inheritProperties(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("inheritProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				if msg, ok := tt.want(s); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestObjectShape_inheritPatternProperties(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	type args struct {
		source *ObjectShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    func(got *ObjectShape) (string, bool)
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						m.Set("/^third_pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
							m := orderedmap.New[string, PatternProperty](0)
							m.Set("/^pattern*/", PatternProperty{
								Pattern: regexp.MustCompile("^pattern*"),
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							m.Set("/^second_pattern*/", PatternProperty{
								Pattern: regexp.MustCompile("^pattern*"),
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							return m
						}(),
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.PatternProperties == nil {
					return "PatternProperties hasn't been inherited", false
				}
				if got.PatternProperties.Len() != 3 {
					return "PatternProperties hasn't been inherited correctly", false
				}
				if _, ok := got.PatternProperties.Get("/^pattern*/"); !ok {
					return "PatternProperties hasn't been inherited correctly", false
				}
				if _, ok := got.PatternProperties.Get("/^second_pattern*/"); !ok {
					return "PatternProperties hasn't been inherited correctly", false
				}
				if _, ok := got.PatternProperties.Get("/^third_pattern*/"); !ok {
					return "PatternProperties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "positive case: nil pattern properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
							m := orderedmap.New[string, PatternProperty](0)
							m.Set("/^pattern*/", PatternProperty{
								Pattern: regexp.MustCompile("^pattern*"),
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							return m
						}(),
					},
				},
			},
			want: func(got *ObjectShape) (string, bool) {
				if got.PatternProperties == nil {
					return "PatternProperties hasn't been inherited", false
				}
				if got.PatternProperties.Len() != 1 {
					return "PatternProperties hasn't been inherited correctly", false
				}
				if _, ok := got.PatternProperties.Get("/^pattern*/"); !ok {
					return "PatternProperties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "negative: cannot inherit pattern properties with different types",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{
									BaseShape: &BaseShape{
										Position: stacktrace.Position{},
									},
								},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
							m := orderedmap.New[string, PatternProperty](0)
							m.Set("/^pattern*/", PatternProperty{
								Pattern: regexp.MustCompile("^pattern*"),
								Base: &BaseShape{
									Shape: &NumberShape{
										BaseShape: &BaseShape{
											Type: "number",
										},
									},
								},
							})
							return m
						}(),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.inheritPatternProperties(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("inheritPatternProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				if msg, ok := tt.want(s); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestObjectShape_inherit(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
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
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MinProperties: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
					MaxProperties: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MinProperties: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
						MaxProperties: func() *uint64 {
							i := uint64(2)
							return &i
						}(),
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("property2", Property{
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							m.Set("property", Property{
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							return m
						}(),
						PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
							m := orderedmap.New[string, PatternProperty](0)
							m.Set("/^pattern*/", PatternProperty{
								Pattern: regexp.MustCompile("^pattern*"),
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							m.Set("/^second_pattern*/", PatternProperty{
								Pattern: regexp.MustCompile("^pattern*"),
								Base: &BaseShape{
									Shape: &StringShape{},
								},
							})
							return m
						}(),
						AdditionalProperties: func() *bool {
							b := false
							return &b
						}(),
						Discriminator: func() *string {
							d := "discriminator2"
							return &d
						}(),
						DiscriminatorValue: func() any {
							d := "discriminator_value2"
							return &d
						}(),
					},
				},
			},
			want: func(got Shape) (string, bool) {
				s, ok := got.(*ObjectShape)
				if !ok {
					return "Shape hasn't been inherited", false
				}
				if s.MinProperties == nil {
					return "MinProperties hasn't been inherited", false
				}
				if *s.MinProperties != 2 {
					return "MinProperties hasn't been inherited correctly", false
				}
				if s.MaxProperties == nil {
					return "MaxProperties hasn't been inherited", false
				}
				if *s.MaxProperties != 4 {
					return "MaxProperties hasn't been inherited correctly", false
				}
				if s.Properties == nil {
					return "Properties hasn't been inherited", false
				}
				if s.Properties.Len() != 2 {
					return "Properties hasn't been inherited correctly", false
				}
				if s.PatternProperties == nil {
					return "PatternProperties hasn't been inherited", false
				}
				if s.PatternProperties.Len() != 2 {
					return "PatternProperties hasn't been inherited correctly", false
				}
				if s.AdditionalProperties == nil {
					return "AdditionalProperties hasn't been inherited", false
				}
				if *s.AdditionalProperties != false {
					return "AdditionalProperties hasn't been inherited correctly", false
				}
				if s.Discriminator == nil {
					return "Discriminator hasn't been inherited", false
				}
				if *s.Discriminator != "discriminator2" {
					return "Discriminator hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "positive case with recursive source",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
			args: args{
				source: &RecursiveShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					Head: &BaseShape{
						// Source
						Shape: &ObjectShape{
							BaseShape: &BaseShape{
								ID: 3,
							},
							ObjectFacets: ObjectFacets{
								Properties: func() *orderedmap.OrderedMap[string, Property] {
									m := orderedmap.New[string, Property](0)
									m.Set("property", Property{
										Base: &BaseShape{
											Shape: &StringShape{},
										},
									})
									return m
								}(),
							},
						},
					},
				},
			},
			want: func(got Shape) (string, bool) {
				if _, ok := got.(*ObjectShape); !ok {
					return "Shape hasn't been inherited", false
				}
				if _, ok := got.(*ObjectShape).Properties.Get("property"); !ok {
					return "Properties hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "negative case: cannot inherit object shape with different type",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
			args: args{
				source: &StringShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative: cannot inherit min properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MinProperties: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MinProperties: func() *uint64 {
							i := uint64(2)
							return &i
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative: cannot inherit max properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MaxProperties: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						MaxProperties: func() *uint64 {
							i := uint64(4)
							return &i
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative: cannot inherit properties with different types",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{
									BaseShape: &BaseShape{
										Position: stacktrace.Position{},
									},
								},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("property", Property{
								Base: &BaseShape{
									Shape: &NumberShape{
										BaseShape: &BaseShape{
											Type: "number",
										},
									},
								},
							})
							return m
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative: cannot inherit pattern properties with different types",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{
									BaseShape: &BaseShape{
										Position: stacktrace.Position{},
									},
								},
							},
						})
						return m
					}(),
				},
			},
			args: args{
				source: &ObjectShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					ObjectFacets: ObjectFacets{
						PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
							m := orderedmap.New[string, PatternProperty](0)
							m.Set("/^pattern*/", PatternProperty{
								Pattern: regexp.MustCompile("^pattern*"),
								Base: &BaseShape{
									Shape: &NumberShape{
										BaseShape: &BaseShape{
											Type: "number",
										},
									},
								},
							})
							return m
						}(),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			got, err := s.inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestObjectShape_checkPatternProperties(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
				},
			},
		},
		{
			name: "positive case: nil pattern properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
		},
		{
			name: "negative case: pattern properties with additional properties false",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
					AdditionalProperties: func() *bool {
						b := false
						return &b
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: invalid pattern property string shape",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{
									// minLength must be less than or equal to maxLength
									StringFacets: StringFacets{
										LengthFacets: LengthFacets{
											MinLength: func() *uint64 {
												i := uint64(10)
												return &i
											}(),
											MaxLength: func() *uint64 {
												i := uint64(5)
												return &i
											}(),
										},
									},
									BaseShape: &BaseShape{
										Position: stacktrace.Position{},
									},
								},
							},
						})
						return m
					}(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.checkPatternProperties(); (err != nil) != tt.wantErr {
				t.Errorf("checkPatternProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectShape_checkProperties(t *testing.T) {
	type fields struct {
		BaseShape    *BaseShape
		ObjectFacets ObjectFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
					Discriminator: func() *string {
						d := "property"
						return &d
					}(),
				},
			},
		},
		{
			name: "positive case: nil properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
		},
		{
			name: "negative case: invalid property string shape",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{
									// minLength must be less than or equal to maxLength
									StringFacets: StringFacets{
										LengthFacets: LengthFacets{
											MinLength: func() *uint64 {
												i := uint64(10)
												return &i
											}(),
											MaxLength: func() *uint64 {
												i := uint64(5)
												return &i
											}(),
										},
									},
									BaseShape: &BaseShape{
										Position: stacktrace.Position{},
									},
								},
							},
						})
						return m
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: discriminator property not found",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						return m
					}(),
					Discriminator: func() *string {
						d := "property"
						return &d
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: discriminator property is not a scalar",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &ObjectShape{},
							},
						})
						return m
					}(),
					Discriminator: func() *string {
						d := "property"
						return &d
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: invalid discriminator value",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{},
							},
						})
						return m
					}(),
					Discriminator: func() *string {
						d := "property"
						return &d
					}(),
					DiscriminatorValue: func() any {
						d := 1
						return &d
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				BaseShape:    tt.fields.BaseShape,
				ObjectFacets: tt.fields.ObjectFacets,
			}
			if err := s.checkProperties(); (err != nil) != tt.wantErr {
				t.Errorf("checkProperties() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectShape_check(t *testing.T) {
	type fields struct {
		NoScalarShape noScalarShape
		BaseShape     *BaseShape
		ObjectFacets  ObjectFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{},
			},
		},
		{
			name: "negative case: minProperties must be less than or equal to maxProperties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					MinProperties: func() *uint64 {
						i := uint64(4)
						return &i
					}(),
					MaxProperties: func() *uint64 {
						i := uint64(2)
						return &i
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: invalid properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Properties: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("property", Property{
							Base: &BaseShape{
								Shape: &StringShape{
									// minLength must be less than or equal to maxLength
									StringFacets: StringFacets{
										LengthFacets: LengthFacets{
											MinLength: func() *uint64 {
												i := uint64(10)
												return &i
											}(),
											MaxLength: func() *uint64 {
												i := uint64(5)
												return &i
											}(),
										},
									},
									BaseShape: &BaseShape{
										Position: stacktrace.Position{},
									},
								},
							},
						})
						return m
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: invalid pattern properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
						m := orderedmap.New[string, PatternProperty](0)
						m.Set("/^pattern*/", PatternProperty{
							Pattern: regexp.MustCompile("^pattern*"),
							Base: &BaseShape{
								Shape: &StringShape{
									// minLength must be less than or equal to maxLength
									StringFacets: StringFacets{
										LengthFacets: LengthFacets{
											MinLength: func() *uint64 {
												i := uint64(10)
												return &i
											}(),
											MaxLength: func() *uint64 {
												i := uint64(5)
												return &i
											}(),
										},
									},
									BaseShape: &BaseShape{
										Position: stacktrace.Position{},
									},
								},
							},
						})
						return m
					}(),
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: discriminator without properties",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				ObjectFacets: ObjectFacets{
					Discriminator: func() *string {
						d := "property"
						return &d
					}(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectShape{
				noScalarShape: tt.fields.NoScalarShape,
				BaseShape:     tt.fields.BaseShape,
				ObjectFacets:  tt.fields.ObjectFacets,
			}
			if err := s.check(); (err != nil) != tt.wantErr {
				t.Errorf("check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_makePatternProperty(t *testing.T) {
	type fields struct {
		fragmentsCache          map[string]Fragment
		fragmentTypes           map[string]map[string]*BaseShape
		fragmentAnnotationTypes map[string]map[string]*BaseShape
		entryPoint              Fragment
		domainExtensions        []*DomainExtension
		shapes                  []*BaseShape
		unresolvedShapes        list.List
		ctx                     context.Context
	}
	type args struct {
		nodeName            string
		propertyName        string
		v                   *yaml.Node
		location            string
		hasImplicitOptional bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got PatternProperty) (string, bool)
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
				entryPoint:              &Library{},
				domainExtensions:        []*DomainExtension{},
				shapes:                  []*BaseShape{},
				unresolvedShapes:        list.List{},
				ctx:                     context.Background(),
			},
			args: args{
				nodeName:     "pattern",
				propertyName: "/^pattern*/",
				v: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "string",
					Tag:   "!!str",
				},
				location:            "location",
				hasImplicitOptional: false,
			},
			want: func(got PatternProperty) (string, bool) {
				if !got.Pattern.MatchString("pattern") {
					return "Pattern hasn't been set correctly", false
				}
				if _, ok := got.Base.Shape.(*StringShape); !ok {
					return "Shape hasn't been set correctly", false
				}
				return "", true
			},
		},
		{
			name: "negative case: make shape error",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
				entryPoint:              &Library{},
				domainExtensions:        []*DomainExtension{},
				shapes:                  []*BaseShape{},
				unresolvedShapes:        list.List{},
				ctx:                     context.Background(),
			},
			args: args{
				nodeName:     "pattern",
				propertyName: "/^pattern*/",
				v: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "string",
					Tag:   "!!int",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: compile regexp pattern error",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
				entryPoint:              &Library{},
				domainExtensions:        []*DomainExtension{},
				shapes:                  []*BaseShape{},
				unresolvedShapes:        list.List{},
				ctx:                     context.Background(),
			},
			args: args{
				nodeName:     "pattern",
				propertyName: "/[a-z/",
				v: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "string",
					Tag:   "!!str",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RAML{
				fragmentsCache:          tt.fields.fragmentsCache,
				fragmentTypes:           tt.fields.fragmentTypes,
				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
				entryPoint:              tt.fields.entryPoint,
				domainExtensions:        tt.fields.domainExtensions,
				shapes:                  tt.fields.shapes,
				unresolvedShapes:        tt.fields.unresolvedShapes,
				ctx:                     tt.fields.ctx,
			}
			got, err := r.makePatternProperty(tt.args.nodeName, tt.args.propertyName, tt.args.v, tt.args.location, tt.args.hasImplicitOptional)
			if (err != nil) != tt.wantErr {
				t.Errorf("makePatternProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestRAML_chompImplicitOptional(t *testing.T) {
	type fields struct {
		fragmentsCache          map[string]Fragment
		fragmentTypes           map[string]map[string]*BaseShape
		fragmentAnnotationTypes map[string]map[string]*BaseShape
		entryPoint              Fragment
		domainExtensions        []*DomainExtension
		shapes                  []*BaseShape
		unresolvedShapes        list.List
		ctx                     context.Context
	}
	type args struct {
		nodeName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		{
			name: "positive case",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
				entryPoint:              &Library{},
				domainExtensions:        []*DomainExtension{},
				shapes:                  []*BaseShape{},
				unresolvedShapes:        list.List{},
				ctx:                     context.Background(),
			},
			args: args{
				nodeName: "string?",
			},
			want:  "string",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RAML{
				fragmentsCache:          tt.fields.fragmentsCache,
				fragmentTypes:           tt.fields.fragmentTypes,
				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
				entryPoint:              tt.fields.entryPoint,
				domainExtensions:        tt.fields.domainExtensions,
				shapes:                  tt.fields.shapes,
				unresolvedShapes:        tt.fields.unresolvedShapes,
				ctx:                     tt.fields.ctx,
			}
			got, got1 := r.chompImplicitOptional(tt.args.nodeName)
			if got != tt.want {
				t.Errorf("chompImplicitOptional() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("chompImplicitOptional() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestRAML_makeProperty(t *testing.T) {
	type fields struct {
		fragmentsCache          map[string]Fragment
		fragmentTypes           map[string]map[string]*BaseShape
		fragmentAnnotationTypes map[string]map[string]*BaseShape
		entryPoint              Fragment
		domainExtensions        []*DomainExtension
		shapes                  []*BaseShape
		unresolvedShapes        list.List
		ctx                     context.Context
	}
	type args struct {
		nodeName            string
		propertyName        string
		v                   *yaml.Node
		location            string
		hasImplicitOptional bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got Property) (string, bool)
		wantErr bool
	}{
		{
			name: "positive case with implicit optional name",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
				entryPoint:              &Library{},
				domainExtensions:        []*DomainExtension{},
				shapes:                  []*BaseShape{},
				unresolvedShapes:        list.List{},
				ctx:                     context.Background(),
			},
			args: args{
				nodeName:     "property?",
				propertyName: "property?",
				v: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "type",
							Tag:   "!!str",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "string",
							Tag:   "!!str",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "required",
							Tag:   "!!str",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "false",
							Tag:   "!!bool",
						},
					},
				},
				location:            "location",
				hasImplicitOptional: true,
			},
			want: func(got Property) (string, bool) {
				if got.Name != "property?" {
					return "Name hasn't been set correctly", false
				}
				if _, ok := got.Base.Shape.(*StringShape); !ok {
					return "Shape hasn't been set correctly", false
				}
				return "", true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RAML{
				fragmentsCache:          tt.fields.fragmentsCache,
				fragmentTypes:           tt.fields.fragmentTypes,
				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
				entryPoint:              tt.fields.entryPoint,
				domainExtensions:        tt.fields.domainExtensions,
				shapes:                  tt.fields.shapes,
				unresolvedShapes:        tt.fields.unresolvedShapes,
				ctx:                     tt.fields.ctx,
			}
			got, err := r.makeProperty(tt.args.nodeName, tt.args.propertyName, tt.args.v, tt.args.location, tt.args.hasImplicitOptional)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestUnionShape_clone(t *testing.T) {
	type fields struct {
		NoScalarShape noScalarShape
		BaseShape     *BaseShape
		EnumFacets    EnumFacets
		UnionFacets   UnionFacets
	}
	type args struct {
		base      *BaseShape
		clonedMap map[int64]*BaseShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(got Shape) (string, bool)
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				UnionFacets: UnionFacets{
					AnyOf: []*BaseShape{
						{
							ID:    2,
							Shape: &StringShape{},
						},
						{
							ID:    3,
							Shape: &NumberShape{},
						},
					},
				},
			},
			args: args{
				base: &BaseShape{
					ID: 4,
				},
				clonedMap: map[int64]*BaseShape{},
			},
			want: func(got Shape) (string, bool) {
				s, ok := got.(*UnionShape)
				if !ok {
					return "Shape hasn't been inherited", false
				}
				if s.AnyOf == nil {
					return "AnyOf hasn't been inherited", false
				}
				if len(s.AnyOf) != 2 {
					return "AnyOf hasn't been inherited correctly", false
				}
				return "", true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UnionShape{
				noScalarShape: tt.fields.NoScalarShape,
				BaseShape:     tt.fields.BaseShape,
				EnumFacets:    tt.fields.EnumFacets,
				UnionFacets:   tt.fields.UnionFacets,
			}
			got := s.clone(tt.args.base, tt.args.clonedMap)
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestUnionShape_validate(t *testing.T) {
	type fields struct {
		NoScalarShape noScalarShape
		BaseShape     *BaseShape
		EnumFacets    EnumFacets
		UnionFacets   UnionFacets
	}
	type args struct {
		v       interface{}
		ctxPath string
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
				BaseShape: &BaseShape{
					ID: 1,
				},
				UnionFacets: UnionFacets{
					AnyOf: []*BaseShape{
						{
							Shape: &NumberShape{},
						},
						{
							Shape: &StringShape{},
						},
					},
				},
			},
			args: args{
				v: "string",
			},
		},
		{
			name: "negative case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				UnionFacets: UnionFacets{
					AnyOf: []*BaseShape{
						{
							Shape: &NumberShape{},
						},
					},
				},
			},
			args: args{
				v: "string",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UnionShape{
				noScalarShape: tt.fields.NoScalarShape,
				BaseShape:     tt.fields.BaseShape,
				EnumFacets:    tt.fields.EnumFacets,
				UnionFacets:   tt.fields.UnionFacets,
			}
			if err := s.validate(tt.args.v, tt.args.ctxPath); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnionShape_inherit(t *testing.T) {
	type fields struct {
		noScalarShape noScalarShape
		BaseShape     *BaseShape
		EnumFacets    EnumFacets
		UnionFacets   UnionFacets
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
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				UnionFacets: UnionFacets{
					AnyOf: []*BaseShape{
						{
							ID:    2,
							Shape: &StringShape{},
						},
					},
				},
			},
			args: args{
				source: &UnionShape{
					BaseShape: &BaseShape{
						ID: 3,
					},
					UnionFacets: UnionFacets{
						AnyOf: []*BaseShape{
							{
								ID:    4,
								Shape: &StringShape{},
							},
						},
					},
				},
			},
			want: func(got Shape) (string, bool) {
				s, ok := got.(*UnionShape)
				if !ok {
					return "Shape hasn't been inherited", false
				}
				if s.AnyOf == nil {
					return "AnyOf hasn't been inherited", false
				}
				if len(s.AnyOf) != 1 {
					return "AnyOf hasn't been inherited correctly", false
				}
				return "", true
			},
		},
		{
			name: "positive case: count of AnyOf is 0",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				UnionFacets: UnionFacets{
					AnyOf: []*BaseShape{},
				},
			},
			args: args{
				source: &UnionShape{
					BaseShape: &BaseShape{},
				},
			},
		},
		{
			name: "negative case: cannot inherit with different types",
			fields: fields{
				BaseShape: &BaseShape{ID: 1},
			},
			args: args{
				source: &StringShape{
					BaseShape: &BaseShape{ID: 2},
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: failed to find compatible union member",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				UnionFacets: UnionFacets{
					AnyOf: []*BaseShape{
						{
							ID: 2,
							Shape: &StringShape{
								BaseShape: &BaseShape{},
							},
						},
					},
				},
			},
			args: args{
				source: &UnionShape{
					BaseShape: &BaseShape{
						ID: 3,
					},
					UnionFacets: UnionFacets{
						AnyOf: []*BaseShape{
							{
								ID: 4,
								Shape: &NumberShape{
									BaseShape: &BaseShape{},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UnionShape{
				noScalarShape: tt.fields.noScalarShape,
				BaseShape:     tt.fields.BaseShape,
				EnumFacets:    tt.fields.EnumFacets,
				UnionFacets:   tt.fields.UnionFacets,
			}
			got, err := s.inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}

func TestUnionShape_check(t *testing.T) {
	type fields struct {
		noScalarShape noScalarShape
		BaseShape     *BaseShape
		EnumFacets    EnumFacets
		UnionFacets   UnionFacets
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				UnionFacets: UnionFacets{
					AnyOf: []*BaseShape{
						{
							Shape: &StringShape{},
						},
					},
				},
			},
		},
		{
			name: "negative case: invalid string anyOf shape",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				UnionFacets: UnionFacets{
					AnyOf: []*BaseShape{
						{
							Shape: &StringShape{
								BaseShape: &BaseShape{
									Position: stacktrace.Position{},
								},
								// minLength must be less than or equal to maxLength
								StringFacets: StringFacets{
									LengthFacets: LengthFacets{
										MinLength: func() *uint64 {
											i := uint64(10)
											return &i
										}(),
										MaxLength: func() *uint64 {
											i := uint64(5)
											return &i
										}(),
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UnionShape{
				noScalarShape: tt.fields.noScalarShape,
				BaseShape:     tt.fields.BaseShape,
				EnumFacets:    tt.fields.EnumFacets,
				UnionFacets:   tt.fields.UnionFacets,
			}
			if err := s.check(); (err != nil) != tt.wantErr {
				t.Errorf("check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSONShape_inherit(t *testing.T) {
	type fields struct {
		noScalarShape noScalarShape
		BaseShape     *BaseShape
		Schema        *JSONSchema
		Raw           string
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
			name: "positive case",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				Schema: &JSONSchema{},
				Raw:    "{}",
			},
			args: args{
				source: &JSONShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					Schema: &JSONSchema{},
					Raw:    "{}",
				},
			},
			want: func(got Shape) (string, bool) {
				s, ok := got.(*JSONShape)
				if !ok {
					return "Shape hasn't been inherited", false
				}
				if s.Schema == nil {
					return "Schema hasn't been inherited", false
				}
				return "", true
			},
		},
		{
			name: "negative case: cannot inherit from different type",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				Schema: &JSONSchema{},
				Raw:    "{}",
			},
			args: args{
				source: &StringShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: cannot inherit from different JSONSchema",
			fields: fields{
				BaseShape: &BaseShape{
					ID: 1,
				},
				Schema: &JSONSchema{},
				Raw:    "{}",
			},
			args: args{
				source: &JSONShape{
					BaseShape: &BaseShape{
						ID: 2,
					},
					Schema: &JSONSchema{},
					Raw:    "[]",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &JSONShape{
				noScalarShape: tt.fields.noScalarShape,
				BaseShape:     tt.fields.BaseShape,
				Schema:        tt.fields.Schema,
				Raw:           tt.fields.Raw,
			}
			got, err := s.inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("Case hasn't been passed: %s", msg)
				}
			}
		})
	}
}
