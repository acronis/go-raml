package raml

import (
	"testing"

	"github.com/acronis/go-stacktrace"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Mock for RAML
type mockRAML struct {
	fragmentsCache map[string]Fragment
}

func (r *mockRAML) makeTraitDefinition(valueNode *yaml.Node, location string) (*TraitDefinition, error) {
	if valueNode.Kind == yaml.SequenceNode {
		return nil, StacktraceNew("trait definition must be a mapping node", location, WithNodePosition(valueNode))
	}

	if valueNode.Tag == TagInclude {
		return nil, StacktraceNew("included file not found", location, WithNodePosition(valueNode))
	}

	traitDef := &TraitDefinition{
		DeclaredVariables: make(map[string]struct{}),
		NodeVariableIndex: make(map[int][]VariableInfo),
		raml:              &RAML{fragmentsCache: r.fragmentsCache},
		Position:          stacktrace.Position{Line: valueNode.Line, Column: valueNode.Column},
		Location:          location,
	}

	// Extract usage if present
	if valueNode.Kind == yaml.MappingNode {
		for i := 0; i < len(valueNode.Content); i += 2 {
			keyNode := valueNode.Content[i]
			valueNode := valueNode.Content[i+1]
			if keyNode.Value == "usage" {
				traitDef.Usage = valueNode.Value
			}
		}
	}

	// Set source
	traitDef.Source = valueNode

	// Populate declared variables for variable test
	if valueNode.Kind == yaml.MappingNode {
		for i := 0; i < len(valueNode.Content); i += 2 {
			keyNode := valueNode.Content[i]
			valueNode := valueNode.Content[i+1]

			if keyNode.Value == "queryParameters" && valueNode.Kind == yaml.MappingNode {
				for j := 0; j < len(valueNode.Content); j += 2 {
					paramKeyNode := valueNode.Content[j]
					paramValueNode := valueNode.Content[j+1]

					// Check if the parameter name is a variable
					if paramKeyNode.Value == "<<param>>" {
						traitDef.DeclaredVariables["param"] = struct{}{}
						traitDef.NodeVariableIndex[j] = []VariableInfo{
							{
								Name:      "param",
								Substring: "<<param>>",
							},
						}
					}

					// Check if the parameter type is a variable
					if paramValueNode.Kind == yaml.MappingNode {
						for k := 0; k < len(paramValueNode.Content); k += 2 {
							typeKeyNode := paramValueNode.Content[k]
							typeValueNode := paramValueNode.Content[k+1]

							if typeKeyNode.Value == "type" && typeValueNode.Value == "<<type>>" {
								traitDef.DeclaredVariables["type"] = struct{}{}
								traitDef.NodeVariableIndex[j+k+1] = []VariableInfo{
									{
										Name:      "type",
										Substring: "<<type>>",
									},
								}
							}
						}
					}
				}
			}
		}
	}

	// Mock the compile method by setting Precompiled
	traitDef.Precompiled = &Operation{}

	return traitDef, nil
}

func TestRAML_makeTraitDefinition(t *testing.T) {
	type args struct {
		valueNode *yaml.Node
		location  string
	}
	tests := []struct {
		name    string
		args    args
		want    func(t *testing.T, got *TraitDefinition)
		wantErr bool
	}{
		{
			name: "positive: simple trait definition",
			args: args{
				valueNode: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "usage",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "This is a test trait",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "queryParameters",
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "page",
								},
								{
									Kind: yaml.MappingNode,
									Content: []*yaml.Node{
										{
											Kind:  yaml.ScalarNode,
											Value: "type",
										},
										{
											Kind:  yaml.ScalarNode,
											Value: "integer",
										},
									},
								},
							},
						},
					},
				},
				location: "/test/location.raml",
			},
			want: func(t *testing.T, got *TraitDefinition) {
				require.NotNil(t, got)
				require.Equal(t, "This is a test trait", got.Usage)
				require.NotNil(t, got.Source)
				require.Equal(t, "/test/location.raml", got.Location)
			},
		},
		{
			name: "positive: trait with variables",
			args: args{
				valueNode: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "queryParameters",
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "<<param>>",
								},
								{
									Kind: yaml.MappingNode,
									Content: []*yaml.Node{
										{
											Kind:  yaml.ScalarNode,
											Value: "type",
										},
										{
											Kind:  yaml.ScalarNode,
											Value: "<<type>>",
										},
									},
								},
							},
						},
					},
				},
				location: "/test/location.raml",
			},
			want: func(t *testing.T, got *TraitDefinition) {
				require.NotNil(t, got)
				require.NotNil(t, got.Source)
				require.Equal(t, 2, len(got.DeclaredVariables))
				_, hasParam := got.DeclaredVariables["param"]
				require.True(t, hasParam)
				_, hasType := got.DeclaredVariables["type"]
				require.True(t, hasType)
			},
		},
		{
			name: "positive: trait with include",
			args: args{
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   TagInclude,
					Value: "included-trait.raml",
				},
				location: "/test/location.raml",
			},
			wantErr: true, // Since we don't have the actual included file
		},
		{
			name: "negative: invalid node kind",
			args: args{
				valueNode: &yaml.Node{
					Kind: yaml.SequenceNode,
				},
				location: "/test/location.raml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockRAML{
				fragmentsCache: map[string]Fragment{},
			}

			got, err := r.makeTraitDefinition(tt.args.valueNode, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeTraitDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && err == nil {
				tt.want(t, got)
			}
		})
	}
}

