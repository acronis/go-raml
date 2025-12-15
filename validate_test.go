package raml

import (
	"container/list"
	"context"
	"fmt"
	"testing"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestRAML_unwrapShape(t *testing.T) {
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
		shape       *BaseShape
		unwrapCache map[int64]*BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(r *RAML)
		want    func(t *testing.T, got *BaseShape, got1 *stacktrace.StackTrace, unwrapCache map[int64]*BaseShape)
	}{
		{
			name: "positive",
			fields: fields{
				shapes: []*BaseShape{
					{
						Shape: &ObjectShape{},
					},
				},
			},
			args: args{
				shape: &BaseShape{
					Shape: &ObjectShape{},
				},
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *BaseShape, got1 *stacktrace.StackTrace, unwrapCache map[int64]*BaseShape) {
				if got == nil {
					t.Errorf("unwrapShape() got = nil, want non-nil")
				}
				if got1 != nil {
					t.Errorf("unwrapShape() got1 = %v, want nil", got1)
				}
				if got != unwrapCache[got.ID] {
					t.Errorf("unwrapShape() got = %v, want %v", got, unwrapCache[got.ID])
				}
			},
		},
		{
			name: "negative: UnwrapShape error",
			fields: fields{
				shapes: []*BaseShape{
					{
						Shape: &MockShape{},
					},
				},
			},
			args: args{
				shape: &BaseShape{
					Type: "object",
					Shape: &MockShape{
						MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
							return nil
						},
					},
				},
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *BaseShape, got1 *stacktrace.StackTrace, unwrapCache map[int64]*BaseShape) {
				if got != nil {
					t.Errorf("unwrapShape() got = %v, want nil", got)
				}
				if got1 == nil {
					t.Errorf("unwrapShape() got1 = nil, want non-nil")
				}
			},
		},
		{
			name: "negative: find and mark recursion error",
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeFindAndMarkRecursion, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			fields: fields{
				ctx: context.Background(),
				shapes: []*BaseShape{
					{
						Shape: &ObjectShape{},
					},
				},
			},
			args: args{
				shape: &BaseShape{
					Shape: &ObjectShape{},
				},
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *BaseShape, got1 *stacktrace.StackTrace, unwrapCache map[int64]*BaseShape) {
				if got != nil {
					t.Errorf("unwrapShape() got = %v, want nil", got)
				}
				if got1 == nil {
					t.Errorf("unwrapShape() got1 = nil, want non-nil")
				}
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
			if tt.prepare != nil {
				tt.prepare(r)
			}
			got, got1 := r.unwrapShape(tt.args.shape, tt.args.unwrapCache)
			if tt.want != nil {
				tt.want(t, got, got1, tt.args.unwrapCache)
			}
		})
	}
}

