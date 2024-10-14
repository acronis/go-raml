package raml

import (
	"container/list"
	"context"
	"errors"
	"reflect"
	"testing"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestRAML_EntryPoint(t *testing.T) {
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
		name   string
		fields fields
		want   Fragment
	}{
		{
			name: "positive",
			fields: fields{
				entryPoint: &Library{},
			},
			want: &Library{},
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
			if got := r.EntryPoint(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EntryPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_SetEntryPoint(t *testing.T) {
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
		entryPoint Fragment
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name: "positive",
			args: args{
				entryPoint: &Library{},
			},
			want: func(t *testing.T, r *RAML) {
				if r.EntryPoint() == nil {
					t.Errorf("entry point is nil")
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
			got := r.SetEntryPoint(tt.args.entryPoint)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_GetLocation(t *testing.T) {
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
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive",
			fields: fields{
				entryPoint: &Library{
					Location: "/tmp/location.raml",
				},
			},
			want: "/tmp/location.raml",
		},
		{
			name: "positive with nil entry point",
			fields: fields{
				entryPoint: nil,
			},
			want: "",
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
			if got := r.GetLocation(); got != tt.want {
				t.Errorf("GetLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_GetAllAnnotationsPtr(t *testing.T) {
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
		name   string
		fields fields
		want   func(t *testing.T, got []*DomainExtension)
	}{
		{
			name: "positive",
			fields: fields{
				domainExtensions: []*DomainExtension{
					{
						Location: "/tmp/location.raml",
					},
				},
			},
			want: func(t *testing.T, got []*DomainExtension) {
				if len(got) != 1 {
					t.Errorf("got %d annotations, want 1", len(got))
				}
				if got[0].Location != "/tmp/location.raml" {
					t.Errorf("got %s, want /tmp/location.raml", got[0].Location)
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
			got := r.GetAllAnnotationsPtr()
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_GetAllAnnotations(t *testing.T) {
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
		name   string
		fields fields
		want   []DomainExtension
	}{
		{
			name: "positive",
			fields: fields{
				domainExtensions: []*DomainExtension{
					{
						Location: "/tmp/location.raml",
					},
				},
			},
			want: []DomainExtension{
				{
					Location: "/tmp/location.raml",
				},
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
			if got := r.GetAllAnnotations(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want *RAML
	}{
		{
			name: "positive",
			args: args{
				ctx: context.Background(),
			},
			want: &RAML{
				fragmentTypes:           make(map[string]map[string]*BaseShape),
				fragmentAnnotationTypes: make(map[string]map[string]*BaseShape),
				fragmentsCache:          make(map[string]Fragment),
				domainExtensions:        make([]*DomainExtension, 0),
				ctx:                     context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_GetShapes(t *testing.T) {
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
		name   string
		fields fields
		want   []*BaseShape
	}{
		{
			name: "positive",
			fields: fields{
				shapes: []*BaseShape{
					{
						Location: "/tmp/location.raml",
					},
				},
			},
			want: []*BaseShape{
				{
					Location: "/tmp/location.raml",
				},
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
			if got := r.GetShapes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetShapes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_PutShape(t *testing.T) {
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
		shape *BaseShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name: "positive",
			args: args{
				shape: &BaseShape{
					Location: "/tmp/location.raml",
				},
			},
			want: func(t *testing.T, r *RAML) {
				if len(r.GetShapes()) != 1 {
					t.Errorf("got %d shapes, want 1", len(r.GetShapes()))
				}
				if r.GetShapes()[0].Location != "/tmp/location.raml" {
					t.Errorf("got %s, want /tmp/location.raml", r.GetShapes()[0].Location)
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
			r.PutShape(tt.args.shape)
			if tt.want != nil {
				tt.want(t, r)
			}
		})
	}
}

func TestRAML_GetFragmentTypePtrs(t *testing.T) {
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
		location string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]*BaseShape
	}{
		{
			name: "positive",
			fields: fields{
				fragmentTypes: map[string]map[string]*BaseShape{
					"location": {
						"key": &BaseShape{},
					},
				},
			},
			args: args{
				location: "location",
			},
			want: map[string]*BaseShape{
				"key": {},
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
			if got := r.GetFragmentTypePtrs(tt.args.location); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFragmentTypePtrs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_GetTypeFromFragmentPtr(t *testing.T) {
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
		location string
		typeName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *BaseShape
		wantErr bool
	}{
		{
			name: "positive",
			fields: fields{
				fragmentTypes: map[string]map[string]*BaseShape{
					"location": {
						"key": &BaseShape{},
					},
				},
			},
			args: args{
				location: "location",
				typeName: "key",
			},
			want: &BaseShape{},
		},
		{
			name: "negative",
			fields: fields{
				fragmentTypes: map[string]map[string]*BaseShape{
					"location": {
						"key": &BaseShape{},
					},
				},
			},
			args: args{
				location: "location2",
				typeName: "key",
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
				entryPoint:              tt.fields.entryPoint,
				domainExtensions:        tt.fields.domainExtensions,
				shapes:                  tt.fields.shapes,
				unresolvedShapes:        tt.fields.unresolvedShapes,
				ctx:                     tt.fields.ctx,
			}
			got, err := r.GetTypeFromFragmentPtr(tt.args.location, tt.args.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypeFromFragmentPtr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTypeFromFragmentPtr() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_PutTypeIntoFragment(t *testing.T) {
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
		name     string
		location string
		shape    *BaseShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name: "positive",
			fields: fields{
				fragmentTypes: map[string]map[string]*BaseShape{},
			},
			args: args{
				name:     "key",
				location: "location",
				shape:    &BaseShape{},
			},
			want: func(t *testing.T, r *RAML) {
				if len(r.GetFragmentTypePtrs("location")) != 1 {
					t.Errorf("got %d types, want 1", len(r.GetFragmentTypePtrs("location")))
				}
				if r.GetFragmentTypePtrs("location")["key"] == nil {
					t.Errorf("key is nil")
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
			r.PutTypeIntoFragment(tt.args.name, tt.args.location, tt.args.shape)
			if tt.want != nil {
				tt.want(t, r)
			}
		})
	}
}

func TestRAML_GetAnnotationTypeFromFragmentPtr(t *testing.T) {
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
		location string
		typeName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *BaseShape
		wantErr bool
	}{
		{
			name: "positive",
			fields: fields{
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{
					"location": {
						"key": &BaseShape{},
					},
				},
			},
			args: args{
				location: "location",
				typeName: "key",
			},
			want: &BaseShape{},
		},
		{
			name: "negative",
			fields: fields{
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
			},
			args: args{
				location: "location",
				typeName: "key",
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
				entryPoint:              tt.fields.entryPoint,
				domainExtensions:        tt.fields.domainExtensions,
				shapes:                  tt.fields.shapes,
				unresolvedShapes:        tt.fields.unresolvedShapes,
				ctx:                     tt.fields.ctx,
			}
			got, err := r.GetAnnotationTypeFromFragmentPtr(tt.args.location, tt.args.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAnnotationTypeFromFragmentPtr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnnotationTypeFromFragmentPtr() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_PutAnnotationTypeIntoFragment(t *testing.T) {
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
		name     string
		location string
		shape    *BaseShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name: "positive",
			fields: fields{
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
			},
			args: args{
				name:     "key",
				location: "location",
				shape:    &BaseShape{},
			},
			want: func(t *testing.T, r *RAML) {
				if len(r.GetFragmentTypePtrs("location")) != 1 {
					t.Errorf("got %d types, want 1", len(r.GetFragmentTypePtrs("location")))
				}
				if r.GetFragmentTypePtrs("location")["key"] == nil {
					t.Errorf("key is nil")
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
			r.PutAnnotationTypeIntoFragment(tt.args.name, tt.args.location, tt.args.shape)
		})
	}
}

func TestRAML_GetFragment(t *testing.T) {
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
		location string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Fragment
	}{
		{
			name: "positive",
			fields: fields{
				fragmentsCache: map[string]Fragment{
					"location": &Library{},
				},
			},
			args: args{
				location: "location",
			},
			want: &Library{},
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
			if got := r.GetFragment(tt.args.location); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFragment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_PutFragment(t *testing.T) {
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
		location string
		fragment Fragment
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name: "positive",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				location: "location",
				fragment: &Library{},
			},
			want: func(t *testing.T, r *RAML) {
				if r.GetFragment("location") == nil {
					t.Errorf("fragment is nil")
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
			r.PutFragment(tt.args.location, tt.args.fragment)
		})
	}
}

func TestRAML_GetReferencedType(t *testing.T) {
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
		refName  string
		location string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *BaseShape
		wantErr bool
	}{
		{
			name: "positive",
			fields: fields{
				fragmentsCache: map[string]Fragment{
					"location": &Library{
						Location: "/tmp/location.raml",
						Types: func() *orderedmap.OrderedMap[string, *BaseShape] {
							m := orderedmap.New[string, *BaseShape](0)
							m.Set("key", &BaseShape{})
							return m
						}(),
					},
				},
			},
			args: args{
				refName:  "key",
				location: "location",
			},
			want: &BaseShape{},
		},
		{
			name: "negative: type not found",
			fields: fields{
				fragmentsCache: map[string]Fragment{
					"location": &Library{
						Location: "/tmp/location.raml",
						Types: func() *orderedmap.OrderedMap[string, *BaseShape] {
							m := orderedmap.New[string, *BaseShape](0)
							return m
						}(),
					},
				},
			},
			args: args{
				refName:  "key",
				location: "location",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "negative: fragment not found",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				refName:  "key",
				location: "location",
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
				entryPoint:              tt.fields.entryPoint,
				domainExtensions:        tt.fields.domainExtensions,
				shapes:                  tt.fields.shapes,
				unresolvedShapes:        tt.fields.unresolvedShapes,
				ctx:                     tt.fields.ctx,
			}
			got, err := r.GetReferencedType(tt.args.refName, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReferencedType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetReferencedType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_GetReferencedAnnotationType(t *testing.T) {
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
		refName  string
		location string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *BaseShape
		wantErr bool
	}{
		{
			name: "positive",
			fields: fields{
				fragmentsCache: map[string]Fragment{
					"location": &Library{
						Location: "/tmp/location.raml",
						AnnotationTypes: func() *orderedmap.OrderedMap[string, *BaseShape] {
							m := orderedmap.New[string, *BaseShape](0)
							m.Set("key", &BaseShape{})
							return m
						}(),
					},
				},
			},
			args: args{
				refName:  "key",
				location: "location",
			},
			want: &BaseShape{},
		},
		{
			name: "negative: type not found",
			fields: fields{
				fragmentsCache: map[string]Fragment{
					"location": &Library{
						Location: "/tmp/location.raml",
						AnnotationTypes: func() *orderedmap.OrderedMap[string, *BaseShape] {
							m := orderedmap.New[string, *BaseShape](0)
							return m
						}(),
					},
				},
			},
			args: args{
				refName:  "key",
				location: "location",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "negative: fragment not found",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				refName:  "key",
				location: "location",
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
				entryPoint:              tt.fields.entryPoint,
				domainExtensions:        tt.fields.domainExtensions,
				shapes:                  tt.fields.shapes,
				unresolvedShapes:        tt.fields.unresolvedShapes,
				ctx:                     tt.fields.ctx,
			}
			got, err := r.GetReferencedAnnotationType(tt.args.refName, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReferencedAnnotationType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetReferencedAnnotationType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_SetHook(t *testing.T) {
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
		key     HookKey
		handler func(ctx context.Context, r *RAML, params ...any) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name: "positive",
			fields: fields{
				ctx: context.Background(),
			},
			args: args{
				key: "key",
				handler: func(ctx context.Context, r *RAML, params ...any) error {
					return nil
				},
			},
			want: func(t *testing.T, r *RAML) {
				if r.ctx.Value("key") == nil {
					t.Errorf("key is nil")
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
			r.AppendHook(tt.args.key, tt.args.handler)
		})
	}
}

func TestRAML_callHook(t *testing.T) {
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
		key HookKey
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, err error)
	}{
		{
			name: "positive: handler is not set",
			fields: fields{
				ctx: context.Background(),
			},
			args: args{
				key: "key",
			},
			want: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("got not nil, want nil")
				}
			},
		},
		{
			name: "negative: handler is set",
			fields: fields{
				ctx: func() context.Context {
					ctx := context.Background()
					ctx = context.WithValue(ctx, HookKey("key"), []HookFunc{
						func(ctx context.Context, r *RAML, params ...any) error {
							return errors.New("error")
						},
					})
					return ctx
				}(),
			},
			args: args{
				key: "key",
			},
			want: func(t *testing.T, err error) {
				if err == nil {
					t.Errorf("got nil, want not nil")
				}
			},
		},
		{
			name: "negative: hook is nil",
			fields: fields{
				ctx: func() context.Context {
					ctx := context.Background()
					ctx = context.WithValue(ctx, HookKey("key"), []HookFunc{nil, nil})
					return ctx
				}(),
			},
			args: args{
				key: "key",
			},
			want: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("got not nil, want nil")
				}
			},
		},
		{
			name: "negative: hooks is nil",
			fields: fields{
				ctx: func() context.Context {
					ctx := context.Background()
					ctx = context.WithValue(ctx, HookKey("key"), []HookFunc(nil))
					return ctx
				}(),
			},
			args: args{
				key: "key",
			},
			want: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("got not nil, want nil")
				}
			},
		},
		{
			name: "negative: context is nil",
			fields: fields{
				ctx: nil,
			},
			args: args{
				key: "key",
			},
			want: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("got not nil, want nil")
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
			got := r.callHooks(tt.args.key)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_setHooks(t *testing.T) {
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
		key   HookKey
		hooks []HookFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name:   "positive",
			fields: fields{},
			args: args{
				key: HookKey("key"),
				hooks: []HookFunc{
					func(ctx context.Context, r *RAML, params ...any) error {
						return nil
					},
				},
			},
			want: func(t *testing.T, r *RAML) {
				if r.ctx.Value(HookKey("key")) == nil {
					t.Errorf("key is nil")
				}
				hooks, ok := r.ctx.Value(HookKey("key")).([]HookFunc)
				if !ok {
					t.Errorf("got not ok, want ok")
				}
				if len(hooks) != 1 {
					t.Errorf("got %d, want 1", len(hooks))
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
			r.setHooks(tt.args.key, tt.args.hooks)
			if tt.want != nil {
				tt.want(t, r)
			}
		})
	}
}

func TestRAML_PrependHook(t *testing.T) {
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
		key  HookKey
		hook HookFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "positive",
			fields: fields{
				ctx: context.Background(),
			},
			args: args{
				key: "key",
				hook: func(ctx context.Context, r *RAML, params ...any) error {
					return nil
				},
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
			r.PrependHook(tt.args.key, tt.args.hook)
		})
	}
}

func TestRAML_RemoveHook(t *testing.T) {
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
		key  HookKey
		hook HookFunc
	}
	f := func(ctx context.Context, r *RAML, params ...any) error {
		return nil
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name: "positive",
			fields: fields{
				ctx: func() context.Context {
					ctx := context.Background()
					ctx = context.WithValue(ctx, HookKey("key"), []HookFunc{
						f,
					})
					return ctx
				}(),
			},
			args: args{
				key:  "key",
				hook: f,
			},
			want: func(t *testing.T, r *RAML) {
				if r.ctx.Value(HookKey("key")) == nil {
					t.Errorf("key is nil")
				}
				hooks, ok := r.ctx.Value(HookKey("key")).([]HookFunc)
				if !ok {
					t.Errorf("got not ok, want ok")
				}
				if len(hooks) != 0 {
					t.Errorf("got %d, want 0", len(hooks))
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
			r.RemoveHook(tt.args.key, tt.args.hook)
			if tt.want != nil {
				tt.want(t, r)
			}
		})
	}
}

func TestRAML_ClearHooks(t *testing.T) {
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
		key HookKey
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, r *RAML)
	}{
		{
			name: "positive",
			fields: fields{
				ctx: func() context.Context {
					ctx := context.Background()
					ctx = context.WithValue(ctx, HookKey("key"), []HookFunc{
						func(ctx context.Context, r *RAML, params ...any) error {
							return nil
						},
					})
					return ctx
				}(),
			},
			args: args{
				key: "key",
			},
			want: func(t *testing.T, r *RAML) {
				if r.ctx.Value(HookKey("key")) == nil {
					t.Errorf("key is nil")
				}
				hooks, ok := r.ctx.Value(HookKey("key")).([]HookFunc)
				if !ok {
					t.Errorf("got not ok, want ok")
				}
				if len(hooks) != 0 {
					t.Errorf("got %d, want 0", len(hooks))
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
			r.ClearHooks(tt.args.key)
			if tt.want != nil {
				tt.want(t, r)
			}
		})
	}
}
