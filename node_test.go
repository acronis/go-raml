package raml

import (
	"container/list"
	"context"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/acronis/go-stacktrace"
	"gopkg.in/yaml.v3"
)

func TestNodes_String(t *testing.T) {
	tests := []struct {
		name string
		n    Nodes
		want string
	}{
		{
			name: "full positive case",
			n: Nodes{
				{
					ID:       "1",
					Value:    "value1",
					Location: "location1",
					Position: stacktrace.Position{Line: 1, Column: 1},
				},
				{
					ID:       "2",
					Value:    "value2",
					Location: "location2",
					Position: stacktrace.Position{Line: 2, Column: 2},
				},
			},
			want: "value1, value2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_String(t *testing.T) {
	type fields struct {
		ID       string
		Value    any
		Location string
		Position stacktrace.Position
		raml     *RAML
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "full positive case",
			fields: fields{
				ID:       "1",
				Value:    "value1",
				Location: "location1",
				Position: stacktrace.Position{Line: 1, Column: 1},
			},
			want: "value1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				ID:       tt.fields.ID,
				Value:    tt.fields.Value,
				Location: tt.fields.Location,
				Position: tt.fields.Position,
				raml:     tt.fields.raml,
			}
			if got := n.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAML_makeIncludedNode(t *testing.T) {
	tempDir, errMkTmp := os.MkdirTemp(os.TempDir(), "go-raml-test-include-node-")
	if errMkTmp != nil {
		t.Fatalf("MkdirTemp() error = %v", errMkTmp)
	}

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
		node     *yaml.Node
		location string
	}

	tests := []struct {
		name    string
		prepare func(tt *testing.T)
		fields  fields
		args    args
		want    func(tt *testing.T, n *Node)
		wantErr bool
	}{
		{
			name: "full positive case: yaml node",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				node: &yaml.Node{
					Value: "filename.yaml",
					Line:  1,
				},
				location: tempDir + "/location.raml",
			},
			prepare: func(tt *testing.T) {
				if err := os.WriteFile(path.Join(tempDir, "filename.yaml"), []byte("key: value"), 0644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
		},
		{
			name: "full positive case: json node",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				node: &yaml.Node{
					Value: "filename.json",
					Line:  1,
				},
				location: tempDir + "/location.raml",
			},
			prepare: func(tt *testing.T) {
				if err := os.WriteFile(path.Join(tempDir, "filename.json"), []byte(`{"key": "value"}`), 0644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
		},
		{
			name: "full positive case: unknown extension",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				node: &yaml.Node{
					Value: "filename.unknown",
					Line:  1,
				},
				location: tempDir + "/location.raml",
			},
			prepare: func(tt *testing.T) {
				if err := os.WriteFile(path.Join(tempDir, "filename.unknown"), []byte("key: value"), 0644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
		},
		{
			name:   "negative case: file not found",
			fields: fields{},
			args: args{
				node: &yaml.Node{
					Value: "notfound.yaml",
					Line:  1,
				},
				location: tempDir + "/location.raml",
			},
			wantErr: true,
		},
		{
			name:   "negative case: json decode error",
			fields: fields{},
			args: args{
				node: &yaml.Node{
					Value: "err.json",
					Line:  1,
				},
				location: tempDir + "/location.raml",
			},
			prepare: func(tt *testing.T) {
				if err := os.WriteFile(path.Join(tempDir, "err.json"), []byte(`{"key": "value`), 0644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
			wantErr: true,
		},
		{
			name:   "negative case: yaml decode error",
			fields: fields{},
			args: args{
				node: &yaml.Node{
					Value: "err.yaml",
					Line:  1,
				},
				location: tempDir + "/location.raml",
			},
			prepare: func(tt *testing.T) {
				if err := os.WriteFile(path.Join(tempDir, "err.yaml"), []byte("key: value\nbad"), 0644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
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
				tt.prepare(t)
			}
			got, errInclude := r.makeIncludedNode(tt.args.node, tt.args.location)
			if (errInclude != nil) != tt.wantErr {
				t.Errorf("makeIncludedNode() error = %v, wantErr %v", errInclude, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
	if errRemove := os.RemoveAll(tempDir); errRemove != nil {
		t.Fatalf("RemoveAll() error = %v", errRemove)
	}
}

func TestRAML_makeRootNode(t *testing.T) {
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
		node     *yaml.Node
		location string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(*testing.T, *Node)
		wantErr bool
	}{
		{
			name: "full positive case: json unmarshal",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				node: &yaml.Node{
					Value: `{"key": "value"}`,
				},
				location: "location.raml",
			},
			want: func(t *testing.T, n *Node) {
				if !reflect.DeepEqual(n.Value, map[string]interface{}{"key": "value"}) {
					t.Errorf("makeRootNode() = %v, want %v", n.Value, map[string]interface{}{"key": "value"})
				}
			},
		},
		{
			name: "negative case: tag include",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				node: &yaml.Node{
					Tag: "!include",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: json unmarshal error",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				node: &yaml.Node{
					Value: `{"key": "value`,
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
			got, err := r.makeRootNode(tt.args.node, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeRootNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_makeYamlNode(t *testing.T) {
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
		node     *yaml.Node
		location string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(*testing.T, *Node)
		wantErr bool
	}{
		{
			name: "full positive case",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value",
				},
			},
			want: func(t *testing.T, n *Node) {
				if n.Value != "value" {
					t.Errorf("makeYamlNode() = %v, want %v", n.Value, "value")
				}
			},
		},
		{
			name: "negative case: yaml node to data node error: unexpected kind",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				node: &yaml.Node{
					Kind: 123,
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
			got, err := r.makeYamlNode(tt.args.node, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeYamlNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func Test_scalarNodeToDataNode(t *testing.T) {
	tempDir, errTmp := os.MkdirTemp(os.TempDir(), "go-raml-test-scalar-node-to-data-node-")
	if errTmp != nil {
		t.Fatalf("MkdirTemp() error = %v", errTmp)
	}
	type args struct {
		node      *yaml.Node
		location  string
		isInclude bool
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
		prepare func(tt *testing.T)
	}{
		{
			name: "positive case: decode default node: int",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "123",
					Tag:   "!!int",
				},
			},
			want: 123,
		},
		{
			name: "negative case: decode default node: int",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value",
					Tag:   "!!int",
				},
			},
			wantErr: true,
		},
		{
			name: "positive case: decode str node",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value",
					Tag:   "!!str",
				},
			},
			want: "value",
		},
		{
			name: "positive case: decode timestamp node",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "2021-09-01T00:00:00Z",
					Tag:   "!!timestamp",
				},
			},
			want: "2021-09-01T00:00:00Z",
		},
		{
			name: "positive case: decode include node: yaml",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "filename.yaml",
					Tag:   "!include",
				},
				location: tempDir + "/location.raml",
			},
			prepare: func(tt *testing.T) {
				if err := os.WriteFile(path.Join(tempDir, "filename.yaml"), []byte("key: value"), 0644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
			want: map[string]interface{}{"key": "value"},
		},
		{
			name: "positive case: decode include node: any",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "filename.txt",
					Tag:   "!include",
				},
				location: tempDir + "/location.raml",
			},
			prepare: func(tt *testing.T) {
				if err := os.WriteFile(path.Join(tempDir, "filename.txt"), []byte("Hello world!"), 0644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
			want: "Hello world!",
		},
		{
			name: "negative case: decode include node: file not found",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "notfound.yaml",
					Tag:   "!include",
				},
				location: tempDir + "/location.raml",
			},
			wantErr: true,
		},
		{
			name: "negative case: nested include are not allowed",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "filename.yaml",
					Tag:   "!include",
				},
				location:  tempDir + "/location.raml",
				isInclude: true,
			},
			wantErr: true,
		},
		{
			name: "negative case: decode included yaml node: bad indent",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "err.yaml",
					Tag:   "!include",
				},
				location: tempDir + "/location.raml",
			},
			prepare: func(tt *testing.T) {
				if err := os.WriteFile(path.Join(tempDir, "err.yaml"), []byte("\tkey: value\n bad: bad\n\t"), 0644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t)
			}
			got, err := scalarNodeToDataNode(tt.args.node, tt.args.location, tt.args.isInclude)
			if (err != nil) != tt.wantErr {
				t.Errorf("scalarNodeToDataNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scalarNodeToDataNode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_yamlNodeToDataNode(t *testing.T) {
	type args struct {
		node      *yaml.Node
		location  string
		isInclude bool
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "negative case: unexpected kind",
			args: args{
				node: &yaml.Node{
					Kind: 123,
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: alias nodes are not supported",
			args: args{
				node: &yaml.Node{
					Kind: yaml.AliasNode,
				},
			},
			wantErr: true,
		},
		{
			name: "positive case: document node",
			args: args{
				node: &yaml.Node{
					Kind: yaml.DocumentNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "value",
						},
					},
				},
			},
			want: "value",
		},
		{
			name: "positive case: scalar node",
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value",
				},
			},
			want: "value",
		},
		{
			name: "positive case: mapping node",
			args: args{
				node: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Value: "key",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "value",
						},
					},
				},
			},
			want: map[string]interface{}{"key": "value"},
		},
		{
			name: "negative case: mapping node should have even number of children",
			args: args{
				node: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Value: "key",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "positive case: sequence node",
			args: args{
				node: &yaml.Node{
					Kind: yaml.SequenceNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "value1",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "value2",
						},
					},
				},
			},
			want: []any{"value1", "value2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := yamlNodeToDataNode(tt.args.node, tt.args.location, tt.args.isInclude)
			if (err != nil) != tt.wantErr {
				t.Errorf("yamlNodeToDataNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("yamlNodeToDataNode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
