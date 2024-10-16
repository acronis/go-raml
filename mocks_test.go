package raml

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMockShape_Base(t *testing.T) {
	type fields struct {
		BaseShape              *BaseShape
		MockInherit            func(source Shape) (Shape, error)
		MockCheck              func() error
		MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
		MockValidate           func(v interface{}, ctxPath string) error
		MockUnmarshalYAMLNodes func(v []*yaml.Node) error
		MockString             func() string
		MockIsScalar           func() bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *BaseShape
	}{
		{
			name: "BaseShape is nil",
			fields: fields{
				BaseShape: nil,
			},
			want: nil,
		},
		{
			name: "BaseShape is not nil",
			fields: fields{
				BaseShape: &BaseShape{},
			},
			want: &BaseShape{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := MockShape{
				BaseShape:              tt.fields.BaseShape,
				MockInherit:            tt.fields.MockInherit,
				MockCheck:              tt.fields.MockCheck,
				MockClone:              tt.fields.MockClone,
				MockValidate:           tt.fields.MockValidate,
				MockUnmarshalYAMLNodes: tt.fields.MockUnmarshalYAMLNodes,
				MockString:             tt.fields.MockString,
				MockIsScalar:           tt.fields.MockIsScalar,
			}
			if got := u.Base(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Base() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockShape_IsScalar(t *testing.T) {
	type fields struct {
		BaseShape              *BaseShape
		MockInherit            func(source Shape) (Shape, error)
		MockCheck              func() error
		MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
		MockValidate           func(v interface{}, ctxPath string) error
		MockUnmarshalYAMLNodes func(v []*yaml.Node) error
		MockString             func() string
		MockIsScalar           func() bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "IsScalar returns true",
			fields: fields{
				MockIsScalar: func() bool { return true },
			},
			want: true,
		},
		{
			name: "IsScalar returns false",
			fields: fields{
				MockIsScalar: func() bool { return false },
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := MockShape{
				BaseShape:              tt.fields.BaseShape,
				MockInherit:            tt.fields.MockInherit,
				MockCheck:              tt.fields.MockCheck,
				MockClone:              tt.fields.MockClone,
				MockValidate:           tt.fields.MockValidate,
				MockUnmarshalYAMLNodes: tt.fields.MockUnmarshalYAMLNodes,
				MockString:             tt.fields.MockString,
				MockIsScalar:           tt.fields.MockIsScalar,
			}
			if got := u.IsScalar(); got != tt.want {
				t.Errorf("IsScalar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockShape_String(t *testing.T) {
	type fields struct {
		BaseShape              *BaseShape
		MockInherit            func(source Shape) (Shape, error)
		MockCheck              func() error
		MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
		MockValidate           func(v interface{}, ctxPath string) error
		MockUnmarshalYAMLNodes func(v []*yaml.Node) error
		MockString             func() string
		MockIsScalar           func() bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "String returns non-empty string",
			fields: fields{
				MockString: func() string { return "mock string" },
			},
			want: "mock string",
		},
		{
			name: "String returns empty string",
			fields: fields{
				MockString: func() string { return "" },
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := MockShape{
				BaseShape:              tt.fields.BaseShape,
				MockInherit:            tt.fields.MockInherit,
				MockCheck:              tt.fields.MockCheck,
				MockClone:              tt.fields.MockClone,
				MockValidate:           tt.fields.MockValidate,
				MockUnmarshalYAMLNodes: tt.fields.MockUnmarshalYAMLNodes,
				MockString:             tt.fields.MockString,
				MockIsScalar:           tt.fields.MockIsScalar,
			}
			if got := u.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockShape_check(t *testing.T) {
	type fields struct {
		BaseShape              *BaseShape
		MockInherit            func(source Shape) (Shape, error)
		MockCheck              func() error
		MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
		MockValidate           func(v interface{}, ctxPath string) error
		MockUnmarshalYAMLNodes func(v []*yaml.Node) error
		MockString             func() string
		MockIsScalar           func() bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "check returns no error",
			fields: fields{
				MockCheck: func() error { return nil },
			},
			wantErr: false,
		},
		{
			name: "check returns error",
			fields: fields{
				MockCheck: func() error { return fmt.Errorf("mock error") },
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := MockShape{
				BaseShape:              tt.fields.BaseShape,
				MockInherit:            tt.fields.MockInherit,
				MockCheck:              tt.fields.MockCheck,
				MockClone:              tt.fields.MockClone,
				MockValidate:           tt.fields.MockValidate,
				MockUnmarshalYAMLNodes: tt.fields.MockUnmarshalYAMLNodes,
				MockString:             tt.fields.MockString,
				MockIsScalar:           tt.fields.MockIsScalar,
			}
			if err := u.check(); (err != nil) != tt.wantErr {
				t.Errorf("check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockShape_clone(t *testing.T) {
	type fields struct {
		BaseShape              *BaseShape
		MockInherit            func(source Shape) (Shape, error)
		MockCheck              func() error
		MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
		MockValidate           func(v interface{}, ctxPath string) error
		MockUnmarshalYAMLNodes func(v []*yaml.Node) error
		MockString             func() string
		MockIsScalar           func() bool
	}
	type args struct {
		base      *BaseShape
		clonedMap map[int64]*BaseShape
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Shape
	}{
		{
			name: "clone returns new shape",
			fields: fields{
				MockClone: func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
					return &MockShape{}
				},
			},
			args: args{
				base:      &BaseShape{},
				clonedMap: make(map[int64]*BaseShape),
			},
			want: &MockShape{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := MockShape{
				BaseShape:              tt.fields.BaseShape,
				MockInherit:            tt.fields.MockInherit,
				MockCheck:              tt.fields.MockCheck,
				MockClone:              tt.fields.MockClone,
				MockValidate:           tt.fields.MockValidate,
				MockUnmarshalYAMLNodes: tt.fields.MockUnmarshalYAMLNodes,
				MockString:             tt.fields.MockString,
				MockIsScalar:           tt.fields.MockIsScalar,
			}
			if got := u.clone(tt.args.base, tt.args.clonedMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockShape_inherit(t *testing.T) {
	type fields struct {
		BaseShape              *BaseShape
		MockInherit            func(source Shape) (Shape, error)
		MockCheck              func() error
		MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
		MockValidate           func(v interface{}, ctxPath string) error
		MockUnmarshalYAMLNodes func(v []*yaml.Node) error
		MockString             func() string
		MockIsScalar           func() bool
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
			name: "inherit returns new shape",
			fields: fields{
				MockInherit: func(source Shape) (Shape, error) {
					return &MockShape{}, nil
				},
			},
			args: args{
				source: &MockShape{},
			},
			want:    &MockShape{},
			wantErr: false,
		},
		{
			name: "inherit returns error",
			fields: fields{
				MockInherit: func(source Shape) (Shape, error) {
					return nil, fmt.Errorf("mock error")
				},
			},
			args: args{
				source: &MockShape{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := MockShape{
				BaseShape:              tt.fields.BaseShape,
				MockInherit:            tt.fields.MockInherit,
				MockCheck:              tt.fields.MockCheck,
				MockClone:              tt.fields.MockClone,
				MockValidate:           tt.fields.MockValidate,
				MockUnmarshalYAMLNodes: tt.fields.MockUnmarshalYAMLNodes,
				MockString:             tt.fields.MockString,
				MockIsScalar:           tt.fields.MockIsScalar,
			}
			got, err := u.inherit(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("inherit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("inherit() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockShape_unmarshalYAMLNodes(t *testing.T) {
	type fields struct {
		BaseShape              *BaseShape
		MockInherit            func(source Shape) (Shape, error)
		MockCheck              func() error
		MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
		MockValidate           func(v interface{}, ctxPath string) error
		MockUnmarshalYAMLNodes func(v []*yaml.Node) error
		MockString             func() string
		MockIsScalar           func() bool
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
			name: "unmarshalYAMLNodes returns no error",
			fields: fields{
				MockUnmarshalYAMLNodes: func(v []*yaml.Node) error { return nil },
			},
			args: args{
				v: []*yaml.Node{},
			},
			wantErr: false,
		},
		{
			name: "unmarshalYAMLNodes returns error",
			fields: fields{
				MockUnmarshalYAMLNodes: func(v []*yaml.Node) error { return fmt.Errorf("mock error") },
			},
			args: args{
				v: []*yaml.Node{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := MockShape{
				BaseShape:              tt.fields.BaseShape,
				MockInherit:            tt.fields.MockInherit,
				MockCheck:              tt.fields.MockCheck,
				MockClone:              tt.fields.MockClone,
				MockValidate:           tt.fields.MockValidate,
				MockUnmarshalYAMLNodes: tt.fields.MockUnmarshalYAMLNodes,
				MockString:             tt.fields.MockString,
				MockIsScalar:           tt.fields.MockIsScalar,
			}
			if err := u.unmarshalYAMLNodes(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalYAMLNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockShape_validate(t *testing.T) {
	type fields struct {
		BaseShape              *BaseShape
		MockInherit            func(source Shape) (Shape, error)
		MockCheck              func() error
		MockClone              func(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
		MockValidate           func(v interface{}, ctxPath string) error
		MockUnmarshalYAMLNodes func(v []*yaml.Node) error
		MockString             func() string
		MockIsScalar           func() bool
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
			name: "validate returns no error",
			fields: fields{
				MockValidate: func(v interface{}, ctxPath string) error { return nil },
			},
			args: args{
				v:       nil,
				ctxPath: "",
			},
			wantErr: false,
		},
		{
			name: "validate returns error",
			fields: fields{
				MockValidate: func(v interface{}, ctxPath string) error { return fmt.Errorf("mock error") },
			},
			args: args{
				v:       nil,
				ctxPath: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := MockShape{
				BaseShape:              tt.fields.BaseShape,
				MockInherit:            tt.fields.MockInherit,
				MockCheck:              tt.fields.MockCheck,
				MockClone:              tt.fields.MockClone,
				MockValidate:           tt.fields.MockValidate,
				MockUnmarshalYAMLNodes: tt.fields.MockUnmarshalYAMLNodes,
				MockString:             tt.fields.MockString,
				MockIsScalar:           tt.fields.MockIsScalar,
			}
			if err := u.validate(tt.args.v, tt.args.ctxPath); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
