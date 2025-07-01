package raml

import (
	"math/big"
	"reflect"
	"regexp"
	"testing"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func Test_optOmitRefs_Apply(t *testing.T) {
	type fields struct {
		omitRefs bool
	}
	type args struct {
		e *JSONSchemaConverterOptions
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, options *JSONSchemaConverterOptions)
	}{
		{
			name: "positive case",
			fields: fields{
				omitRefs: true,
			},
			args: args{
				e: &JSONSchemaConverterOptions{},
			},
			want: func(tt *testing.T, options *JSONSchemaConverterOptions) {
				if !options.omitRefs {
					tt.Errorf("expected options.OmitRefs to be true, got false")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := optOmitRefs{
				omitRefs: tt.fields.omitRefs,
			}
			o.Apply(tt.args.e)
			if tt.want != nil {
				tt.want(t, tt.args.e)
			}
		})
	}
}

func TestWithOmitRefs(t *testing.T) {
	type args struct {
		omitRefs bool
	}
	tests := []struct {
		name string
		args args
		want func(tt *testing.T, options JSONSchemaConverterOpt)
	}{
		{
			name: "positive case",
			args: args{
				omitRefs: true,
			},
			want: func(tt *testing.T, options JSONSchemaConverterOpt) {
				opt, ok := options.(optOmitRefs)
				if !ok {
					tt.Errorf("expected options to be of type optOmitRefs, got %T", options)
				}
				if !opt.omitRefs {
					tt.Errorf("expected options.OmitRefs to be true, got false")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithOmitRefs(tt.args.omitRefs)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestNewJSONSchemaConverter(t *testing.T) {
	type args struct {
		opts []JSONSchemaConverterOpt
	}
	tests := []struct {
		name string
		args args
		want func(tt *testing.T, converter *JSONSchemaConverter)
	}{
		{
			name: "positive case",
			args: args{
				opts: []JSONSchemaConverterOpt{
					WithOmitRefs(true),
				},
			},
			want: func(tt *testing.T, converter *JSONSchemaConverter) {
				if !converter.opts.omitRefs {
					tt.Errorf("expected converter.OmitRefs to be true, got false")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewJSONSchemaConverter(tt.args.opts...)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_Convert(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s Shape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(tt *testing.T, schema *JSONSchema)
		wantErr bool
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &ObjectShape{
					BaseShape: &BaseShape{
						Type:      "object",
						unwrapped: true,
						Name:      "test",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Ref != "#/definitions/test" {
					tt.Errorf("expected schema.Ref to be #/definitions/test, got %s", schema.Ref)
				}
			},
		},
		{
			name:   "negative case: base shape must be unwrapped",
			fields: fields{},
			args: args{
				s: &ObjectShape{
					BaseShape: &BaseShape{
						Type:      "object",
						unwrapped: false,
						Name:      "test",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got, err := c.Convert(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_Visit(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s Shape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name: "positive case: visit object shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &ObjectShape{
					BaseShape: &BaseShape{
						Type: "object",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit array shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &ArrayShape{
					BaseShape: &BaseShape{
						Type: "array",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name:   "positive case: visit string shape",
			fields: fields{},
			args: args{
				s: &StringShape{
					BaseShape: &BaseShape{
						Type: "string",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name:   "positive case: visit number shape",
			fields: fields{},
			args: args{
				s: &NumberShape{
					BaseShape: &BaseShape{
						Type: "number",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name:   "positive case: visit integer shape",
			fields: fields{},
			args: args{
				s: &IntegerShape{
					BaseShape: &BaseShape{
						Type: "integer",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name:   "positive case: visit boolean shape",
			fields: fields{},
			args: args{
				s: &BooleanShape{
					BaseShape: &BaseShape{
						Type: "boolean",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name:   "positive case: visit nil shape",
			fields: fields{},
			args: args{
				s: &NilShape{
					BaseShape: &BaseShape{
						Type: "nil",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name:   "positive case: visit file shape",
			fields: fields{},
			args: args{
				s: &FileShape{
					BaseShape: &BaseShape{
						Type: "file",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit union shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &UnionShape{
					BaseShape: &BaseShape{
						Type: "union",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit any shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &AnyShape{
					BaseShape: &BaseShape{
						Type: "any",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit datetime shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &DateTimeShape{
					BaseShape: &BaseShape{
						Type: "datetime",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit date-only shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &DateOnlyShape{
					BaseShape: &BaseShape{
						Type: "date-only",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit time-only shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &TimeOnlyShape{
					BaseShape: &BaseShape{
						Type: "time-only",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit datetime-only shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &DateTimeOnlyShape{
					BaseShape: &BaseShape{
						Type: "datetime-only",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit schema shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &JSONShape{
					BaseShape: &BaseShape{
						Type: "schema",
					},
					Schema: &JSONSchema{},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit recursive shape",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
				definitions:    make(Definitions[JSONSchema]),
			},
			args: args{
				s: &RecursiveShape{
					BaseShape: &BaseShape{
						Type: "recursive",
					},
					Head: &BaseShape{
						Type: "object",
						Shape: &ObjectShape{
							BaseShape: &BaseShape{
								Type: "object",
							},
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
			},
		},
		{
			name: "positive case: visit nil",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema != nil {
					tt.Errorf("expected schema to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.Visit(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitObjectShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *ObjectShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name: "positive case",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &ObjectShape{
					BaseShape: &BaseShape{
						Type: "object",
					},
					ObjectFacets: ObjectFacets{
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("test", Property{
								Required: true,
								Name:     "test",
								Base: &BaseShape{
									Type: "string",
									Shape: &StringShape{
										BaseShape: &BaseShape{
											Type: "string",
										},
									},
								},
							})
							return m
						}(),
						PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
							m := orderedmap.New[string, PatternProperty](0)
							m.Set("/test/", PatternProperty{
								Base: &BaseShape{
									Type: "string",
									Shape: &StringShape{
										BaseShape: &BaseShape{
											Type: "string",
										},
									},
								},
							})
							return m
						}(),
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "object" {
					tt.Errorf("expected schema.Type to be object, got %s", schema.Type)
				}
				if schema.Properties == nil {
					tt.Errorf("expected schema.Properties to be non-nil, got nil")
				}
				if schema.PatternProperties == nil {
					tt.Errorf("expected schema.PatternProperties to be non-nil, got nil")
				}
				if schema.MinProperties != nil {
					tt.Errorf("expected schema.MinProperties to be nil, got non-nil")
				}
				if schema.MaxProperties != nil {
					tt.Errorf("expected schema.MaxProperties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitObjectShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitArrayShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *ArrayShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name: "positive case",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &ArrayShape{
					BaseShape: &BaseShape{
						Type: "array",
					},
					ArrayFacets: ArrayFacets{
						Items: &BaseShape{
							Shape: &StringShape{
								BaseShape: &BaseShape{
									Type: "string",
								},
							},
							Type: "string",
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "array" {
					tt.Errorf("expected schema.Type to be array, got %s", schema.Type)
				}
				if schema.Items == nil {
					tt.Errorf("expected schema.Items to be non-nil, got nil")
				}
				if schema.MinItems != nil {
					tt.Errorf("expected schema.MinItems to be nil, got non-nil")
				}
				if schema.MaxItems != nil {
					tt.Errorf("expected schema.MaxItems to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitArrayShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitUnionShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *UnionShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name: "positive case",
			fields: fields{
				complexSchemas: make(map[int64]*JSONSchema),
			},
			args: args{
				s: &UnionShape{
					BaseShape: &BaseShape{
						Type: "union",
					},
					UnionFacets: UnionFacets{
						AnyOf: []*BaseShape{
							{
								Shape: &StringShape{
									BaseShape: &BaseShape{},
								},
							},
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.AnyOf == nil {
					tt.Errorf("expected schema.AnyOf to be non-nil, got nil")
				}
				if schema.OneOf != nil {
					tt.Errorf("expected schema.OneOf to be nil, got non-nil")
				}
				if schema.AllOf != nil {
					tt.Errorf("expected schema.AllOf to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitUnionShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitStringShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *StringShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &StringShape{
					BaseShape: &BaseShape{
						Type: "string",
					},
					StringFacets: StringFacets{
						Pattern: regexp.MustCompile(".*"),
					},
					EnumFacets: EnumFacets{
						Enum: Nodes{
							{
								Value: "test",
							},
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "string" {
					tt.Errorf("expected schema.Type to be string, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Pattern != ".*" {
					tt.Errorf("expected schema.Pattern to be .*, got %s", schema.Pattern)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if !reflect.DeepEqual(schema.Enum, []interface{}{"test"}) {
					tt.Errorf("expected schema.Enum to be [test], got %v", schema.Enum)
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitStringShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitIntegerShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *IntegerShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &IntegerShape{
					BaseShape: &BaseShape{
						Type: "integer",
					},
					IntegerFacets: IntegerFacets{
						Minimum: func() *big.Int {
							i := big.NewInt(0)
							return i
						}(),
						Maximum: func() *big.Int {
							i := big.NewInt(100)
							return i
						}(),
						MultipleOf: func() *float64 {
							f := 2.0
							return &f
						}(),
					},
					EnumFacets: EnumFacets{
						Enum: Nodes{
							{
								Value: 50,
							},
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "integer" {
					tt.Errorf("expected schema.Type to be integer, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Minimum != "0" {
					tt.Errorf("expected schema.Minimum to be 0, got %s", schema.Minimum)
				}
				if schema.Maximum != "100" {
					tt.Errorf("expected schema.Maximum to be 100, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "2" {
					tt.Errorf("expected schema.MultipleOf to be 2, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if !reflect.DeepEqual(schema.Enum, []interface{}{50}) {
					tt.Errorf("expected schema.Enum to be [50], got %v", schema.Enum)
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitIntegerShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitNumberShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *NumberShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &NumberShape{
					BaseShape: &BaseShape{
						Type: "number",
					},
					NumberFacets: NumberFacets{
						Minimum: func() *float64 {
							f := 0.0
							return &f
						}(),
						Maximum: func() *float64 {
							f := 100.0
							return &f
						}(),
						MultipleOf: func() *float64 {
							f := 2.0
							return &f
						}(),
					},
					EnumFacets: EnumFacets{
						Enum: Nodes{
							{
								Value: 50.0,
							},
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "number" {
					tt.Errorf("expected schema.Type to be number, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Minimum != "0" {
					tt.Errorf("expected schema.Minimum to be 0, got %s", schema.Minimum)
				}
				if schema.Maximum != "100" {
					tt.Errorf("expected schema.Maximum to be 100, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "2" {
					tt.Errorf("expected schema.MultipleOf to be 2, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if !reflect.DeepEqual(schema.Enum, []interface{}{50.0}) {
					tt.Errorf("expected schema.Enum to be [50.0], got %v", schema.Enum)
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitNumberShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitFileShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *FileShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &FileShape{
					BaseShape: &BaseShape{
						Type: "file",
					},
					FileFacets: FileFacets{
						FileTypes: Nodes{
							{
								Value: "application/json",
							},
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "string" {
					tt.Errorf("expected schema.Type to be string, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.ContentMediaType != "application/json" {
					tt.Errorf("expected schema.ContentMediaType to be application/json, got %s", schema.ContentMediaType)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitFileShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitBooleanShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *BooleanShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &BooleanShape{
					BaseShape: &BaseShape{
						Type: "boolean",
					},
					EnumFacets: EnumFacets{
						Enum: Nodes{
							{
								Value: true,
							},
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "boolean" {
					tt.Errorf("expected schema.Type to be boolean, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Enum == nil {
					tt.Errorf("expected schema.Enum to be non-nil, got nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitBooleanShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitDateTimeShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *DateTimeShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case: rfc3339 format",
			fields: fields{},
			args: args{
				s: &DateTimeShape{
					BaseShape: &BaseShape{
						Type: "datetime",
					},
					FormatFacets: FormatFacets{
						Format: func() *string {
							s := "rfc3339"
							return &s
						}(),
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "string" {
					tt.Errorf("expected schema.Type to be string, got %s", schema.Type)
				}
				if schema.Format != "date-time" {
					tt.Errorf("expected schema.Format to be date-time, got %s", schema.Format)
				}
				if schema.Pattern != "" {
					tt.Errorf("expected schema.Pattern to be empty, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
		{
			name:   "positive case: rfc2616 format",
			fields: fields{},
			args: args{
				s: &DateTimeShape{
					BaseShape: &BaseShape{
						Type: "datetime",
					},
					FormatFacets: FormatFacets{
						Format: func() *string {
							s := "rfc2616"
							return &s
						}(),
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "string" {
					tt.Errorf("expected schema.Type to be string, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Pattern != "^(Mon|Tue|Wed|Thu|Fri|Sat|Sun), ([0-3][0-9]) (Jan|Feb|Mar|Apr|May"+
					"|Jun|Jul|Aug|Sep|Oct|Nov|Dec) ([0-9]{4}) ([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9] GMT$" {
					tt.Errorf("expected schema.Pattern to be set, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
		{
			name:   "positive case: nil format",
			fields: fields{},
			args: args{
				s: &DateTimeShape{
					BaseShape: &BaseShape{
						Type: "datetime",
					},
					FormatFacets: FormatFacets{},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "string" {
					tt.Errorf("expected schema.Type to be string, got %s", schema.Type)
				}
				if schema.Format != "date-time" {
					tt.Errorf("expected schema.Format to be date-time, got %s", schema.Format)
				}
				if schema.Pattern != "" {
					tt.Errorf("expected schema.Pattern to be empty, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitDateTimeShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitDateTimeOnlyShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *DateTimeOnlyShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &DateTimeOnlyShape{
					BaseShape: &BaseShape{
						Type: "datetime-only",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "string" {
					tt.Errorf("expected schema.Type to be string, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Pattern != "^[0-9]{4}-(?:0[0-9]|1[0-2])-(?:[0-2][0-9]|3[01])T(?:[01][0-9]|2[0-3]):[0-5]"+
					"[0-9]:[0-5][0-9]$" {
					tt.Errorf("expected schema.Pattern to be ^[0-9]{4}-(?:0[0-9]|1[0-2])-(?:[0-2][0-9]|3[01])"+
						"T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitDateTimeOnlyShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitDateOnlyShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *DateOnlyShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &DateOnlyShape{
					BaseShape: &BaseShape{
						Type: "date-only",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "string" {
					tt.Errorf("expected schema.Type to be string, got %s", schema.Type)
				}
				if schema.Format != "date" {
					tt.Errorf("expected schema.Format to be date, got %s", schema.Format)
				}
				if schema.Pattern != "" {
					tt.Errorf("expected schema.Pattern to be empty, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitDateOnlyShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitTimeOnlyShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *TimeOnlyShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &TimeOnlyShape{
					BaseShape: &BaseShape{
						Type: "time-only",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "string" {
					tt.Errorf("expected schema.Type to be string, got %s", schema.Type)
				}
				if schema.Format != "time" {
					tt.Errorf("expected schema.Format to be time, got %s", schema.Format)
				}
				if schema.Pattern != "" {
					tt.Errorf("expected schema.Pattern to be empty, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitTimeOnlyShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitNilShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *NilShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				s: &NilShape{
					BaseShape: &BaseShape{
						Type: "nil",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "null" {
					tt.Errorf("expected schema.Type to be null, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Pattern != "" {
					tt.Errorf("expected schema.Pattern to be empty, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitNilShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitRecursiveShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *RecursiveShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name: "positive case",
			fields: fields{
				definitions: Definitions[JSONSchema]{},
			},
			args: args{
				s: &RecursiveShape{
					BaseShape: &BaseShape{
						Type: "recursive",
					},
					Head: &BaseShape{
						Type: "string",
						Shape: &StringShape{
							BaseShape: &BaseShape{
								Type: "string",
							},
						},
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "" {
					tt.Errorf("expected schema.Type to be empty, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Pattern != "" {
					tt.Errorf("expected schema.Pattern to be empty, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "#/definitions/" {
					tt.Errorf("expected schema.Ref to be #/definitions/, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitRecursiveShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_VisitJSONShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		s *JSONShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name: "positive case",
			fields: fields{
				definitions: Definitions[JSONSchema]{},
			},
			args: args{
				s: &JSONShape{
					BaseShape: &BaseShape{
						Type: "json",
					},
					Schema: &JSONSchema{
						JSONSchemaGeneric: JSONSchemaGeneric[JSONSchema]{
							Type: "object",
						},
					},
					Raw: `{}`,
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "object" {
					tt.Errorf("expected schema.Type to be object, got %s", schema.Type)
				}
				if schema.Format != "" {
					tt.Errorf("expected schema.Format to be empty, got %s", schema.Format)
				}
				if schema.Pattern != "" {
					tt.Errorf("expected schema.Pattern to be empty, got %s", schema.Pattern)
				}
				if schema.Minimum != "" {
					tt.Errorf("expected schema.Minimum to be empty, got %s", schema.Minimum)
				}
				if schema.Maximum != "" {
					tt.Errorf("expected schema.Maximum to be empty, got %s", schema.Maximum)
				}
				if schema.MultipleOf != "" {
					tt.Errorf("expected schema.MultipleOf to be empty, got %s", schema.MultipleOf)
				}
				if schema.MinLength != nil {
					tt.Errorf("expected schema.MinLength to be nil, got non-nil")
				}
				if schema.MaxLength != nil {
					tt.Errorf("expected schema.MaxLength to be nil, got non-nil")
				}
				if schema.Enum != nil {
					tt.Errorf("expected schema.Enum to be nil, got non-nil")
				}
				if schema.Default != nil {
					tt.Errorf("expected schema.Default to be nil, got non-nil")
				}
				if schema.Examples != nil {
					tt.Errorf("expected schema.Examples to be nil, got non-nil")
				}
				if schema.Description != "" {
					tt.Errorf("expected schema.Description to be empty, got %s", schema.Description)
				}
				if schema.Title != "" {
					tt.Errorf("expected schema.Title to be empty, got %s", schema.Title)
				}
				if schema.Ref != "" {
					tt.Errorf("expected schema.Ref to be empty, got %s", schema.Ref)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.VisitJSONShape(tt.args.s)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_overrideCommonProperties(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		parent *JSONSchema
		child  *JSONSchema
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				parent: &JSONSchema{
					JSONSchemaGeneric: JSONSchemaGeneric[JSONSchema]{
						Type:        "object",
						Title:       "parent",
						Description: "parent description",
						Default:     "parent default",
						Examples: []interface{}{
							"parent",
						},
					},

					Annotations: func() *orderedmap.OrderedMap[string, any] {
						m := orderedmap.New[string, any](0)
						m.Set("parent", "parent")
						return m
					}(),
				},
				child: &JSONSchema{
					JSONSchemaGeneric: JSONSchemaGeneric[JSONSchema]{
						Type:        "object",
						Title:       "child",
						Description: "child description",
						Default:     "child default",
						Examples: []interface{}{
							"child",
						},
					},
					Annotations: func() *orderedmap.OrderedMap[string, any] {
						m := orderedmap.New[string, any](0)
						m.Set("child", "child")
						return m
					}(),
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "object" {
					tt.Errorf("expected schema.Type to be object, got %s", schema.Type)
				}
				if schema.Title != "parent" {
					tt.Errorf("expected schema.Title to be parent, got %s", schema.Title)
				}
				if schema.Description != "parent description" {
					tt.Errorf("expected schema.Description to be child description, got %s", schema.Description)
				}
				if schema.Default != "parent default" {
					tt.Errorf("expected schema.Default to be parent default, got %s", schema.Default)
				}
				if !reflect.DeepEqual(schema.Examples, []interface{}{"parent"}) {
					tt.Errorf("expected schema.Examples to be [parent], got %v", schema.Examples)
				}
				if schema.Annotations == nil {
					tt.Errorf("expected schema.Annotations to be non-nil, got nil")
				}
				if v, ok := schema.Annotations.Get("parent"); !ok {
					tt.Errorf("expected schema.Annotations to have parent key, got %v", v)
				}
				if v, ok := schema.Annotations.Get("child"); !ok {
					tt.Errorf("expected schema.Annotations to have child key, got %v", v)
				}
			},
		},
		{
			name:   "positive case: annotations nil",
			fields: fields{},
			args: args{
				parent: &JSONSchema{
					JSONSchemaGeneric: JSONSchemaGeneric[JSONSchema]{
						Type: "object",
					},
					Annotations: func() *orderedmap.OrderedMap[string, any] {
						m := orderedmap.New[string, any](0)
						m.Set("parent", "parent")
						return m
					}(),
				},
				child: &JSONSchema{
					JSONSchemaGeneric: JSONSchemaGeneric[JSONSchema]{
						Type: "object",
					},
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema.Annotations == nil {
					tt.Errorf("expected schema.Annotations to be non-nil, got nil")
				}
				if v, ok := schema.Annotations.Get("parent"); !ok {
					tt.Errorf("expected schema.Annotations to have parent key, got %v", v)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.overrideCommonProperties(tt.args.parent, tt.args.child)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestJSONSchemaConverter_makeSchemaFromBaseShape(t *testing.T) {
	type fields struct {
		ShapeVisitor   ShapeVisitor[JSONSchema]
		definitions    Definitions[JSONSchema]
		complexSchemas map[int64]*JSONSchema
		opts           JSONSchemaConverterOptions
	}
	type args struct {
		base *BaseShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(tt *testing.T, schema *JSONSchema)
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				base: &BaseShape{
					Type: "string",
					DisplayName: func() *string {
						v := "display name"
						return &v
					}(),
					Description: func() *string {
						v := "description"
						return &v
					}(),
					Default: &Node{
						Value: "default",
					},
					Examples: &Examples{
						Map: func() *orderedmap.OrderedMap[string, *Example] {
							m := orderedmap.New[string, *Example](0)
							m.Set("examples", &Example{
								Name: "examples",
								Data: &Node{
									Value: "value",
								},
							})
							return m
						}(),
					},
					Example: &Example{
						Name: "example",
						Data: &Node{
							Value: "value",
						},
					},
					CustomDomainProperties: func() *orderedmap.OrderedMap[string, *DomainExtension] {
						m := orderedmap.New[string, *DomainExtension](0)
						m.Set("custom", &DomainExtension{
							Name: "custom",
							Extension: &Node{
								Value: "value",
							},
						})
						return m
					}(),
					CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						prop := Property{
							Name: "custom",
							Base: &BaseShape{
								Type: "string",
							},
						}
						m.Set("custom", prop)
						return m
					}(),
					CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
						m := orderedmap.New[string, *Node](0)
						m.Set("custom", &Node{
							Value: "value",
						})
						return m
					}(),
				},
			},
			want: func(tt *testing.T, schema *JSONSchema) {
				if schema == nil {
					tt.Errorf("expected schema to be non-nil, got nil")
				}
				if schema.Type != "" {
					tt.Errorf("expected schema.Type to be empty, got %s", schema.Type)
				}
				if schema.Title != "display name" {
					tt.Errorf("expected schema.Title to be display name, got %s", schema.Title)
				}
				if schema.Description != "description" {
					tt.Errorf("expected schema.Description to be description, got %s", schema.Description)
				}
				if schema.Default != "default" {
					tt.Errorf("expected schema.Default to be default, got %s", schema.Default)
				}
				if schema.Examples == nil {
					tt.Errorf("expected schema.Examples to be non-nil, got nil")
				}
				if len(schema.Examples) != 1 {
					tt.Errorf("expected schema.Examples to have 1 item, got %d", len(schema.Examples))
				}
				if schema.Examples[0] != "value" {
					tt.Errorf("expected schema.Examples to have value, got %v", schema.Examples)
				}
				if schema.Annotations == nil {
					tt.Errorf("expected schema.Annotations to be non-nil, got nil")
				}
				if v, ok := schema.Annotations.Get("custom"); !ok && v != "value" {
					tt.Errorf("expected schema.Annotations to have custom key, got %v", v)
				}
				if schema.FacetData == nil {
					tt.Errorf("expected schema.FacetData to be non-nil, got nil")
				}
				if v, ok := schema.FacetData.Get("custom"); !ok && v != "value" {
					tt.Errorf("expected schema.FacetData to have custom key, got %v", v)
				}
				if schema.FacetDefinitions == nil {
					tt.Errorf("expected schema.FacetDefinitions to be non-nil, got nil")
				}
				if v, ok := schema.FacetDefinitions.Get("custom"); !ok {
					tt.Errorf("expected schema.FacetDefinitions to have custom key in definitions, got %v", v)
				}
				if schema.Properties != nil {
					tt.Errorf("expected schema.Properties to be nil, got non-nil")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JSONSchemaConverter{
				ShapeVisitor:   tt.fields.ShapeVisitor,
				definitions:    tt.fields.definitions,
				complexSchemas: tt.fields.complexSchemas,
				opts:           tt.fields.opts,
			}
			got := c.makeSchemaFromBaseShape(tt.args.base)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}