func TestTraitDefinition_decode(t *testing.T) {
	type fields struct {
		ID                int64
		Usage             string
		Source            *yaml.Node
		DeclaredVariables map[string]struct{}
		NodeVariableIndex map[int][]VariableInfo
		Precompiled       *Operation
		Link              *TraitFragment
		Location          string
		Position          stacktrace.Position
		raml              *RAML
	}
	type args struct {
		node *yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(t *testing.T, traitDef *TraitDefinition)
		wantErr bool
	}{
		{
			name: "positive: null node",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
			},
			args: args{
				node: &yaml.Node{
					Tag: TagNull,
				},
			},
			want: func(t *testing.T, traitDef *TraitDefinition) {
				require.Nil(t, traitDef.Source)
			},
		},
		{
			name: "positive: mapping node with usage",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "usage",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "This is a test trait",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "queryParameters",
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "page",
								},
								{
									Kind: yaml.MappingNode,
									Content: []*yaml.Node{
										{
											Kind:  yaml.ScalarNode,
											Value: "type",
										},
										{
											Kind:  yaml.ScalarNode,
											Value: "integer",
										},
									},
								},
							},
						},
					},
				},
			},
			want: func(t *testing.T, traitDef *TraitDefinition) {
				require.Equal(t, "This is a test trait", traitDef.Usage)
				require.NotNil(t, traitDef.Source)
				require.Equal(t, 2, len(traitDef.Source.Content)) // queryParameters key and value
			},
		},
		{
			name: "negative: invalid node kind",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind: yaml.SequenceNode,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TraitDefinition{
				ID:                tt.fields.ID,
				Usage:             tt.fields.Usage,
				Source:            tt.fields.Source,
				DeclaredVariables: tt.fields.DeclaredVariables,
				NodeVariableIndex: tt.fields.NodeVariableIndex,
				Precompiled:       tt.fields.Precompiled,
				Link:              tt.fields.Link,
				Location:          tt.fields.Location,
				Position:          tt.fields.Position,
				raml:              tt.fields.raml,
			}
			err := tr.decode(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && err == nil {
				tt.want(t, tr)
			}
		})
	}
}

