package raml

import (
	"container/list"
	"context"
	"testing"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

func TestExample_decode(t *testing.T) {
	type fields struct {
		ID                     string
		Name                   string
		DisplayName            string
		Description            string
		Data                   *Node
		Strict                 bool
		CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]
		Location               string
		Position               stacktrace.Position
		raml                   *RAML
	}
	type args struct {
		node      *yaml.Node
		valueNode *yaml.Node
		location  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case: strict",
			fields: fields{
				Strict: false,
				raml:   &RAML{},
			},
			args: args{
				node: &yaml.Node{
					Value: "strict",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "true",
				},
				location: "location",
			},
			wantErr: false,
		},
		{
			name: "positive case: displayName",
			fields: fields{
				DisplayName: "",
			},
			args: args{
				node: &yaml.Node{
					Value: "displayName",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "display name",
				},
			},
			wantErr: false,
		},
		{
			name: "positive case: description",
			fields: fields{
				Description: "",
			},
			args: args{
				node: &yaml.Node{
					Value: "description",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "description",
				},
			},
			wantErr: false,
		},
		{
			name: "positive case: custom domain extension",
			fields: fields{
				raml:                   &RAML{},
				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
			},
			args: args{
				node: &yaml.Node{
					Value: "(custom)",
				},
				valueNode: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "key",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "value",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "negative case: invalid strict",
			fields: fields{
				Strict: false,
				raml:   &RAML{},
			},
			args: args{
				node: &yaml.Node{
					Value: "strict",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.MappingNode,
					Value: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: invalid displayName",
			fields: fields{
				DisplayName: "",
			},
			args: args{
				node: &yaml.Node{
					Value: "displayName",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.MappingNode,
					Value: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: invalid description",
			fields: fields{
				Description: "",
			},
			args: args{
				node: &yaml.Node{
					Value: "description",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.MappingNode,
					Value: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: invalid custom domain extension",
			fields: fields{
				raml:                   &RAML{},
				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
			},
			args: args{
				node: &yaml.Node{
					Value: "()",
				},
				valueNode: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "key",
							Tag:   "!!int",
						},
						{
							Kind:  yaml.MappingNode,
							Value: "invalid",
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
			ex := &Example{
				ID:                     tt.fields.ID,
				Name:                   tt.fields.Name,
				DisplayName:            tt.fields.DisplayName,
				Description:            tt.fields.Description,
				Data:                   tt.fields.Data,
				Strict:                 tt.fields.Strict,
				CustomDomainProperties: tt.fields.CustomDomainProperties,
				Location:               tt.fields.Location,
				Position:               tt.fields.Position,
				raml:                   tt.fields.raml,
			}
			if err := ex.decode(tt.args.node, tt.args.valueNode, tt.args.location); (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExample_fill(t *testing.T) {
	type fields struct {
		ID                     string
		Name                   string
		DisplayName            string
		Description            string
		Data                   *Node
		Strict                 bool
		CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]
		Location               string
		Position               stacktrace.Position
		raml                   *RAML
	}
	type args struct {
		location string
		value    *yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive case: fill example from mapping node",
			fields: fields{
				raml: &RAML{},
			},
			args: args{
				location: "location",
				value: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "value",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "example",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "negative case: value key not found",
			fields: fields{
				raml: &RAML{},
			},
			args: args{
				location: "location",
				value: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "key",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "example",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: decode example",
			fields: fields{
				raml: &RAML{},
			},
			args: args{
				location: "location",
				value: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "value",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "123",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "strict",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "123",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: make node error",
			fields: fields{
				raml: &RAML{},
			},
			args: args{
				location: "location",
				value: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "value",
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "key",
									Tag:   "!!int",
								},
								{
									Kind:  yaml.ScalarNode,
									Value: "value",
									Tag:   "!!int",
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
			ex := &Example{
				ID:                     tt.fields.ID,
				Name:                   tt.fields.Name,
				DisplayName:            tt.fields.DisplayName,
				Description:            tt.fields.Description,
				Data:                   tt.fields.Data,
				Strict:                 tt.fields.Strict,
				CustomDomainProperties: tt.fields.CustomDomainProperties,
				Location:               tt.fields.Location,
				Position:               tt.fields.Position,
				raml:                   tt.fields.raml,
			}
			if err := ex.fill(tt.args.location, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("fill() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_makeExample(t *testing.T) {
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
		value    *yaml.Node
		name     string
		location string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got *Example) (string, bool)
		wantErr bool
	}{
		{
			name:   "positive case",
			fields: fields{},
			args: args{
				value: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "value",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "example",
						},
					},
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
			got, err := r.makeExample(tt.args.value, tt.args.name, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeExample() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if msg, ok := tt.want(got); !ok {
					t.Errorf("makeExample() case hasn't been passed: %s", msg)
				}
			}
		})
	}
}
