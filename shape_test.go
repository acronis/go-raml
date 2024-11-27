package raml

// import (
// 	"container/list"
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/acronis/go-stacktrace"
// 	orderedmap "github.com/wk8/go-ordered-map/v2"
// 	"gopkg.in/yaml.v3"
// )

// func TestBaseShape_SetShape(t *testing.T) {
// 	type fields struct {
// 		Shape Shape
// 	}
// 	type args struct {
// 		shape Shape
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		{
// 			name:   "positive",
// 			fields: fields{},
// 			args: args{
// 				shape: &MockShape{},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape: tt.fields.Shape,
// 			}
// 			s.SetShape(tt.args.shape)
// 		})
// 	}
// }

// func TestBaseShape_Validate(t *testing.T) {
// 	type fields struct {
// 		Shape Shape
// 	}
// 	type args struct {
// 		v interface{}
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				Shape: &MockShape{
// 					MockValidate: func(v interface{}, ctxPath string) error {
// 						return nil
// 					},
// 				},
// 			},
// 			args: args{
// 				v: nil,
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape: tt.fields.Shape,
// 			}
// 			if err := s.Validate(tt.args.v); (err != nil) != tt.wantErr {
// 				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_Inherit(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		sourceBase *BaseShape
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		prepare func(s *BaseShape)
// 		want    func(t *testing.T, got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
// 					m := orderedmap.New[string, *Node]()
// 					return m
// 				}(),
// 				CustomDomainProperties: func() *orderedmap.OrderedMap[string, *DomainExtension] {
// 					m := orderedmap.New[string, *DomainExtension]()
// 					return m
// 				}(),
// 				Shape: &MockShape{
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return &MockShape{}, nil
// 					},
// 				},
// 			},
// 			args: args{
// 				sourceBase: &BaseShape{
// 					CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
// 						m := orderedmap.New[string, *Node]()
// 						m.Set("key", &Node{})
// 						return m
// 					}(),
// 					CustomDomainProperties: func() *orderedmap.OrderedMap[string, *DomainExtension] {
// 						m := orderedmap.New[string, *DomainExtension]()
// 						m.Set("key", &DomainExtension{})
// 						return m
// 					}(),
// 				},
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: any shape",
// 			fields: fields{
// 				Shape: &MockShape{
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return &MockShape{}, nil
// 					},
// 				},
// 			},
// 			args: args{
// 				sourceBase: &BaseShape{
// 					Shape: &AnyShape{},
// 				},
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: source is union",
// 			fields: fields{
// 				raml: New(context.Background()),
// 				Type: "string",
// 				Shape: &MockShape{
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return &MockShape{}, nil
// 					},
// 					MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 						return &MockShape{
// 							MockInherit: func(source Shape) (Shape, error) {
// 								return &MockShape{}, nil
// 							},
// 						}
// 					},
// 				},
// 			},
// 			args: args{
// 				sourceBase: &BaseShape{
// 					Shape: &UnionShape{
// 						UnionFacets: UnionFacets{
// 							AnyOf: []*BaseShape{
// 								{
// 									Type: "string",
// 									Shape: &MockShape{
// 										MockInherit: func(source Shape) (Shape, error) {
// 											return &MockShape{}, nil
// 										},
// 										MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 											return &MockShape{
// 												MockInherit: func(source Shape) (Shape, error) {
// 													return &MockShape{}, nil
// 												},
// 											}
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: target is union",
// 			fields: fields{
// 				Shape: &UnionShape{
// 					BaseShape: &BaseShape{},
// 				},
// 			},
// 			args: args{
// 				sourceBase: &BaseShape{
// 					Shape: &MockShape{
// 						MockInherit: func(source Shape) (Shape, error) {
// 							return &MockShape{}, nil
// 						},
// 					},
// 				},
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name:   "negative: hook error",
// 			fields: fields{},
// 			prepare: func(s *BaseShape) {
// 				s.AppendRAMLHook(HookBeforeBaseShapeInherit, func(ctx context.Context, raml *RAML, params ...any) error {
// 					return fmt.Errorf("error")
// 				})
// 			},
// 			args:    args{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative target inherit error",
// 			fields: fields{
// 				CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
// 					m := orderedmap.New[string, *Node]()
// 					return m
// 				}(),
// 				CustomDomainProperties: func() *orderedmap.OrderedMap[string, *DomainExtension] {
// 					m := orderedmap.New[string, *DomainExtension]()
// 					return m
// 				}(),
// 				Shape: &MockShape{
// 					BaseShape: &BaseShape{},
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return nil, fmt.Errorf("error")
// 					},
// 				},
// 			},
// 			args: args{
// 				sourceBase: &BaseShape{
// 					CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
// 						m := orderedmap.New[string, *Node]()
// 						m.Set("key", &Node{})
// 						return m
// 					}(),
// 					CustomDomainProperties: func() *orderedmap.OrderedMap[string, *DomainExtension] {
// 						m := orderedmap.New[string, *DomainExtension]()
// 						m.Set("key", &DomainExtension{})
// 						return m
// 					}(),
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			if tt.prepare != nil {
// 				tt.prepare(s)
// 			}
// 			got, err := s.Inherit(tt.args.sourceBase)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Inherit() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_inheritUnionSource(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		sourceUnion *UnionShape
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		prepare func(s *BaseShape)
// 		want    func(t *testing.T, got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive: one filtered type",
// 			fields: fields{
// 				raml: New(context.Background()),
// 				Shape: &MockShape{
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return &MockShape{}, nil
// 					},
// 					MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 						return &MockShape{
// 							MockInherit: func(source Shape) (Shape, error) {
// 								return &MockShape{}, nil
// 							},
// 						}
// 					},
// 				},
// 			},
// 			args: args{
// 				sourceUnion: &UnionShape{
// 					UnionFacets: UnionFacets{
// 						AnyOf: []*BaseShape{
// 							{
// 								Shape: &MockShape{
// 									MockInherit: func(source Shape) (Shape, error) {
// 										return &MockShape{}, nil
// 									},
// 								},
// 							},
// 							{
// 								Shape: &AnyShape{},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: two filtered types",
// 			fields: fields{
// 				raml: New(context.Background()),
// 				Shape: &MockShape{
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return &MockShape{}, nil
// 					},
// 					MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 						return &MockShape{
// 							MockInherit: func(source Shape) (Shape, error) {
// 								return &MockShape{}, nil
// 							},
// 						}
// 					},
// 				},
// 			},
// 			args: args{
// 				sourceUnion: &UnionShape{
// 					UnionFacets: UnionFacets{
// 						AnyOf: []*BaseShape{
// 							{
// 								Shape: &MockShape{
// 									MockInherit: func(source Shape) (Shape, error) {
// 										return &MockShape{}, nil
// 									},
// 								},
// 							},
// 							{
// 								Shape: &MockShape{
// 									MockInherit: func(source Shape) (Shape, error) {
// 										return &MockShape{}, nil
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name:   "negative: hook error",
// 			fields: fields{},
// 			prepare: func(s *BaseShape) {
// 				s.AppendRAMLHook(HookBeforeBaseShapeInheritUnionSource, func(ctx context.Context, raml *RAML, params ...any) error {
// 					return fmt.Errorf("error")
// 				})
// 			},
// 			args:    args{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: inherit error",
// 			fields: fields{
// 				raml: New(context.Background()),
// 				Shape: &MockShape{
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return nil, fmt.Errorf("error")
// 					},
// 					MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 						return &MockShape{
// 							BaseShape: &BaseShape{},
// 							MockInherit: func(source Shape) (Shape, error) {
// 								return nil, fmt.Errorf("error")
// 							},
// 						}
// 					},
// 				},
// 			},
// 			args: args{
// 				sourceUnion: &UnionShape{
// 					UnionFacets: UnionFacets{
// 						AnyOf: []*BaseShape{
// 							{
// 								Shape: &MockShape{
// 									MockInherit: func(source Shape) (Shape, error) {
// 										return nil, fmt.Errorf("error")
// 									},
// 								},
// 							},
// 							{
// 								Shape: &MockShape{
// 									MockInherit: func(source Shape) (Shape, error) {
// 										return nil, fmt.Errorf("error")
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			if tt.prepare != nil {
// 				tt.prepare(s)
// 			}
// 			got, err := s.inheritUnionSource(tt.args.sourceUnion)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("inheritUnionSource() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_inheritUnionTarget(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		targetUnion *UnionShape
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		prepare func(s *BaseShape)
// 		want    func(t *testing.T, got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				raml: New(context.Background()),
// 				Shape: &MockShape{
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return &MockShape{}, nil
// 					},
// 				},
// 			},
// 			args: args{
// 				targetUnion: &UnionShape{
// 					BaseShape: &BaseShape{},
// 					UnionFacets: UnionFacets{
// 						AnyOf: []*BaseShape{
// 							{
// 								Shape: &MockShape{
// 									MockInherit: func(source Shape) (Shape, error) {
// 										return &MockShape{}, nil
// 									},
// 									MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 										return &MockShape{
// 											MockInherit: func(source Shape) (Shape, error) {
// 												return &MockShape{}, nil
// 											},
// 										}
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name:   "negative: hook error",
// 			fields: fields{},
// 			prepare: func(s *BaseShape) {
// 				s.AppendRAMLHook(HookBeforeBaseShapeInheritUnionTarget, func(ctx context.Context, raml *RAML, params ...any) error {
// 					return fmt.Errorf("error")
// 				})
// 			},
// 			args:    args{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: inherit error",
// 			fields: fields{
// 				raml: New(context.Background()),
// 				Shape: &MockShape{
// 					MockInherit: func(source Shape) (Shape, error) {
// 						return nil, fmt.Errorf("error")
// 					},
// 				},
// 			},
// 			args: args{
// 				targetUnion: &UnionShape{
// 					BaseShape: &BaseShape{
// 						raml: New(context.Background()),
// 					},
// 					UnionFacets: UnionFacets{
// 						AnyOf: []*BaseShape{
// 							{
// 								Shape: &MockShape{
// 									BaseShape: &BaseShape{},
// 									MockInherit: func(source Shape) (Shape, error) {
// 										return nil, fmt.Errorf("error")
// 									},
// 									MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 										return &MockShape{
// 											BaseShape: &BaseShape{},
// 											MockInherit: func(source Shape) (Shape, error) {
// 												return nil, fmt.Errorf("error")
// 											},
// 										}
// 									},
// 								},
// 							},
// 							{
// 								Shape: &MockShape{
// 									BaseShape: &BaseShape{},
// 									MockInherit: func(source Shape) (Shape, error) {
// 										return nil, fmt.Errorf("error")
// 									},
// 									MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 										return &MockShape{
// 											BaseShape: &BaseShape{},
// 											MockInherit: func(source Shape) (Shape, error) {
// 												return nil, fmt.Errorf("error")
// 											},
// 										}
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			if tt.prepare != nil {
// 				tt.prepare(s)
// 			}
// 			got, err := s.inheritUnionTarget(tt.args.targetUnion)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("inheritUnionTarget() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }
// func TestBaseShape_AliasTo(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataType
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		source *BaseShape
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(t *testing.T, source *BaseShape, got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				Shape: &MockShape{},
// 			},
// 			args: args{
// 				source: &BaseShape{
// 					ID:          2,
// 					DisplayName: func() *string { s := "display name"; return &s }(),
// 					Description: func() *string { s := "description"; return &s }(),
// 					Type:        "type",
// 					TypeLabel:   "type label",
// 					Example:     &Example{},
// 					Examples:    &Examples{},
// 					Inherits:    []*BaseShape{},
// 					Alias:       &BaseShape{},
// 					Shape:       &MockShape{},
// 					Default:     &Node{},
// 					Required:    func() *bool { b := true; return &b }(),
// 					Link:        &DataType{},
// 					CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
// 						return orderedmap.New[string, *Node]()
// 					}(),
// 					CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
// 						return orderedmap.New[string, Property]()
// 					}(),
// 					CustomDomainProperties: func() *orderedmap.OrderedMap[string, *DomainExtension] {
// 						return orderedmap.New[string, *DomainExtension]()
// 					}(),
// 				},
// 			},
// 			want: func(t *testing.T, source *BaseShape, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if got.ID != 0 {
// 					t.Errorf("ID = %v, want %v", got.ID, 0)
// 				}
// 				if *got.DisplayName != "display name" {
// 					t.Errorf("DisplayName = %v, want %v", *got.DisplayName, "display name")
// 				}
// 				if got.DisplayName != source.DisplayName {
// 					t.Errorf("DisplayName is not linked")
// 				}
// 				if *got.Description != "description" {
// 					t.Errorf("Description = %v, want %v", *got.Description, "description")
// 				}
// 				if got.Description != source.Description {
// 					t.Errorf("Description is not linked")
// 				}
// 				if got.Type != "" {
// 					t.Errorf("Type = %v, want %v", got.Type, "empty string")
// 				}
// 				if got.TypeLabel != "" {
// 					t.Errorf("TypeLabel = %v, want %v", got.TypeLabel, "emtpy string")
// 				}
// 				if got.Example == nil {
// 					t.Errorf("Example is nil")
// 				}
// 				if got.Examples == nil {
// 					t.Errorf("Examples is nil")
// 				}
// 				if got.Inherits == nil {
// 					t.Errorf("Inherits is nil")
// 				}
// 				if got.Alias == nil {
// 					t.Errorf("Alias is nil")
// 				}
// 				if got.Shape == nil {
// 					t.Errorf("Shape is nil")
// 				}
// 				if got.Default == nil {
// 					t.Errorf("Default is nil")
// 				}
// 				if got.Required == nil {
// 					t.Errorf("Required is nil")
// 				}
// 				if got.Link != nil {
// 					t.Errorf("Link should be nil")
// 				}
// 				if got.CustomShapeFacets == nil {
// 					t.Errorf("CustomShapeFacets is nil")
// 				}
// 				if got.CustomShapeFacetDefinitions == nil {
// 					t.Errorf("CustomShapeFacetDefinitions is nil")
// 				}
// 				if got.CustomDomainProperties == nil {
// 					t.Errorf("CustomDomainProperties is nil")
// 				}
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			got, err := s.AliasTo(tt.args.source)
// 			if err != nil && !tt.wantErr {
// 				t.Errorf("AliasTo() error = %v", err)
// 			}
// 			if tt.want != nil {
// 				tt.want(t, tt.args.source, got)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_Clone(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		clonedMap map[int64]*BaseShape
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   func(t *testing.T, got *BaseShape)
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				Shape: &MockShape{
// 					MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 						return &MockShape{
// 							BaseShape: &BaseShape{},
// 						}
// 					},
// 				},
// 				ID:          1,
// 				Name:        "name",
// 				DisplayName: func() *string { s := "display name"; return &s }(),
// 				Description: func() *string { s := "description"; return &s }(),
// 				Type:        "type",
// 				TypeLabel:   "type label",
// 				Example:     &Example{},
// 				Examples:    &Examples{},
// 				Inherits: []*BaseShape{
// 					{
// 						Shape: &MockShape{
// 							BaseShape: &BaseShape{},
// 							MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 								return &MockShape{
// 									BaseShape: &BaseShape{},
// 								}
// 							},
// 						},
// 					},
// 				},
// 				Alias:    &BaseShape{},
// 				Default:  &Node{},
// 				Required: func() *bool { b := true; return &b }(),
// 				Link: &DataTypeFragment{
// 					Shape: &BaseShape{
// 						Shape: &MockShape{
// 							MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 								return &MockShape{
// 									BaseShape: &BaseShape{},
// 								}
// 							},
// 						},
// 					},
// 				},
// 				CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
// 					m := orderedmap.New[string, *Node]()
// 					m.Set("key", &Node{})
// 					return m
// 				}(),
// 				CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
// 					m := orderedmap.New[string, Property]()
// 					m.Set("key", Property{
// 						Base: &BaseShape{
// 							Shape: &MockShape{
// 								MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
// 									return &MockShape{
// 										BaseShape: &BaseShape{},
// 									}
// 								},
// 							},
// 						},
// 					})
// 					return m
// 				}(),
// 				CustomDomainProperties: func() *orderedmap.OrderedMap[string, *DomainExtension] {
// 					m := orderedmap.New[string, *DomainExtension]()
// 					m.Set("key", &DomainExtension{})
// 					return m
// 				}(),
// 				unwrapped:    true,
// 				ShapeVisited: true,
// 				raml:         New(context.Background()),
// 				Location:     "location",
// 				Position:     *stacktrace.NewPosition(1, 1),
// 			},
// 			args: args{
// 				clonedMap: map[int64]*BaseShape{},
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if got.ID != 1 {
// 					t.Errorf("ID = %v, want %v", got.ID, 1)
// 				}
// 				if got.Name != "name" {
// 					t.Errorf("Name = %v, want %v", got.Name, "name")
// 				}
// 				if *got.DisplayName != "display name" {
// 					t.Errorf("DisplayName = %v, want %v", *got.DisplayName, "display name")
// 				}
// 				if *got.Description != "description" {
// 					t.Errorf("Description = %v, want %v", *got.Description, "description")
// 				}
// 				if got.Type != "type" {
// 					t.Errorf("Type = %v, want %v", got.Type, "type")
// 				}
// 				if got.TypeLabel != "type label" {
// 					t.Errorf("TypeLabel = %v, want %v", got.TypeLabel, "type label")
// 				}
// 				if got.Example == nil {
// 					t.Errorf("Example is nil")
// 				}
// 				if got.Examples == nil {
// 					t.Errorf("Examples is nil")
// 				}
// 				if got.Inherits == nil {
// 					t.Errorf("Inherits is nil")
// 				}
// 				if got.Alias == nil {
// 					t.Errorf("Alias is nil")
// 				}
// 				if got.Default == nil {
// 					t.Errorf("Default is nil")
// 				}
// 				if got.Required == nil {
// 					t.Errorf("Required is nil")
// 				}
// 				if got.Link == nil {
// 					t.Errorf("Link is nil")
// 				}
// 				if got.CustomShapeFacets == nil {
// 					t.Errorf("CustomShapeFacets is nil")
// 				}
// 				if got.CustomShapeFacetDefinitions == nil {
// 					t.Errorf("CustomShapeFacetDefinitions is nil")
// 				}
// 				if got.CustomDomainProperties == nil {
// 					t.Errorf("CustomDomainProperties is nil")
// 				}
// 				if got.unwrapped != true {
// 					t.Errorf("unwrapped = %v, want %v", got.unwrapped, true)
// 				}
// 				if got.ShapeVisited != true {
// 					t.Errorf("ShapeVisited = %v, want %v", got.ShapeVisited, true)
// 				}
// 				if got.raml == nil {
// 					t.Errorf("raml is nil")
// 				}
// 				if got.Location != "location" {
// 					t.Errorf("Location = %v, want %v", got.Location, "location")
// 				}
// 				if got.Position == (stacktrace.Position{}) {
// 					t.Errorf("Position is empty")
// 				}
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			got := s.Clone(tt.args.clonedMap)
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_Check(t *testing.T) {
// 	type fields struct {
// 		Shape Shape
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 	}{
// 		{
// 			name: "negative",
// 			fields: fields{
// 				Shape: &MockShape{
// 					MockCheck: func() error {
// 						return fmt.Errorf("error")
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape: tt.fields.Shape,
// 			}
// 			if err := s.Check(); (err != nil) != tt.wantErr {
// 				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_IsUnwrapped(t *testing.T) {
// 	type fields struct {
// 		unwrapped bool
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   bool
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				unwrapped: true,
// 			},
// 			want: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				unwrapped: tt.fields.unwrapped,
// 			}
// 			if got := s.IsUnwrapped(); got != tt.want {
// 				t.Errorf("IsUnwrapped() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_String(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   string
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				Name: "name",
// 				Type: "type",
// 				Inherits: []*BaseShape{
// 					{
// 						ID:   2,
// 						Name: "name2",
// 					},
// 				},
// 			},
// 			want: "Type: type: Name: name: Inherits: name2",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			if got := s.String(); got != tt.want {
// 				t.Errorf("String() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_identifyShapeType(t *testing.T) {
// 	type args struct {
// 		shapeFacets []*yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    string
// 		wantErr bool
// 	}{
// 		{
// 			name:    "positive: without facets",
// 			args:    args{},
// 			want:    "string",
// 			wantErr: false,
// 		},
// 		{
// 			name: "positive: with facets: string",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minLength",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "maxLength",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "pattern",
// 					},
// 					{},
// 				},
// 			},
// 			want: "string",
// 		},
// 		{
// 			name: "positive: with facets: number",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minimum",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "maximum",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "multipleOf",
// 					},
// 					{},
// 				},
// 			},
// 			want: "number",
// 		},
// 		{
// 			name: "positive: with facets: object",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minProperties",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "maxProperties",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "properties",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "additionalProperties",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "discriminator",
// 					},
// 					{},
// 				},
// 			},
// 			want: "object",
// 		},
// 		{
// 			name: "positive: with facets: array",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minItems",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "maxItems",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "uniqueItems",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "items",
// 					},
// 					{},
// 				},
// 			},
// 			want: "array",
// 		},
// 		{
// 			name: "positive: with facets: file",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "fileTypes",
// 					},
// 					{},
// 				},
// 			},
// 			want: "file",
// 		},
// 		{
// 			name: "positive: with facets: file and string: should be file",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "fileTypes",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minLength",
// 					},
// 					{},
// 				},
// 			},
// 			want: "file",
// 		},
// 		{
// 			name: "positive: with facets: string and file: should be file",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minLength",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "fileTypes",
// 					},
// 					{},
// 				},
// 			},
// 			want: "file",
// 		},
// 		{
// 			name: "negative: incompatible facets",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minItems",
// 					},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "0",
// 					},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "maxLength",
// 					},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "10",
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: incompatible facets: string with pattern and file",
// 			args: args{
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minLength",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "fileTypes",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "pattern",
// 					},
// 					{},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := identifyShapeType(tt.args.shapeFacets)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("identifyShapeType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("identifyShapeType() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeRecursiveShape(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		idCounter               int64
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		headBase *BaseShape
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   func(t *testing.T, got *BaseShape)
// 	}{
// 		{
// 			name: "positive",
// 			args: args{
// 				headBase: &BaseShape{
// 					Shape: &MockShape{},
// 				},
// 			},
// 			// TODO: to add more checks
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{
// 				fragmentsCache:          tt.fields.fragmentsCache,
// 				fragmentTypes:           tt.fields.fragmentTypes,
// 				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
// 				entryPoint:              tt.fields.entryPoint,
// 				domainExtensions:        tt.fields.domainExtensions,
// 				shapes:                  tt.fields.shapes,
// 				unresolvedShapes:        tt.fields.unresolvedShapes,
// 				idCounter:               tt.fields.idCounter,
// 				ctx:                     tt.fields.ctx,
// 			}
// 			got := r.MakeRecursiveShape(tt.args.headBase)
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeJSONShape(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		idCounter               int64
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		base      *BaseShape
// 		rawSchema string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(t *testing.T, got *JSONShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive",
// 			args: args{
// 				base:      &BaseShape{},
// 				rawSchema: `{}`,
// 			},
// 			want: func(t *testing.T, got *JSONShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if got.Raw != `{}` {
// 					t.Errorf("Raw = %v, want %v", got.Raw, `{}`)
// 				}
// 			},
// 		},
// 		{
// 			name: "negative: invalid schema",
// 			args: args{
// 				base:      &BaseShape{},
// 				rawSchema: `{`,
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{
// 				fragmentsCache:          tt.fields.fragmentsCache,
// 				fragmentTypes:           tt.fields.fragmentTypes,
// 				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
// 				entryPoint:              tt.fields.entryPoint,
// 				domainExtensions:        tt.fields.domainExtensions,
// 				shapes:                  tt.fields.shapes,
// 				unresolvedShapes:        tt.fields.unresolvedShapes,
// 				idCounter:               tt.fields.idCounter,
// 				ctx:                     tt.fields.ctx,
// 			}
// 			got, err := r.MakeJSONShape(tt.args.base, tt.args.rawSchema)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("MakeJSONShape() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeConcreteShapeYAML(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		idCounter               int64
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		base        *BaseShape
// 		shapeType   string
// 		shapeFacets []*yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(t *testing.T, got Shape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive: unknown",
// 			args: args{
// 				base: &BaseShape{},
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*UnknownShape); !ok {
// 					t.Errorf("got is not UnknownShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: any",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "any",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*AnyShape); !ok {
// 					t.Errorf("got is not AnyShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: nil",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "nil",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*NilShape); !ok {
// 					t.Errorf("got is not NilShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: object",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "object",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*ObjectShape); !ok {
// 					t.Errorf("got is not ObjectShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: array",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "array",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*ArrayShape); !ok {
// 					t.Errorf("got is not ArrayShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: string",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "string",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*StringShape); !ok {
// 					t.Errorf("got is not StringShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: integer",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "integer",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*IntegerShape); !ok {
// 					t.Errorf("got is not IntegerShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: number",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "number",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*NumberShape); !ok {
// 					t.Errorf("got is not NumberShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: datetime",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "datetime",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*DateTimeShape); !ok {
// 					t.Errorf("got is not DateTimeShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: datetime-only",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "datetime-only",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*DateTimeOnlyShape); !ok {
// 					t.Errorf("got is not DateTimeOnlyShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: date-only",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "date-only",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*DateOnlyShape); !ok {
// 					t.Errorf("got is not DateOnlyShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: time-only",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "time-only",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*TimeOnlyShape); !ok {
// 					t.Errorf("got is not TimeOnlyShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: file",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "file",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*FileShape); !ok {
// 					t.Errorf("got is not FileShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: boolean",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "boolean",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*BooleanShape); !ok {
// 					t.Errorf("got is not BooleanShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: union",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "union",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*UnionShape); !ok {
// 					t.Errorf("got is not UnionShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: json",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "json",
// 			},
// 			want: func(t *testing.T, got Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if _, ok := got.(*JSONShape); !ok {
// 					t.Errorf("got is not JSONShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "negative: unmarshal yaml nodes error",
// 			args: args{
// 				base:      &BaseShape{},
// 				shapeType: "object",
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "properties",
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{
// 				fragmentsCache:          tt.fields.fragmentsCache,
// 				fragmentTypes:           tt.fields.fragmentTypes,
// 				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
// 				entryPoint:              tt.fields.entryPoint,
// 				domainExtensions:        tt.fields.domainExtensions,
// 				shapes:                  tt.fields.shapes,
// 				unresolvedShapes:        tt.fields.unresolvedShapes,
// 				idCounter:               tt.fields.idCounter,
// 				ctx:                     tt.fields.ctx,
// 			}
// 			got, err := r.MakeConcreteShapeYAML(tt.args.base, tt.args.shapeType, tt.args.shapeFacets)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("MakeConcreteShapeYAML() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeBaseShape(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		idCounter               int64
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		name     string
// 		location string
// 		position *stacktrace.Position
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   func(t *testing.T, got *BaseShape)
// 	}{
// 		{
// 			name: "positive",
// 			args: args{
// 				name:     "name",
// 				location: "location",
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if got.Name != "name" {
// 					t.Errorf("Name = %v, want %v", got.Name, "name")
// 				}
// 				if got.Location != "location" {
// 					t.Errorf("Location = %v, want %v", got.Location, "location")
// 				}
// 				if got.Position == (stacktrace.Position{}) {
// 					t.Errorf("Position is empty")
// 				}
// 				if got.CustomShapeFacets == nil {
// 					t.Errorf("CustomShapeFacets is nil")
// 				}
// 				if got.CustomShapeFacetDefinitions == nil {
// 					t.Errorf("CustomShapeFacetDefinitions is nil")
// 				}
// 				if got.CustomDomainProperties == nil {
// 					t.Errorf("CustomDomainProperties is nil")
// 				}
// 				if got.raml == nil {
// 					t.Errorf("raml is nil")
// 				}
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{
// 				fragmentsCache:          tt.fields.fragmentsCache,
// 				fragmentTypes:           tt.fields.fragmentTypes,
// 				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
// 				entryPoint:              tt.fields.entryPoint,
// 				domainExtensions:        tt.fields.domainExtensions,
// 				shapes:                  tt.fields.shapes,
// 				unresolvedShapes:        tt.fields.unresolvedShapes,
// 				idCounter:               tt.fields.idCounter,
// 				ctx:                     tt.fields.ctx,
// 			}
// 			got := r.MakeBaseShape(tt.args.name, tt.args.location, tt.args.position)
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestRAML_generateShapeID(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		idCounter               int64
// 		ctx                     context.Context
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   int64
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				idCounter: 1,
// 			},
// 			want: 2,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{
// 				fragmentsCache:          tt.fields.fragmentsCache,
// 				fragmentTypes:           tt.fields.fragmentTypes,
// 				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
// 				entryPoint:              tt.fields.entryPoint,
// 				domainExtensions:        tt.fields.domainExtensions,
// 				shapes:                  tt.fields.shapes,
// 				unresolvedShapes:        tt.fields.unresolvedShapes,
// 				idCounter:               tt.fields.idCounter,
// 				ctx:                     tt.fields.ctx,
// 			}
// 			if got := r.generateShapeID(); got != tt.want {
// 				t.Errorf("generateShapeID() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestRAML_makeShapeType(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		idCounter               int64
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		shapeTypeNode *yaml.Node
// 		shapeFacets   []*yaml.Node
// 		name          string
// 		location      string
// 		base          *BaseShape
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		prepare func(t *testing.T, r *RAML)
// 		want    func(t *testing.T, got string, got1 Shape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive: scalar node: tag !!str: value: string",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Tag:   "!!str",
// 					Value: "string",
// 				},
// 				shapeFacets: []*yaml.Node{},
// 				name:        "name",
// 				location:    "location",
// 				base:        &BaseShape{},
// 			},
// 			want: func(t *testing.T, got string, got1 Shape) {
// 				if got != "string" {
// 					t.Errorf("got = %v, want %v", got, "string")
// 				}
// 				if got1 != nil {
// 					t.Errorf("got1 is not nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: scalar node: tag !!str: value: empty -> string",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Tag:   "!!str",
// 					Value: "",
// 				},
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minLength",
// 					},
// 					{},
// 				},
// 				name:     "name",
// 				location: "location",
// 				base:     &BaseShape{},
// 			},
// 			want: func(t *testing.T, got string, got1 Shape) {
// 				if got != "string" {
// 					t.Errorf("got = %v, want %v", got, "string")
// 				}
// 				if got1 != nil {
// 					t.Errorf("got1 is not nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: scalar node: tag !!str: value: {} -> json",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Tag:   "!!str",
// 					Value: "{}",
// 				},
// 				shapeFacets: []*yaml.Node{},
// 				name:        "name",
// 				location:    "location",
// 				base:        &BaseShape{},
// 			},
// 			want: func(t *testing.T, got string, got1 Shape) {
// 				if got != "{}" {
// 					t.Errorf("got = %v, want %v", got, "json")
// 				}
// 				if got1 == nil {
// 					t.Errorf("got1 is not nil")
// 					return
// 				}
// 				if _, ok := got1.(*JSONShape); !ok {
// 					t.Errorf("got1 is not JSONShape")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: scalar node: tag !include: value: fixtures/dtype.raml -> object",
// 			fields: fields{
// 				fragmentsCache: map[string]Fragment{
// 					"fixtures/dtype.raml": &DataTypeFragment{
// 						Shape: &BaseShape{
// 							Type: "object",
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Tag:   "!include",
// 					Value: "dtype.raml",
// 				},
// 				base:     &BaseShape{},
// 				location: "fixtures/lib.raml",
// 			},
// 			want: func(t *testing.T, got string, got1 Shape) {
// 				if got != "" {
// 					t.Errorf("got = %v, want %v", got, "empty")
// 				}
// 				if got1 != nil {
// 					t.Errorf("got1 is not nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: scalar node: tag !!null: -> string",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.ScalarNode,
// 					Tag:  "!!null",
// 				},
// 				base:     &BaseShape{},
// 				location: "fixtures/lib.raml",
// 			},
// 			want: func(t *testing.T, got string, got1 Shape) {
// 				if got != "string" {
// 					t.Errorf("got = %v, want %v", got, "string")
// 				}
// 				if got1 != nil {
// 					t.Errorf("got1 is not nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: sequence node: scalar content: string",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.SequenceNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Tag:   "!!str",
// 							Value: "string",
// 						},
// 					},
// 				},
// 				shapeFacets: []*yaml.Node{},
// 				name:        "name",
// 				location:    "location",
// 				base:        &BaseShape{},
// 			},
// 			want: func(t *testing.T, got string, got1 Shape) {
// 				if got != "composite" {
// 					t.Errorf("got = %v, want %v", got, "composite")
// 				}
// 				if got1 != nil {
// 					t.Errorf("got1 is not nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "negative: unknown kind",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.Kind(1023),
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: document node",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.DocumentNode,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: alias node",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.AliasNode,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: mapping node",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: identify shape by facets error",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.ScalarNode,
// 					Tag:  "!!str",
// 				},
// 				shapeFacets: []*yaml.Node{
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "minItems",
// 					},
// 					{},
// 					{
// 						Kind:  yaml.ScalarNode,
// 						Value: "maxLength",
// 					},
// 					{},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: scalar node: invalid json",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Tag:   "!!str",
// 					Value: "{",
// 				},
// 				shapeFacets: []*yaml.Node{},
// 				name:        "name",
// 				location:    "location",
// 				base:        &BaseShape{},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: scalar node: invalid include",
// 			fields: fields{
// 				fragmentsCache: map[string]Fragment{
// 					"fixtures/dtype.raml": &DataTypeFragment{
// 						Shape: &BaseShape{
// 							Type: "object",
// 						},
// 					},
// 				},
// 			},
// 			prepare: func(t *testing.T, r *RAML) {
// 				r.AppendHook(HookBeforeParseDataType, func(ctx context.Context, r *RAML, params ...any) error {
// 					return fmt.Errorf("error")
// 				})
// 			},
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Tag:   "!include",
// 					Value: "dtype.raml",
// 				},
// 				base:     &BaseShape{},
// 				location: "fixtures/lib.raml",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: scalar node: unknown tag",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.ScalarNode,
// 					Tag:  "!!unknown",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: sequence node: no scalar content",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.SequenceNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind: yaml.SequenceNode,
// 						},
// 					},
// 				},
// 				shapeFacets: []*yaml.Node{},
// 				name:        "name",
// 				location:    "location",
// 				base:        &BaseShape{},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: sequence node: !include error",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.SequenceNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind: yaml.ScalarNode,
// 							Tag:  "!include",
// 						},
// 					},
// 				},
// 				shapeFacets: []*yaml.Node{},
// 				name:        "name",
// 				location:    "location",
// 				base:        &BaseShape{},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative: sequence node: make new shape error",
// 			args: args{
// 				shapeTypeNode: &yaml.Node{
// 					Kind: yaml.SequenceNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Tag:   "!!str",
// 							Value: "string",
// 						},
// 					},
// 				},
// 				shapeFacets: []*yaml.Node{},
// 				name:        "name",
// 				location:    "location",
// 				base:        &BaseShape{},
// 			},
// 			prepare: func(t *testing.T, r *RAML) {
// 				r.AppendHook(HookBeforeRAMLMakeNewShapeYAML, func(ctx context.Context, r *RAML, params ...any) error {
// 					return fmt.Errorf("error")
// 				})
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{
// 				fragmentsCache:          tt.fields.fragmentsCache,
// 				fragmentTypes:           tt.fields.fragmentTypes,
// 				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
// 				entryPoint:              tt.fields.entryPoint,
// 				domainExtensions:        tt.fields.domainExtensions,
// 				shapes:                  tt.fields.shapes,
// 				unresolvedShapes:        tt.fields.unresolvedShapes,
// 				idCounter:               tt.fields.idCounter,
// 				ctx:                     tt.fields.ctx,
// 			}
// 			if tt.prepare != nil {
// 				tt.prepare(t, r)
// 			}
// 			got, got1, err := r.makeShapeType(tt.args.shapeTypeNode, tt.args.shapeFacets, tt.args.name, tt.args.location, tt.args.base)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("makeShapeType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got, got1)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeNewShape(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		idCounter               int64
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		name      string
// 		shapeType string
// 		location  string
// 		position  *stacktrace.Position
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		prepare func(t *testing.T, r *RAML)
// 		want    func(t *testing.T, got *BaseShape, got1 Shape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive",
// 			args: args{},
// 			want: func(t *testing.T, got *BaseShape, got1 Shape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if got1 == nil {
// 					t.Errorf("got1 is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "negative: make concrete shape yaml error",
// 			args: args{},
// 			prepare: func(t *testing.T, r *RAML) {
// 				r.AppendHook(HookBeforeRAMLMakeConcreteShapeYAML, func(ctx context.Context, r *RAML, params ...any) error {
// 					return fmt.Errorf("error")
// 				})
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{
// 				fragmentsCache:          tt.fields.fragmentsCache,
// 				fragmentTypes:           tt.fields.fragmentTypes,
// 				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
// 				entryPoint:              tt.fields.entryPoint,
// 				domainExtensions:        tt.fields.domainExtensions,
// 				shapes:                  tt.fields.shapes,
// 				unresolvedShapes:        tt.fields.unresolvedShapes,
// 				idCounter:               tt.fields.idCounter,
// 				ctx:                     tt.fields.ctx,
// 			}
// 			if tt.prepare != nil {
// 				tt.prepare(t, r)
// 			}
// 			got, got1, err := r.MakeNewShape(tt.args.name, tt.args.shapeType, tt.args.location, tt.args.position)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("MakeNewShape() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got, got1)
// 			}
// 		})
// 	}
// }

// func TestRAML_makeNewShapeYAML(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		idCounter               int64
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		v        *yaml.Node
// 		name     string
// 		location string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		prepare func(t *testing.T, r *RAML)
// 		want    func(t *testing.T, got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name:   "positive: shape type is detected by facets",
// 			fields: fields{},
// 			args: args{
// 				v: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "minLength",
// 							Tag:   "!!str",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "0",
// 							Tag:   "!!int",
// 						},
// 					},
// 				},
// 				name:     "name",
// 				location: "location",
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name:   "positive: shape type is defined: string",
// 			fields: fields{},
// 			args: args{
// 				v: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "type",
// 							Tag:   "!!str",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "string",
// 							Tag:   "!!str",
// 						},
// 					},
// 				},
// 				name:     "name",
// 				location: "location",
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name:   "positive: shape type is defined: json",
// 			fields: fields{},
// 			args: args{
// 				v: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "{}",
// 					Tag:   "!!str",
// 				},
// 				name:     "name",
// 				location: "location",
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name:   "positive: shape type is defined: custom type: unknown shape",
// 			fields: fields{},
// 			args: args{
// 				v: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "type",
// 							Tag:   "!!str",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "custom",
// 							Tag:   "!!str",
// 						},
// 					},
// 				},
// 				name:     "name",
// 				location: "location",
// 			},
// 			want: func(t *testing.T, got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 			},
// 		},
// 		// TODO: add negative tests
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{
// 				fragmentsCache:          tt.fields.fragmentsCache,
// 				fragmentTypes:           tt.fields.fragmentTypes,
// 				fragmentAnnotationTypes: tt.fields.fragmentAnnotationTypes,
// 				entryPoint:              tt.fields.entryPoint,
// 				domainExtensions:        tt.fields.domainExtensions,
// 				shapes:                  tt.fields.shapes,
// 				unresolvedShapes:        tt.fields.unresolvedShapes,
// 				idCounter:               tt.fields.idCounter,
// 				ctx:                     tt.fields.ctx,
// 			}
// 			if tt.prepare != nil {
// 				tt.prepare(t, r)
// 			}
// 			got, err := r.makeNewShapeYAML(tt.args.v, tt.args.name, tt.args.location)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("makeNewShapeYAML() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_decodeExamples(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		valueNode *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name:   "positive",
// 			fields: fields{},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "case1",
// 							Tag:   "!!str",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "value1",
// 							Tag:   "!!str",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "positive: included example",
// 			fields: fields{
// 				raml: &RAML{
// 					fragmentsCache: map[string]Fragment{
// 						"fixtures/example.raml": &NamedExample{},
// 					},
// 				},
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "fixtures/example.raml",
// 					Tag:   "!include",
// 				},
// 			},
// 		},
// 		// TODO: add negative tests
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			if err := s.decodeExamples(tt.args.valueNode); (err != nil) != tt.wantErr {
// 				t.Errorf("decodeExamples() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_decodeFacets(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		valueNode *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive: facets: customFacet: string",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "customFacet",
// 							Tag:   "!!str",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "string",
// 							Tag:   "!!str",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			if err := s.decodeFacets(tt.args.valueNode); (err != nil) != tt.wantErr {
// 				t.Errorf("decodeFacets() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_decodeExample(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		valueNode *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive: example: value",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 					Tag:   "!!str",
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			if err := s.decodeExample(tt.args.valueNode); (err != nil) != tt.wantErr {
// 				t.Errorf("decodeExample() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_decodeValueNode(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		node      *yaml.Node
// 		valueNode *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive: facet: type: string",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "type",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "string",
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if got1 == nil {
// 					t.Errorf("got1 is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: facet: display name: value",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "displayName",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if base.DisplayName == nil {
// 					t.Errorf("base.DisplayName is nil")
// 					return
// 				}
// 				if *base.DisplayName != "value" {
// 					t.Errorf("base.DisplayName = %v, want %v", *base.DisplayName, "value")
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: facet: description: value",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "description",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if base.Description == nil {
// 					t.Errorf("base.Description is nil")
// 					return
// 				}
// 				if *base.Description != "value" {
// 					t.Errorf("base.Description = %v, want %v", *base.Description, "value")
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: facet: required: true",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "required",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "true",
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if base.Required == nil {
// 					t.Errorf("base.Required is nil")
// 					return
// 				}
// 				if *base.Required != true {
// 					t.Errorf("base.Required = %v, want %v", *base.Required, true)
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: facet: facets: custom: string",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "facets",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "custom",
// 							Tag:   "!!str",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "string",
// 							Tag:   "!!str",
// 						},
// 					},
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if base.CustomShapeFacetDefinitions == nil {
// 					t.Errorf("base.CustomShapeFacetDefinitions is nil")
// 					return
// 				}
// 				if base.CustomShapeFacetDefinitions.Len() != 1 {
// 					t.Errorf("base.CustomShapeFacetDefinitions.Len() = %v, want %v", base.CustomShapeFacetDefinitions.Len(), 1)
// 					return
// 				}
// 				if _, ok := base.CustomShapeFacetDefinitions.Get("custom"); !ok {
// 					t.Errorf("base.CustomShapeFacets.Get('custom') not ok")
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: facet: example: value",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "example",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if base.Example == nil {
// 					t.Errorf("base.Example is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: facet: examples: example: value",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "examples",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "example",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "value",
// 						},
// 					},
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if base.Examples == nil {
// 					t.Errorf("base.Examples is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: facet: default: value",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "default",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if base.Default == nil {
// 					t.Errorf("base.Default is nil")
// 					return
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: custom domain extension",
// 			fields: fields{
// 				raml:                   New(context.Background()),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "(custom)",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if base.CustomDomainProperties == nil {
// 					t.Errorf("base.CustomDomainProperties is nil")
// 					return
// 				}
// 				if base.CustomDomainProperties.Len() != 1 {
// 					t.Errorf("base.CustomDomainProperties.Len() = %v, want %v", base.CustomDomainProperties.Len(), 1)
// 					return
// 				}
// 				if _, ok := base.CustomDomainProperties.Get("custom"); !ok {
// 					t.Errorf("base.CustomDomainProperties.Get('custom') not ok")
// 				}
// 			},
// 		},
// 		{
// 			name: "positive: custom facets",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				node: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "custom",
// 				},
// 				valueNode: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			want: func(t *testing.T, base *BaseShape, got *yaml.Node, got1 []*yaml.Node) {
// 				if got1 == nil {
// 					t.Errorf("got1 is nil")
// 					return
// 				}
// 				if len(got1) != 2 {
// 					t.Errorf("len(got1) = %v, want %v", len(got1), 1)
// 					return
// 				}
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			got, got1, err := s.decodeValueNode(tt.args.node, tt.args.valueNode)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("decodeValueNode() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, s, got, got1)
// 			}
// 		})
// 	}
// }

// func TestBaseShape_decode(t *testing.T) {
// 	type fields struct {
// 		Shape                       Shape
// 		ID                          int64
// 		Name                        string
// 		DisplayName                 *string
// 		Description                 *string
// 		Type                        string
// 		TypeLabel                   string
// 		Example                     *Example
// 		Examples                    *Examples
// 		Inherits                    []*BaseShape
// 		Alias                       *BaseShape
// 		Default                     *Node
// 		Required                    *bool
// 		Link                        *DataTypeFragment
// 		CustomShapeFacets           *orderedmap.OrderedMap[string, *Node]
// 		CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
// 		CustomDomainProperties      *orderedmap.OrderedMap[string, *DomainExtension]
// 		unwrapped                   bool
// 		ShapeVisited                bool
// 		raml                        *RAML
// 		Location                    string
// 		Position                    stacktrace.Position
// 	}
// 	type args struct {
// 		value *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(t *testing.T, got *yaml.Node, got1 []*yaml.Node)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive",
// 			fields: fields{
// 				raml: New(context.Background()),
// 			},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "type",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "string",
// 						},
// 					},
// 				},
// 			},
// 			want: func(t *testing.T, got *yaml.Node, got1 []*yaml.Node) {
// 				if got == nil {
// 					t.Errorf("got is nil")
// 					return
// 				}
// 				if got1 == nil {
// 					t.Errorf("got1 is nil")
// 					return
// 				}
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &BaseShape{
// 				Shape:                       tt.fields.Shape,
// 				ID:                          tt.fields.ID,
// 				Name:                        tt.fields.Name,
// 				DisplayName:                 tt.fields.DisplayName,
// 				Description:                 tt.fields.Description,
// 				Type:                        tt.fields.Type,
// 				TypeLabel:                   tt.fields.TypeLabel,
// 				Example:                     tt.fields.Example,
// 				Examples:                    tt.fields.Examples,
// 				Inherits:                    tt.fields.Inherits,
// 				Alias:                       tt.fields.Alias,
// 				Default:                     tt.fields.Default,
// 				Required:                    tt.fields.Required,
// 				Link:                        tt.fields.Link,
// 				CustomShapeFacets:           tt.fields.CustomShapeFacets,
// 				CustomShapeFacetDefinitions: tt.fields.CustomShapeFacetDefinitions,
// 				CustomDomainProperties:      tt.fields.CustomDomainProperties,
// 				unwrapped:                   tt.fields.unwrapped,
// 				ShapeVisited:                tt.fields.ShapeVisited,
// 				raml:                        tt.fields.raml,
// 				Location:                    tt.fields.Location,
// 				Position:                    tt.fields.Position,
// 			}
// 			got, got1, err := s.decode(tt.args.value)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(t, got, got1)
// 			}
// 		})
// 	}
// }