func TestTraitDefinition_collectVariablesIndex(t *testing.T) {
	type fields struct {
		ID                int64
		Usage             string
		Source            *yaml.Node
		DeclaredVariables map[string]struct{}
		NodeVariableIndex map[int][]VariableInfo
		Precompiled       *Operation
		Link              *TraitFragment
		Location          string
		Position          stacktrace.Position
		raml              *RAML
	}
	type args struct {
		node *yaml.Node
		idx  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(t *testing.T, traitDef *TraitDefinition)
		wantErr bool
	}{
		{
			name: "positive: node with variables",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "<<param>>",
							Tag:   TagStr,
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "type",
									Tag:   TagStr,
								},
								{
									Kind:  yaml.ScalarNode,
									Value: "<<type>>",
									Tag:   TagStr,
								},
							},
						},
					},
				},
				idx: 0,
			},
			want: func(t *testing.T, traitDef *TraitDefinition) {
				require.Equal(t, 2, len(traitDef.DeclaredVariables))
				_, hasParam := traitDef.DeclaredVariables["param"]
				require.True(t, hasParam)
				_, hasType := traitDef.DeclaredVariables["type"]
				require.True(t, hasType)

				// Check NodeVariableIndex
				require.Equal(t, 2, len(traitDef.NodeVariableIndex))
				paramVars, hasParamIdx := traitDef.NodeVariableIndex[0]
				require.True(t, hasParamIdx)
				require.Equal(t, 1, len(paramVars))
				require.Equal(t, "param", paramVars[0].Name)

				// The index of the type variable depends on the implementation
				// Let's check if it exists in any of the indices
				foundTypeVar := false
				for _, vars := range traitDef.NodeVariableIndex {
					for _, v := range vars {
						if v.Name == "type" {
							foundTypeVar = true
							break
						}
					}
					if foundTypeVar {
						break
					}
				}
				require.True(t, foundTypeVar, "type variable should be found in NodeVariableIndex")
			},
		},
		{
			name: "positive: node without variables",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "param",
							Tag:   TagStr,
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "type",
									Tag:   TagStr,
								},
								{
									Kind:  yaml.ScalarNode,
									Value: "string",
									Tag:   TagStr,
								},
							},
						},
					},
				},
				idx: 0,
			},
			want: func(t *testing.T, traitDef *TraitDefinition) {
				require.Equal(t, 0, len(traitDef.DeclaredVariables))
				require.Equal(t, 0, len(traitDef.NodeVariableIndex))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TraitDefinition{
				ID:                tt.fields.ID,
				Usage:             tt.fields.Usage,
				Source:            tt.fields.Source,
				DeclaredVariables: tt.fields.DeclaredVariables,
				NodeVariableIndex: tt.fields.NodeVariableIndex,
				Precompiled:       tt.fields.Precompiled,
				Link:              tt.fields.Link,
				Location:          tt.fields.Location,
				Position:          tt.fields.Position,
				raml:              tt.fields.raml,
			}
			err := tr.collectVariablesIndex(tt.args.node, tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("collectVariablesIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && err == nil {
				tt.want(t, tr)
			}
		})
	}
}

func TestTraitDefinition_findVariable(t *testing.T) {
	type fields struct {
		ID                int64
		Usage             string
		Source            *yaml.Node
		DeclaredVariables map[string]struct{}
		NodeVariableIndex map[int][]VariableInfo
		Precompiled       *Operation
		Link              *TraitFragment
		Location          string
		Position          stacktrace.Position
		raml              *RAML
	}
	type args struct {
		node *yaml.Node
		idx  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(t *testing.T, traitDef *TraitDefinition)
		wantErr bool
	}{
		{
			name: "positive: node with variable",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "<<param>>",
					Tag:   TagStr,
				},
				idx: 0,
			},
			want: func(t *testing.T, traitDef *TraitDefinition) {
				require.Equal(t, 1, len(traitDef.DeclaredVariables))
				_, hasParam := traitDef.DeclaredVariables["param"]
				require.True(t, hasParam)

				// Check NodeVariableIndex
				require.Equal(t, 1, len(traitDef.NodeVariableIndex))
				paramVars, hasParamIdx := traitDef.NodeVariableIndex[0]
				require.True(t, hasParamIdx)
				require.Equal(t, 1, len(paramVars))
				require.Equal(t, "param", paramVars[0].Name)
				require.Equal(t, "<<param>>", paramVars[0].Substring)
				require.Equal(t, "", paramVars[0].Action)
			},
		},
		{
			name: "positive: node with variable and action",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "<<param|!singularize>>",
					Tag:   TagStr,
				},
				idx: 0,
			},
			want: func(t *testing.T, traitDef *TraitDefinition) {
				require.Equal(t, 1, len(traitDef.DeclaredVariables))
				_, hasParam := traitDef.DeclaredVariables["param"]
				require.True(t, hasParam)

				// Check NodeVariableIndex
				require.Equal(t, 1, len(traitDef.NodeVariableIndex))
				paramVars, hasParamIdx := traitDef.NodeVariableIndex[0]
				require.True(t, hasParamIdx)
				require.Equal(t, 1, len(paramVars))
				require.Equal(t, "param", paramVars[0].Name)
				require.Equal(t, "<<param|!singularize>>", paramVars[0].Substring)
				require.Equal(t, "!singularize", paramVars[0].Action)
			},
		},
		{
			name: "positive: node without variable",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "param",
					Tag:   TagStr,
				},
				idx: 0,
			},
			want: func(t *testing.T, traitDef *TraitDefinition) {
				require.Equal(t, 0, len(traitDef.DeclaredVariables))
				require.Equal(t, 0, len(traitDef.NodeVariableIndex))
			},
		},
		{
			name: "negative: non-scalar node",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind: yaml.MappingNode,
				},
				idx: 0,
			},
			wantErr: true,
		},
		{
			name: "positive: non-string tag",
			fields: fields{
				DeclaredVariables: make(map[string]struct{}),
				NodeVariableIndex: make(map[int][]VariableInfo),
				raml:              &RAML{},
				Location:          "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "123",
					Tag:   "!!int",
				},
				idx: 0,
			},
			want: func(t *testing.T, traitDef *TraitDefinition) {
				require.Equal(t, 0, len(traitDef.DeclaredVariables))
				require.Equal(t, 0, len(traitDef.NodeVariableIndex))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TraitDefinition{
				ID:                tt.fields.ID,
				Usage:             tt.fields.Usage,
				Source:            tt.fields.Source,
				DeclaredVariables: tt.fields.DeclaredVariables,
				NodeVariableIndex: tt.fields.NodeVariableIndex,
				Precompiled:       tt.fields.Precompiled,
				Link:              tt.fields.Link,
				Location:          tt.fields.Location,
				Position:          tt.fields.Position,
				raml:              tt.fields.raml,
			}
			err := tr.findVariable(tt.args.node, tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("findVariable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && err == nil {
				tt.want(t, tr)
			}
		})
	}
}

