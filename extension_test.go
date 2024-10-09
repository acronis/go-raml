package raml

import (
	"container/list"
	"context"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRAML_unmarshalCustomDomainExtension(t *testing.T) {
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
		location  string
		keyNode   *yaml.Node
		valueNode *yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(got string, de *DomainExtension) (string, bool)
		wantErr bool
	}{
		{
			name: "positive case",
			args: args{
				location: "location",
				keyNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "(name)",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value",
				},
			},
			want: func(name string, de *DomainExtension) (string, bool) {
				if name != "name" {
					return "name is not equal to 'name'", false
				}
				if de.Name != "name" {
					return "de.Name is not equal to 'name'", false
				}
				if de.Extension.Value != "value" {
					return "extension value is not equal to 'value'", false
				}
				return "", true
			},
		},
		{
			name: "negative case: annotation name must not be empty",
			args: args{
				location: "location",
				keyNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "()",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value",
				},
			},
			wantErr: true,
		},
		{
			name: "negative case: make node error",
			args: args{
				location: "location",
				keyNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "(name)",
				},
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value",
					Tag:   "!!int",
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
			name, de, err := r.unmarshalCustomDomainExtension(tt.args.location, tt.args.keyNode, tt.args.valueNode)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalCustomDomainExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if msg, ok := tt.want(name, de); !ok {
					t.Errorf("unmarshalCustomDomainExtension() case hasn't been passed: %s", msg)
				}
			}
		})
	}
}
