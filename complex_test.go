package raml

import (
	"testing"

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