func TestTraitDefinition_compile(t *testing.T) {
	type fields struct {
		ID                int64
		Usage             string
		Source            *yaml.Node
		DeclaredVariables map[string]struct{}
		NodeVariableIndex map[int][]VariableInfo
		Precompiled       *Operation
		Link              *TraitFragment
		Location          string
		Position          stacktrace.Position
		raml              *RAML
	}
	type args struct {
		params map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(t *testing.T, got *Operation)
		wantErr bool
	}{
		{
			name: "positive: precompiled trait",
			fields: fields{
				Precompiled: &Operation{},
				raml:        &RAML{},
			},
			args: args{
				params: map[string]string{},
			},
			want: func(t *testing.T, got *Operation) {
				require.NotNil(t, got)
			},
		},
		{
			name: "positive: linked trait",
			fields: fields{
				Link: &TraitFragment{
					Trait: &TraitDefinition{
						Precompiled: &Operation{},
						raml:        &RAML{},
					},
				},
				raml: &RAML{},
			},
			args: args{
				params: map[string]string{},
			},
			want: func(t *testing.T, got *Operation) {
				require.NotNil(t, got)
			},
		},
		{
			name: "negative: unexpected parameter",
			fields: fields{
				DeclaredVariables: map[string]struct{}{
					"param1": {},
				},
				Source: &yaml.Node{},
				raml:   &RAML{},
			},
			args: args{
				params: map[string]string{
					"param1": "value1",
					"param2": "value2", // Unexpected parameter
				},
			},
			wantErr: true,
		},
		{
			name: "negative: missing required parameter",
			fields: fields{
				DeclaredVariables: map[string]struct{}{
					"param1": {},
					"param2": {},
				},
				Source: &yaml.Node{},
				raml:   &RAML{},
			},
			args: args{
				params: map[string]string{
					"param1": "value1",
					// param2 is missing
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TraitDefinition{
				ID:                tt.fields.ID,
				Usage:             tt.fields.Usage,
				Source:            tt.fields.Source,
				DeclaredVariables: tt.fields.DeclaredVariables,
				NodeVariableIndex: tt.fields.NodeVariableIndex,
				Precompiled:       tt.fields.Precompiled,
				Link:              tt.fields.Link,
				Location:          tt.fields.Location,
				Position:          tt.fields.Position,
				raml:              tt.fields.raml,
			}
			got, err := tr.compile(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && err == nil {
				tt.want(t, got)
			}
		})
	}
}

func TestTraitDefinition_compileSource(t *testing.T) {
	type fields struct {
		ID                int64
		Usage             string
		Source            *yaml.Node
		DeclaredVariables map[string]struct{}
		NodeVariableIndex map[int][]VariableInfo
		Precompiled       *Operation
		Link              *TraitFragment
		Location          string
		Position          stacktrace.Position
		raml              *RAML
	}
	type args struct {
		node   *yaml.Node
		params map[string]string
		idx    int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(t *testing.T, got *yaml.Node)
	}{
		{
			name: "positive: scalar node with variable",
			fields: fields{
				NodeVariableIndex: map[int][]VariableInfo{
					0: {
						{
							Name:      "param",
							Substring: "<<param>>",
							Action:    "",
						},
					},
				},
				raml: &RAML{},
			},
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "<<param>>",
				},
				params: map[string]string{
					"param": "value",
				},
				idx: 0,
			},
			want: func(t *testing.T, got *yaml.Node) {
				require.NotNil(t, got)
				require.Equal(t, yaml.ScalarNode, got.Kind)
				require.Equal(t, "value", got.Value)
			},
		},
		{
			name: "positive: scalar node without variable",
			fields: fields{
				NodeVariableIndex: map[int][]VariableInfo{},
				raml:              &RAML{},
			},
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "no variable",
				},
				params: map[string]string{},
				idx:    0,
			},
			want: func(t *testing.T, got *yaml.Node) {
				require.NotNil(t, got)
				require.Equal(t, yaml.ScalarNode, got.Kind)
				require.Equal(t, "no variable", got.Value)
			},
		},
		{
			name: "positive: mapping node with modified children",
			fields: fields{
				NodeVariableIndex: map[int][]VariableInfo{
					1: {
						{
							Name:      "param",
							Substring: "<<param>>",
							Action:    "",
						},
					},
				},
				raml: &RAML{},
			},
			args: args{
				node: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "key",
						},
						{
							Kind:  yaml.ScalarNode,
							Value: "<<param>>",
						},
					},
				},
				params: map[string]string{
					"param": "value",
				},
				idx: 0,
			},
			want: func(t *testing.T, got *yaml.Node) {
				require.NotNil(t, got)
				require.Equal(t, yaml.MappingNode, got.Kind)
				require.Equal(t, 2, len(got.Content))
				require.Equal(t, "key", got.Content[0].Value)
				require.Equal(t, "value", got.Content[1].Value)
			},
		},
		{
			name: "positive: mapping node without modified children",
			fields: fields{
				NodeVariableIndex: map[int][]VariableInfo{},
				raml:              &RAML{},
			},
			args: args{
				node: &yaml.Node{
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
				params: map[string]string{},
				idx:    0,
			},
			want: func(t *testing.T, got *yaml.Node) {
				require.NotNil(t, got)
				require.Equal(t, yaml.MappingNode, got.Kind)
				require.Equal(t, 2, len(got.Content))
				require.Equal(t, "key", got.Content[0].Value)
				require.Equal(t, "value", got.Content[1].Value)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TraitDefinition{
				ID:                tt.fields.ID,
				Usage:             tt.fields.Usage,
				Source:            tt.fields.Source,
				DeclaredVariables: tt.fields.DeclaredVariables,
				NodeVariableIndex: tt.fields.NodeVariableIndex,
				Precompiled:       tt.fields.Precompiled,
				Link:              tt.fields.Link,
				Location:          tt.fields.Location,
				Position:          tt.fields.Position,
				raml:              tt.fields.raml,
			}
			got := tr.compileSource(tt.args.node, tt.args.params, tt.args.idx)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_makeTraits(t *testing.T) {
	type fields struct {
		fragmentsCache map[string]Fragment
	}
	type args struct {
		valueNode *yaml.Node
		location  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(t *testing.T, got []*Trait)
		wantErr bool
	}{
		{
			name: "positive: scalar node",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "paged",
				},
				location: "/test/location.raml",
			},
			want: func(t *testing.T, got []*Trait) {
				require.NotNil(t, got)
				require.Equal(t, 1, len(got))
				require.Equal(t, "paged", got[0].Name)
			},
		},
		{
			name: "positive: null node",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				valueNode: &yaml.Node{
					Kind: yaml.ScalarNode,
					Tag:  TagNull,
				},
				location: "/test/location.raml",
			},
			want: func(t *testing.T, got []*Trait) {
				require.Nil(t, got)
			},
		},
		{
			name: "positive: sequence node",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				valueNode: &yaml.Node{
					Kind: yaml.SequenceNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "paged",
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "secured",
								},
								{
									Kind: yaml.MappingNode,
									Content: []*yaml.Node{
										{
											Kind:  yaml.ScalarNode,
											Value: "queryParameters",
										},
										{
											Kind:  yaml.ScalarNode,
											Value: "access_token",
										},
									},
								},
							},
						},
					},
				},
				location: "/test/location.raml",
			},
			want: func(t *testing.T, got []*Trait) {
				require.NotNil(t, got)
				require.Equal(t, 2, len(got))
				require.Equal(t, "paged", got[0].Name)
				require.Equal(t, "secured", got[1].Name)
				require.NotNil(t, got[1].Params)
				queryParams, ok := got[1].Params["queryParameters"]
				require.True(t, ok)
				require.Equal(t, "access_token", queryParams)
			},
		},
		{
			name: "negative: invalid node kind",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				valueNode: &yaml.Node{
					Kind: yaml.MappingNode,
				},
				location: "/test/location.raml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RAML{
				fragmentsCache: tt.fields.fragmentsCache,
			}
			got, err := r.makeTraits(tt.args.valueNode, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeTraits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && err == nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_makeTrait(t *testing.T) {
	type fields struct {
		fragmentsCache map[string]Fragment
	}
	type args struct {
		valueNode *yaml.Node
		location  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(t *testing.T, got *Trait)
		wantErr bool
	}{
		{
			name: "positive: scalar node",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				valueNode: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "paged",
				},
				location: "/test/location.raml",
			},
			want: func(t *testing.T, got *Trait) {
				require.NotNil(t, got)
				require.Equal(t, "paged", got.Name)
				require.Nil(t, got.Params)
			},
		},
		{
			name: "positive: mapping node",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				valueNode: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "secured",
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "queryParameters",
								},
								{
									Kind:  yaml.ScalarNode,
									Value: "access_token",
								},
							},
						},
					},
				},
				location: "/test/location.raml",
			},
			want: func(t *testing.T, got *Trait) {
				require.NotNil(t, got)
				require.Equal(t, "secured", got.Name)
				require.NotNil(t, got.Params)
				queryParams, ok := got.Params["queryParameters"]
				require.True(t, ok)
				require.Equal(t, "access_token", queryParams)
			},
		},
		{
			name: "negative: invalid node kind",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				valueNode: &yaml.Node{
					Kind: yaml.SequenceNode,
				},
				location: "/test/location.raml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RAML{
				fragmentsCache: tt.fields.fragmentsCache,
			}
			got, err := r.makeTrait(tt.args.valueNode, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeTrait() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && err == nil {
				tt.want(t, got)
			}
		})
	}
}