func TestRAML_validateTypes(t *testing.T) {
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
		unwrapCache map[int64]*BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(r *RAML)
		want    func(t *testing.T, got *stacktrace.StackTrace)
	}{
		{
			name:   "positive",
			fields: fields{},
			args: args{
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *stacktrace.StackTrace) {
				if got != nil {
					t.Errorf("validateTypes() got = %v, want nil", got)
				}
			},
		},
		{
			name: "negative: unwrap shape error",
			fields: fields{
				shapes: []*BaseShape{
					{
						Shape: &MockShape{},
					},
				},
			},
			args: args{
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *stacktrace.StackTrace) {
				if got == nil {
					t.Errorf("validateTypes() got = nil, want non-nil")
				}
				if len(got.List) != 1 {
					t.Errorf("validateTypes() got = %v, want 1", len(got.List))
				}
			},
		},
		{
			name: "negative: shape check error",
			fields: fields{
				shapes: []*BaseShape{
					{
						Shape: &MockShape{},
					},
				},
			},
			args: args{
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *stacktrace.StackTrace) {
				if got == nil {
					t.Errorf("validateTypes() got = nil, want non-nil")
				}
			},
		},
		{
			name: "negative: validate shape commons error",
			fields: fields{
				ctx: context.Background(),
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapeCommons, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			args: args{
				unwrapCache: map[int64]*BaseShape{},
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
			if tt.prepare != nil {
				tt.prepare(r)
			}
			got := r.validateTypes(tt.args.unwrapCache)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_validateDomainExtensions(t *testing.T) {
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
		unwrapCache map[int64]*BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(r *RAML)
		want    func(t *testing.T, got *stacktrace.StackTrace)
	}{
		{
			name: "positive",
			fields: fields{
				domainExtensions: []*DomainExtension{
					{
						DefinedBy: &BaseShape{
							Shape: &MockShape{
								MockValidate: func(v interface{}, ctxPath string) error {
									return nil
								},
							},
							unwrapped: true,
						},
						Extension: &Node{},
					},
					{
						DefinedBy: &BaseShape{
							Shape: &MockShape{
								MockValidate: func(v interface{}, ctxPath string) error {
									return nil
								},
							},
							unwrapped: true,
						},
						Extension: &Node{},
					},
					{
						DefinedBy: &BaseShape{
							ID: 1,
							Shape: &MockShape{
								MockValidate: func(v interface{}, ctxPath string) error {
									return nil
								},
							},
							unwrapped: false,
						},
						Extension: &Node{},
					},
				},
			},
			args: args{
				unwrapCache: map[int64]*BaseShape{
					1: {
						ID: 1,
						Shape: &MockShape{
							MockValidate: func(v interface{}, ctxPath string) error {
								return nil
							},
						},
						unwrapped: true,
					},
				},
			},
			want: func(t *testing.T, got *stacktrace.StackTrace) {
				if got != nil {
					t.Errorf("validateDomainExtensions() got = %v, want nil", got)
				}
			},
		},
		{
			name: "negative: handle hook error",
			fields: fields{
				ctx: context.Background(),
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateDomainExtensions, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			args: args{
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *stacktrace.StackTrace) {
				if got == nil {
					t.Errorf("validateDomainExtensions() got = nil, want non-nil")
				}
			},
		},
		{
			name: "negative: unwrapped shape not found",
			fields: fields{
				domainExtensions: []*DomainExtension{
					{
						DefinedBy: &BaseShape{
							Shape: &MockShape{
								MockValidate: func(v interface{}, ctxPath string) error {
									return nil
								},
							},
							unwrapped: false,
						},
						Extension: &Node{},
					},
					{
						DefinedBy: &BaseShape{
							Shape: &MockShape{
								MockValidate: func(v interface{}, ctxPath string) error {
									return nil
								},
							},
							unwrapped: false,
						},
						Extension: &Node{},
					},
				},
			},
			args: args{
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *stacktrace.StackTrace) {
				if got == nil {
					t.Errorf("validateDomainExtensions() got = nil, want non-nil")
				}
			},
		},
		{
			name: "negative: validate error",
			fields: fields{
				domainExtensions: []*DomainExtension{
					{
						DefinedBy: &BaseShape{
							Shape: &MockShape{
								MockValidate: func(v interface{}, ctxPath string) error {
									return fmt.Errorf("error1")
								},
							},
							unwrapped: true,
						},
						Extension: &Node{},
					},
					{
						DefinedBy: &BaseShape{
							Shape: &MockShape{
								MockValidate: func(v interface{}, ctxPath string) error {
									return fmt.Errorf("error2")
								},
							},
							unwrapped: true,
						},
						Extension: &Node{},
					},
				},
			},
			args: args{
				unwrapCache: map[int64]*BaseShape{},
			},
			want: func(t *testing.T, got *stacktrace.StackTrace) {
				if got == nil {
					t.Errorf("validateDomainExtensions() got = nil, want non-nil")
				}
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
			if tt.prepare != nil {
				tt.prepare(r)
			}
			got := r.validateDomainExtensions(tt.args.unwrapCache)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_ValidateShapes(t *testing.T) {
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
	tests := []struct {
		name    string
		fields  fields
		prepare func(r *RAML)
		wantErr bool
	}{
		{
			name: "positive",
			fields: fields{
				shapes: []*BaseShape{
					{
						Shape: &ObjectShape{},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "negative: hook error",
			fields: fields{},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapes, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: validate domain extensions hook error",
			fields: fields{},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateDomainExtensions, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error domain extensions")
				})
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
			if tt.prepare != nil {
				tt.prepare(r)
			}
			if err := r.ValidateShapes(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateShapes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_validateObjectShape(t *testing.T) {
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
		s *ObjectShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(r *RAML)
		wantErr bool
	}{
		{
			name:   "positive",
			fields: fields{},
			args: args{
				s: &ObjectShape{
					ObjectFacets: ObjectFacets{
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("key", Property{
								Base: &BaseShape{},
							})
							m.Set("key2", Property{
								Base: &BaseShape{},
							})
							return m
						}(),
						PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
							m := orderedmap.New[string, PatternProperty](0)
							m.Set("key", PatternProperty{
								Base: &BaseShape{},
							})
							m.Set("key2", PatternProperty{
								Base: &BaseShape{},
							})
							return m
						}(),
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "negative: hook error",
			fields: fields{},
			args: args{
				s: &ObjectShape{},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateObjectShape, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: validate properties error",
			fields: fields{},
			args: args{
				s: &ObjectShape{
					ObjectFacets: ObjectFacets{
						Properties: func() *orderedmap.OrderedMap[string, Property] {
							m := orderedmap.New[string, Property](0)
							m.Set("key", Property{
								Base: &BaseShape{},
							})
							m.Set("key2", Property{
								Base: &BaseShape{},
							})
							return m
						}(),
					},
				},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapeCommons, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: validate pattern properties error",
			fields: fields{},
			args: args{
				s: &ObjectShape{
					ObjectFacets: ObjectFacets{
						PatternProperties: func() *orderedmap.OrderedMap[string, PatternProperty] {
							m := orderedmap.New[string, PatternProperty](0)
							m.Set("key", PatternProperty{
								Base: &BaseShape{},
							})
							m.Set("key2", PatternProperty{
								Base: &BaseShape{},
							})
							return m
						}(),
					},
				},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapeCommons, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
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
			if tt.prepare != nil {
				tt.prepare(r)
			}
			if err := r.validateObjectShape(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("validateObjectShape() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_validateShapeCommons(t *testing.T) {
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
		s *BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(r *RAML)
		wantErr bool
	}{
		{
			name:   "positive: object shape",
			fields: fields{},
			args: args{
				s: &BaseShape{
					Shape: &ObjectShape{},
				},
			},
			wantErr: false,
		},
		{
			name:   "positive: array shape",
			fields: fields{},
			args: args{
				s: &BaseShape{
					Shape: &ArrayShape{
						ArrayFacets: ArrayFacets{
							Items: &BaseShape{},
						},
					},
				},
			},
		},
		{
			name:   "positive: union shape",
			fields: fields{},
			args: args{
				s: &BaseShape{
					Shape: &UnionShape{
						UnionFacets: UnionFacets{
							AnyOf: []*BaseShape{},
						},
					},
				},
			},
		},
		{
			name:   "negative: hook error",
			fields: fields{},
			args: args{
				s: &BaseShape{},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapeCommons, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: hook validate shape facets error",
			fields: fields{},
			args: args{
				s: &BaseShape{},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapeFacets, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: hook validate examples error",
			fields: fields{},
			args: args{
				s: &BaseShape{},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateExamples, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: object shape error",
			fields: fields{},
			args: args{
				s: &BaseShape{
					Shape: &ObjectShape{},
				},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateObjectShape, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: array shape error",
			fields: fields{},
			args: args{
				s: &BaseShape{
					Shape: &ArrayShape{
						BaseShape: &BaseShape{
							ID: 2,
						},
						ArrayFacets: ArrayFacets{
							Items: &BaseShape{
								ID: 1,
								Shape: &ObjectShape{
									BaseShape: &BaseShape{
										ID: 1,
									},
								},
							},
						},
					},
				},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapeCommons, func(_ context.Context, _ *RAML, params ...any) error {
					p := params[0].(*BaseShape)
					if p.ID == 1 {
						return fmt.Errorf("error")
					}
					return nil
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: union shape error",
			fields: fields{},
			args: args{
				s: &BaseShape{
					Shape: &UnionShape{
						BaseShape: &BaseShape{
							ID: 2,
						},
						UnionFacets: UnionFacets{
							AnyOf: []*BaseShape{
								{
									ID: 1,
									Shape: &ObjectShape{
										BaseShape: &BaseShape{
											ID: 1,
										},
									},
								},
							},
						},
					},
				},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapeCommons, func(_ context.Context, _ *RAML, params ...any) error {
					p := params[0].(*BaseShape)
					if p.ID == 1 {
						return fmt.Errorf("error")
					}
					return nil
				})
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
			if tt.prepare != nil {
				tt.prepare(r)
			}
			if err := r.validateShapeCommons(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("validateShapeCommons() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_validateExamples(t *testing.T) {
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
		base *BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(r *RAML)
		wantErr bool
	}{
		{
			name:   "positive",
			fields: fields{},
			args: args{
				base: &BaseShape{
					Shape: &MockShape{
						MockValidate: func(v interface{}, ctxPath string) error {
							return nil
						},
					},
					Example: &Example{
						Data: &Node{},
					},
					Examples: &Examples{
						Map: func() *orderedmap.OrderedMap[string, *Example] {
							m := orderedmap.New[string, *Example](0)
							m.Set("key", &Example{
								Data: &Node{},
							})
							return m
						}(),
					},
					Default: &Node{},
				},
			},
			wantErr: false,
		},
		{
			name:   "negative: hook error",
			fields: fields{},
			args: args{
				base: &BaseShape{},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateExamples, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: validate example error",
			fields: fields{},
			args: args{
				base: &BaseShape{
					Shape: &MockShape{
						MockValidate: func(v interface{}, ctxPath string) error {
							return fmt.Errorf("error")
						},
					},
					Example: &Example{
						Data: &Node{},
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "negative: validate examples error",
			fields: fields{},
			args: args{
				base: &BaseShape{
					Shape: &MockShape{
						MockValidate: func(v interface{}, ctxPath string) error {
							return fmt.Errorf("error")
						},
					},
					Examples: &Examples{
						Map: func() *orderedmap.OrderedMap[string, *Example] {
							m := orderedmap.New[string, *Example](0)
							m.Set("key", &Example{
								Data: &Node{},
							})
							return m
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "negative: validate default error",
			fields: fields{},
			args: args{
				base: &BaseShape{
					Shape: &MockShape{
						MockValidate: func(v interface{}, ctxPath string) error {
							return fmt.Errorf("error")
						},
					},
					Default: &Node{},
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
			if tt.prepare != nil {
				tt.prepare(r)
			}
			if err := r.validateExamples(tt.args.base); (err != nil) != tt.wantErr {
				t.Errorf("validateExamples() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_validateShapeFacets(t *testing.T) {
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
		base *BaseShape
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(r *RAML)
		wantErr bool
	}{
		{
			name:   "positive",
			fields: fields{},
			args: args{
				base: &BaseShape{
					CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						return m
					}(),
					CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
						m := orderedmap.New[string, *Node](0)
						m.Set("key", &Node{})
						return m
					}(),
					Inherits: []*BaseShape{
						{
							Shape: &MockShape{},
							CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
								m := orderedmap.New[string, Property](0)
								m.Set("key", Property{
									Name:     "key",
									Required: true,
									Base: &BaseShape{
										Shape: &MockShape{
											MockValidate: func(v interface{}, ctxPath string) error {
												return nil
											},
										},
									},
								})
								return m
							}(),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "negative: hook error",
			fields: fields{},
			args: args{
				base: &BaseShape{},
			},
			prepare: func(r *RAML) {
				r.AppendHook(HookBeforeValidateShapeFacets, func(_ context.Context, _ *RAML, _ ...any) error {
					return fmt.Errorf("error")
				})
			},
			wantErr: true,
		},
		{
			name:   "negative: duplicate custom facet",
			fields: fields{},
			args: args{
				base: &BaseShape{
					CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						m.Set("key", Property{})
						return m
					}(),
					CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
						m := orderedmap.New[string, *Node](0)
						m.Set("key", &Node{})
						return m
					}(),
					Inherits: []*BaseShape{
						{
							Shape: &MockShape{},
							CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
								m := orderedmap.New[string, Property](0)
								m.Set("key", Property{
									Name: "key",
									Base: &BaseShape{
										Shape: &MockShape{
											MockValidate: func(v interface{}, ctxPath string) error {
												return nil
											},
										},
									},
								})
								return m
							}(),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "negative: required custom facet not found",
			fields: fields{},
			args: args{
				base: &BaseShape{
					CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						return m
					}(),
					CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
						m := orderedmap.New[string, *Node](0)
						return m
					}(),
					Inherits: []*BaseShape{
						{
							Shape: &MockShape{},
							CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
								m := orderedmap.New[string, Property](0)
								m.Set("key", Property{
									Name:     "key",
									Required: true,
									Base: &BaseShape{
										Shape: &MockShape{
											MockValidate: func(v interface{}, ctxPath string) error {
												return nil
											},
										},
									},
								})
								return m
							}(),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "negative: validate custom facet error",
			fields: fields{},
			args: args{
				base: &BaseShape{
					CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
						m := orderedmap.New[string, Property](0)
						return m
					}(),
					CustomShapeFacets: func() *orderedmap.OrderedMap[string, *Node] {
						m := orderedmap.New[string, *Node](0)
						m.Set("key", &Node{})
						return m
					}(),
					Inherits: []*BaseShape{
						{
							Shape: &MockShape{},
							CustomShapeFacetDefinitions: func() *orderedmap.OrderedMap[string, Property] {
								m := orderedmap.New[string, Property](0)
								m.Set("key", Property{
									Name:     "key",
									Required: true,
									Base: &BaseShape{
										Shape: &MockShape{
											MockValidate: func(v interface{}, ctxPath string) error {
												return nil
											},
										},
									},
								})
								return m
							}(),
						},
					},
				},
			},
			wantErr: false,
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
			if tt.prepare != nil {
				tt.prepare(r)
			}
			if err := r.validateShapeFacets(tt.args.base); (err != nil) != tt.wantErr {
				t.Errorf("validateShapeFacets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