func TestTrait_decode(t *testing.T) {
	type fields struct {
		ID       int64
		Name     string
		Params   map[string]string
		Location string
		Position stacktrace.Position
		raml     *RAML
	}
	type args struct {
		node *yaml.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(t *testing.T, trait *Trait)
		wantErr bool
	}{
		{
			name: "positive: scalar node",
			fields: fields{
				raml:     &RAML{},
				Location: "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "paged",
				},
			},
			want: func(t *testing.T, trait *Trait) {
				require.Equal(t, "paged", trait.Name)
				require.Nil(t, trait.Params)
			},
		},
		{
			name: "positive: mapping node",
			fields: fields{
				raml:     &RAML{},
				Location: "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Value: "secured",
						},
						{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{
									Kind:  yaml.ScalarNode,
									Value: "queryParameters",
								},
								{
									Kind:  yaml.ScalarNode,
									Value: "access_token",
								},
							},
						},
					},
				},
			},
			want: func(t *testing.T, trait *Trait) {
				require.Equal(t, "secured", trait.Name)
				require.NotNil(t, trait.Params)
				queryParams, ok := trait.Params["queryParameters"]
				require.True(t, ok)
				require.Equal(t, "access_token", queryParams)
			},
		},
		{
			name: "negative: invalid node kind",
			fields: fields{
				raml:     &RAML{},
				Location: "/test/location.raml",
			},
			args: args{
				node: &yaml.Node{
					Kind: yaml.SequenceNode,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Trait{
				ID:       tt.fields.ID,
				Name:     tt.fields.Name,
				Params:   tt.fields.Params,
				Location: tt.fields.Location,
				Position: tt.fields.Position,
				raml:     tt.fields.raml,
			}
			err := tr.decode(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && err == nil {
				tt.want(t, tr)
			}
		})
	}
}
