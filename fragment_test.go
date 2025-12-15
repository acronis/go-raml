package raml

// func TestCutReferenceName(t *testing.T) {
// 	type args struct {
// 		refName string
// 	}
// 	tests := []struct {
// 		name  string
// 		args  args
// 		want  string
// 		want1 string
// 		want2 bool
// 	}{
// 		{
// 			name: "positive case",
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			want:  "fragment",
// 			want1: "identifier",
// 			want2: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got1, got2 := CutReferenceName(tt.args.refName)
// 			if got != tt.want {
// 				t.Errorf("CutReferenceName() got = %v, want %v", got, tt.want)
// 			}
// 			if got1 != tt.want1 {
// 				t.Errorf("CutReferenceName() got1 = %v, want %v", got1, tt.want1)
// 			}
// 			if got2 != tt.want2 {
// 				t.Errorf("CutReferenceName() got2 = %v, want %v", got2, tt.want2)
// 			}
// 		})
// 	}
// }

// func TestLibrary_GetReferenceType(t *testing.T) {
// 	type fields struct {
// 		ID                     string
// 		Usage                  string
// 		AnnotationTypes        *orderedmap.OrderedMap[string, *BaseShape]
// 		Types                  *orderedmap.OrderedMap[string, *BaseShape]
// 		Uses                   *orderedmap.OrderedMap[string, *LibraryLink]
// 		CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]
// 		Location               string
// 		raml                   *RAML
// 	}
// 	type args struct {
// 		refName string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive case: local ref",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Types: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 					m := orderedmap.New[string, *BaseShape](0)
// 					m.Set("identifier", &BaseShape{})
// 					return m
// 				}(),
// 				AnnotationTypes:        orderedmap.New[string, *BaseShape](0),
// 				Uses:                   orderedmap.New[string, *LibraryLink](0),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "identifier",
// 			},
// 			want: func(got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("GetReferenceType() got = %v, want %v", got, "not nil")
// 				}
// 			},
// 		},
// 		{
// 			name: "positive case: uses ref",
// 			fields: fields{
// 				ID:              "id",
// 				Usage:           "usage",
// 				Types:           orderedmap.New[string, *BaseShape](0),
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					uses.Set("fragment", &LibraryLink{
// 						Link: &Library{
// 							Types: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 								types := orderedmap.New[string, *BaseShape](0)
// 								types.Set("identifier", &BaseShape{})
// 								return types
// 							}(),
// 						},
// 					})
// 					return uses
// 				}(),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			want: func(got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("GetReferenceType() got = %v, want %v", got, "not nil")
// 				}
// 			},
// 		},
// 		{
// 			name: "negative case: local ref not found",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Types: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 					m := orderedmap.New[string, *BaseShape](0)
// 					return m
// 				}(),
// 				AnnotationTypes:        orderedmap.New[string, *BaseShape](0),
// 				Uses:                   orderedmap.New[string, *LibraryLink](0),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "identifier",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative case: uses fragment not found",
// 			fields: fields{
// 				ID:              "id",
// 				Usage:           "usage",
// 				Types:           orderedmap.New[string, *BaseShape](0),
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					return uses
// 				}(),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "positive case: uses identifier not found",
// 			fields: fields{
// 				ID:              "id",
// 				Usage:           "usage",
// 				Types:           orderedmap.New[string, *BaseShape](0),
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					uses.Set("fragment", &LibraryLink{
// 						Link: &Library{
// 							Types: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 								types := orderedmap.New[string, *BaseShape](0)
// 								return types
// 							}(),
// 						},
// 					})
// 					return uses
// 				}(),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := &Library{
// 				ID:                     tt.fields.ID,
// 				Usage:                  tt.fields.Usage,
// 				AnnotationTypes:        tt.fields.AnnotationTypes,
// 				Types:                  tt.fields.Types,
// 				Uses:                   tt.fields.Uses,
// 				CustomDomainProperties: tt.fields.CustomDomainProperties,
// 				Location:               tt.fields.Location,
// 				raml:                   tt.fields.raml,
// 			}
// 			got, err := l.GetReferenceType(tt.args.refName)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetReferenceType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(got)
// 			}
// 		})
// 	}
// }

// func TestLibrary_GetReferenceAnnotationType(t *testing.T) {
// 	type fields struct {
// 		ID                     string
// 		Usage                  string
// 		AnnotationTypes        *orderedmap.OrderedMap[string, *BaseShape]
// 		Types                  *orderedmap.OrderedMap[string, *BaseShape]
// 		Uses                   *orderedmap.OrderedMap[string, *LibraryLink]
// 		CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]
// 		Location               string
// 		raml                   *RAML
// 	}
// 	type args struct {
// 		refName string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive case: local ref",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Types: orderedmap.New[string, *BaseShape](0),
// 				AnnotationTypes: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 					m := orderedmap.New[string, *BaseShape](0)
// 					m.Set("identifier", &BaseShape{})
// 					return m
// 				}(),
// 				Uses:                   orderedmap.New[string, *LibraryLink](0),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "identifier",
// 			},
// 			want: func(got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("GetReferenceAnnotationType() got = %v, want %v", got, "not nil")
// 				}
// 			},
// 		},
// 		{
// 			name: "positive case: uses ref",
// 			fields: fields{
// 				ID:              "id",
// 				Usage:           "usage",
// 				Types:           orderedmap.New[string, *BaseShape](0),
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					uses.Set("fragment", &LibraryLink{
// 						Link: &Library{
// 							AnnotationTypes: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 								types := orderedmap.New[string, *BaseShape](0)
// 								types.Set("identifier", &BaseShape{})
// 								return types
// 							}(),
// 						},
// 					})
// 					return uses
// 				}(),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			want: func(got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("GetReferenceAnnotationType() got = %v, want %v", got, "not nil")
// 				}
// 			},
// 		},
// 		{
// 			name: "negative case: local ref not found",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Types: orderedmap.New[string, *BaseShape](0),
// 				AnnotationTypes: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 					m := orderedmap.New[string, *BaseShape](0)
// 					return m
// 				}(),
// 				Uses:                   orderedmap.New[string, *LibraryLink](0),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "identifier",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative case: uses fragment not found",
// 			fields: fields{
// 				ID:              "id",
// 				Usage:           "usage",
// 				Types:           orderedmap.New[string, *BaseShape](0),
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					return uses
// 				}(),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "positive case: uses identifier not found",
// 			fields: fields{
// 				ID:              "id",
// 				Usage:           "usage",
// 				Types:           orderedmap.New[string, *BaseShape](0),
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					uses.Set("fragment", &LibraryLink{
// 						Link: &Library{
// 							AnnotationTypes: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 								types := orderedmap.New[string, *BaseShape](0)
// 								return types
// 							}(),
// 						},
// 					})
// 					return uses
// 				}(),
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				Location:               "location",
// 				raml:                   &RAML{},
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := &Library{
// 				ID:                     tt.fields.ID,
// 				Usage:                  tt.fields.Usage,
// 				AnnotationTypes:        tt.fields.AnnotationTypes,
// 				Types:                  tt.fields.Types,
// 				Uses:                   tt.fields.Uses,
// 				CustomDomainProperties: tt.fields.CustomDomainProperties,
// 				Location:               tt.fields.Location,
// 				raml:                   tt.fields.raml,
// 			}
// 			got, err := l.GetReferenceAnnotationType(tt.args.refName)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetReferenceAnnotationType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(got)
// 			}
// 		})
// 	}
// }

// func TestLibrary_GetLocation(t *testing.T) {
// 	type fields struct {
// 		Location string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   string
// 	}{
// 		{
// 			name: "positive case",
// 			fields: fields{
// 				Location: "location",
// 			},
// 			want: "location",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := &Library{
// 				Location: tt.fields.Location,
// 			}
// 			if got := l.GetLocation(); got != tt.want {
// 				t.Errorf("GetLocation() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestLibrary_unmarshalUses(t *testing.T) {
// 	type fields struct {
// 		Uses *orderedmap.OrderedMap[string, *LibraryLink]
// 	}
// 	type args struct {
// 		valueNode *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(got *Library)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive case: tag null",
// 			fields: fields{
// 				Uses: orderedmap.New[string, *LibraryLink](0),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Tag:  "!!null",
// 				},
// 			},
// 		},
// 		{
// 			name: "negative case: invalid node kind",
// 			fields: fields{
// 				Uses: orderedmap.New[string, *LibraryLink](0),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.SequenceNode,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "positive case",
// 			fields: fields{
// 				Uses: orderedmap.New[string, *LibraryLink](0),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "fragment",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "identifier",
// 						},
// 					},
// 				},
// 			},
// 			want: func(got *Library) {
// 				if got.Uses == nil {
// 					t.Errorf("unmarshalUses() got = %v, want %v", got.Uses, "not nil")
// 				}
// 				frag, present := got.Uses.Get("fragment")
// 				if !present {
// 					t.Errorf("unmarshalUses() 'fragment' must be present")
// 				}
// 				if frag.Value != "identifier" {
// 					t.Errorf("unmarshalUses() 'fragment' value must be 'identifier'")
// 				}
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := &Library{
// 				Uses: tt.fields.Uses,
// 			}
// 			if err := l.unmarshalUses(tt.args.valueNode); (err != nil) != tt.wantErr {
// 				t.Errorf("unmarshalUses() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if tt.want != nil {
// 				tt.want(l)
// 			}
// 		})
// 	}
// }

// func TestLibrary_unmarshalTypes(t *testing.T) {
// 	type fields struct {
// 		Types *orderedmap.OrderedMap[string, *BaseShape]
// 		raml  *RAML
// 	}
// 	type args struct {
// 		valueNode *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(got *Library)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive case: tag null",
// 			fields: fields{
// 				Types: orderedmap.New[string, *BaseShape](0),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Tag:  "!!null",
// 				},
// 			},
// 		},
// 		{
// 			name: "positive case",
// 			fields: fields{
// 				Types: orderedmap.New[string, *BaseShape](0),
// 				raml: &RAML{
// 					fragmentTypes: make(map[string]map[string]*BaseShape),
// 				},
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "identifier",
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
// 			want: func(got *Library) {
// 				if got.Types == nil {
// 					t.Errorf("unmarshalTypes() got = %v, want %v", got.Types, "not nil")
// 				}
// 				shape, present := got.Types.Get("identifier")
// 				if !present {
// 					t.Errorf("unmarshalTypes() 'identifier' must be present")
// 				}
// 				if shape.Type != "string" {
// 					t.Errorf("unmarshalTypes() 'identifier' type must be 'string'")
// 				}
// 			},
// 		},
// 		{
// 			name: "negative case: invalid node kind",
// 			fields: fields{
// 				Types: orderedmap.New[string, *BaseShape](0),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.SequenceNode,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative case: invalid type tag",
// 			fields: fields{
// 				Types: orderedmap.New[string, *BaseShape](0),
// 				raml: &RAML{
// 					fragmentTypes: make(map[string]map[string]*BaseShape),
// 				},
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "identifier",
// 							Tag:   "!!int",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "string",
// 							Tag:   "!!int",
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := &Library{
// 				Types: tt.fields.Types,
// 				raml:  tt.fields.raml,
// 			}
// 			err := l.unmarshalTypes(tt.args.valueNode)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("unmarshalTypes() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(l)
// 			}
// 		})
// 	}
// }

// func TestLibrary_unmarshalAnnotationTypes(t *testing.T) {
// 	type fields struct {
// 		AnnotationTypes *orderedmap.OrderedMap[string, *BaseShape]
// 		raml            *RAML
// 	}
// 	type args struct {
// 		valueNode *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(got *Library)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive case: tag null",
// 			fields: fields{
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Tag:  "!!null",
// 				},
// 			},
// 		},
// 		{
// 			name: "positive case",
// 			fields: fields{
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 				raml: &RAML{
// 					fragmentAnnotationTypes: make(map[string]map[string]*BaseShape),
// 				},
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "identifier",
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
// 			want: func(got *Library) {
// 				if got.AnnotationTypes == nil {
// 					t.Errorf("unmarshalAnnotationTypes() got = %v, want %v", got.Types, "not nil")
// 				}
// 				shape, present := got.AnnotationTypes.Get("identifier")
// 				if !present {
// 					t.Errorf("unmarshalAnnotationTypes() 'identifier' must be present")
// 				}
// 				if shape.Type != "string" {
// 					t.Errorf("unmarshalAnnotationTypes() 'identifier' type must be 'string'")
// 				}
// 			},
// 		},
// 		{
// 			name: "negative case: invalid node kind",
// 			fields: fields{
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.SequenceNode,
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative case: invalid type tag",
// 			fields: fields{
// 				AnnotationTypes: orderedmap.New[string, *BaseShape](0),
// 				raml: &RAML{
// 					fragmentAnnotationTypes: make(map[string]map[string]*BaseShape),
// 				},
// 			},
// 			args: args{
// 				valueNode: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "identifier",
// 							Tag:   "!!int",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "string",
// 							Tag:   "!!int",
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := &Library{
// 				AnnotationTypes: tt.fields.AnnotationTypes,
// 				raml:            tt.fields.raml,
// 			}
// 			err := l.unmarshalAnnotationTypes(tt.args.valueNode)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("unmarshalAnnotationTypes() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(l)
// 			}
// 		})
// 	}
// }

// func TestLibrary_UnmarshalYAML(t *testing.T) {
// 	type fields struct {
// 		ID                     string
// 		Usage                  string
// 		AnnotationTypes        *orderedmap.OrderedMap[string, *BaseShape]
// 		Types                  *orderedmap.OrderedMap[string, *BaseShape]
// 		Uses                   *orderedmap.OrderedMap[string, *LibraryLink]
// 		CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]
// 		Location               string
// 		raml                   *RAML
// 	}
// 	type args struct {
// 		value *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 		want    func(got *Library)
// 	}{
// 		{
// 			name: "positive case: all fields",
// 			fields: fields{
// 				CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
// 				raml: &RAML{
// 					domainExtensions: make([]*DomainExtension, 0),
// 				},
// 			},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "uses",
// 						},
// 						{
// 							Kind: yaml.MappingNode,
// 							Tag:  "!!null",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "types",
// 						},
// 						{
// 							Kind: yaml.MappingNode,
// 							Tag:  "!!null",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "annotationTypes",
// 						},
// 						{
// 							Kind: yaml.MappingNode,
// 							Tag:  "!!null",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "usage",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "usage",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "(custom)",
// 						},
// 						{
// 							Kind: yaml.MappingNode,
// 							Tag:  "!!null",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name:   "negative case: must be map error",
// 			fields: fields{},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name:   "negative case: invalid types",
// 			fields: fields{},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "types",
// 						},
// 						{
// 							Kind: yaml.SequenceNode,
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name:   "negative case: invalid uses",
// 			fields: fields{},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "uses",
// 						},
// 						{
// 							Kind: yaml.SequenceNode,
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name:   "negative case: invalid annotationTypes",
// 			fields: fields{},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "annotationTypes",
// 						},
// 						{
// 							Kind: yaml.SequenceNode,
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name:   "negative case: invalid usage",
// 			fields: fields{},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "usage",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "string",
// 							Tag:   "!!int",
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name:   "negative case: invalid custom domain properties: (empty)",
// 			fields: fields{},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "()",
// 						},
// 						{
// 							Kind: yaml.SequenceNode,
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := &Library{
// 				ID:                     tt.fields.ID,
// 				Usage:                  tt.fields.Usage,
// 				AnnotationTypes:        tt.fields.AnnotationTypes,
// 				Types:                  tt.fields.Types,
// 				Uses:                   tt.fields.Uses,
// 				CustomDomainProperties: tt.fields.CustomDomainProperties,
// 				Location:               tt.fields.Location,
// 				raml:                   tt.fields.raml,
// 			}
// 			if err := l.UnmarshalYAML(tt.args.value); (err != nil) != tt.wantErr {
// 				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if tt.want != nil {
// 				tt.want(l)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeLibrary(t *testing.T) {
// 	type fields struct{}
// 	type args struct {
// 		path string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   func(got *Library)
// 	}{
// 		{
// 			name:   "positive case",
// 			fields: fields{},
// 			args: args{
// 				path: "testdata/fragment.raml",
// 			},
// 			want: func(got *Library) {
// 				if got.Location != "testdata/fragment.raml" {
// 					t.Errorf("MakeLibrary() got = %v, want %v", got.Location, "testdata/fragment.raml")
// 				}
// 				if got.Uses == nil {
// 					t.Errorf("MakeLibrary() got = %v, want %v", got.Uses, "not nil")
// 				}
// 				if got.Types == nil {
// 					t.Errorf("MakeLibrary() got = %v, want %v", got.Types, "not nil")
// 				}
// 				if got.AnnotationTypes == nil {
// 					t.Errorf("MakeLibrary() got = %v, want %v", got.AnnotationTypes, "not nil")
// 				}
// 				if got.CustomDomainProperties == nil {
// 					t.Errorf("MakeLibrary() got = %v, want %v", got.CustomDomainProperties, "not nil")
// 				}
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{}
// 			got := r.MakeLibrary(tt.args.path)
// 			if tt.want != nil {
// 				tt.want(got)
// 			}
// 		})
// 	}
// }

// func TestDataType_GetReferenceType(t *testing.T) {
// 	type fields struct {
// 		ID       string
// 		Usage    string
// 		Uses     *orderedmap.OrderedMap[string, *LibraryLink]
// 		Shape    *BaseShape
// 		Location string
// 		raml     *RAML
// 	}
// 	type args struct {
// 		refName string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "negative case: local ref",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Uses:  orderedmap.New[string, *LibraryLink](0),
// 				Shape: &BaseShape{
// 					Type: "string",
// 				},
// 				Location: "location",
// 				raml:     &RAML{},
// 			},
// 			args: args{
// 				refName: "identifier",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "positive case: uses ref",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					uses.Set("fragment", &LibraryLink{
// 						Link: &Library{
// 							Types: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 								types := orderedmap.New[string, *BaseShape](0)
// 								types.Set("identifier", &BaseShape{
// 									Type: "string",
// 								})
// 								return types
// 							}(),
// 						},
// 					})
// 					return uses
// 				}(),
// 				Shape: &BaseShape{
// 					Type: "string",
// 				},
// 				Location: "location",
// 				raml:     &RAML{},
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			want: func(got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("GetReferenceType() got = %v, want %v", got, "not nil")
// 				}
// 			},
// 		},
// 		{
// 			name: "negative case: uses fragment not found",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					return uses
// 				}(),
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "positive case: uses identifier not found",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					uses.Set("fragment", &LibraryLink{
// 						Link: &Library{
// 							Types: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 								types := orderedmap.New[string, *BaseShape](0)
// 								return types
// 							}(),
// 						},
// 					})
// 					return uses
// 				}(),
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			dt := &DataTypeFragment{
// 				ID:       tt.fields.ID,
// 				Uses:     tt.fields.Uses,
// 				Shape:    tt.fields.Shape,
// 				Location: tt.fields.Location,
// 				raml:     tt.fields.raml,
// 			}
// 			got, err := dt.GetReferenceType(tt.args.refName)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetReferenceType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(got)
// 			}
// 		})
// 	}
// }

// func TestDataType_GetReferenceAnnotationType(t *testing.T) {
// 	type fields struct {
// 		ID       string
// 		Usage    string
// 		Uses     *orderedmap.OrderedMap[string, *LibraryLink]
// 		Shape    *BaseShape
// 		Location string
// 		raml     *RAML
// 	}
// 	type args struct {
// 		refName string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(got *BaseShape)
// 		wantErr bool
// 	}{
// 		{
// 			name: "negative case: local ref",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Uses:  orderedmap.New[string, *LibraryLink](0),
// 				Shape: &BaseShape{
// 					Type: "string",
// 				},
// 				Location: "location",
// 				raml:     &RAML{},
// 			},
// 			args: args{
// 				refName: "identifier",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "positive case: uses ref",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					uses.Set("fragment", &LibraryLink{
// 						Link: &Library{
// 							AnnotationTypes: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 								types := orderedmap.New[string, *BaseShape](0)
// 								types.Set("identifier", &BaseShape{
// 									Type: "string",
// 								})
// 								return types
// 							}(),
// 						},
// 					})
// 					return uses
// 				}(),
// 				Shape: &BaseShape{
// 					Type: "string",
// 				},
// 				Location: "location",
// 				raml:     &RAML{},
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			want: func(got *BaseShape) {
// 				if got == nil {
// 					t.Errorf("GetReferenceAnnotationType() got = %v, want %v", got, "not nil")
// 				}
// 			},
// 		},
// 		{
// 			name: "negative case: uses fragment not found",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					return uses
// 				}(),
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "positive case: uses identifier not found",
// 			fields: fields{
// 				ID:    "id",
// 				Usage: "usage",
// 				Uses: func() *orderedmap.OrderedMap[string, *LibraryLink] {
// 					uses := orderedmap.New[string, *LibraryLink](0)
// 					uses.Set("fragment", &LibraryLink{
// 						Link: &Library{
// 							AnnotationTypes: func() *orderedmap.OrderedMap[string, *BaseShape] {
// 								types := orderedmap.New[string, *BaseShape](0)
// 								return types
// 							}(),
// 						},
// 					})
// 					return uses
// 				}(),
// 			},
// 			args: args{
// 				refName: "fragment.identifier",
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			dt := &DataTypeFragment{
// 				ID:       tt.fields.ID,
// 				Uses:     tt.fields.Uses,
// 				Shape:    tt.fields.Shape,
// 				Location: tt.fields.Location,
// 				raml:     tt.fields.raml,
// 			}
// 			got, err := dt.GetReferenceAnnotationType(tt.args.refName)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetReferenceAnnotationType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(got)
// 			}
// 		})
// 	}
// }

// func TestDataType_GetLocation(t *testing.T) {
// 	type fields struct {
// 		Location string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   string
// 	}{
// 		{
// 			name: "positive case",
// 			fields: fields{
// 				Location: "location",
// 			},
// 			want: "location",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			dt := &DataTypeFragment{
// 				Location: tt.fields.Location,
// 			}
// 			if got := dt.GetLocation(); got != tt.want {
// 				t.Errorf("GetLocation() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestDataType_UnmarshalYAML(t *testing.T) {
// 	type fields struct {
// 		ID       string
// 		Usage    string
// 		Uses     *orderedmap.OrderedMap[string, *LibraryLink]
// 		Shape    *BaseShape
// 		Location string
// 		raml     *RAML
// 	}
// 	type args struct {
// 		value *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive case: all fields",
// 			fields: fields{
// 				Uses: orderedmap.New[string, *LibraryLink](0),
// 				raml: &RAML{
// 					shapes: []*BaseShape{},
// 				},
// 			},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "uses",
// 						},
// 						{
// 							Kind: yaml.MappingNode,
// 							Tag:  "!!null",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "uses",
// 						},
// 						{
// 							Kind: yaml.MappingNode,
// 							Content: []*yaml.Node{
// 								{
// 									Kind:  yaml.ScalarNode,
// 									Value: "fragment",
// 								},
// 								{
// 									Kind:  yaml.ScalarNode,
// 									Value: "identifier",
// 								},
// 							},
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "usage",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "usage",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "(custom)",
// 						},
// 						{
// 							Kind: yaml.MappingNode,
// 							Tag:  "!!null",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name:   "negative case: must be map error",
// 			fields: fields{},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative case: parse types: make shape",
// 			fields: fields{
// 				raml: &RAML{
// 					shapes: []*BaseShape{},
// 				},
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
// 							Value: "123",
// 							Tag:   "!!int",
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			dt := &DataTypeFragment{
// 				ID:       tt.fields.ID,
// 				Uses:     tt.fields.Uses,
// 				Shape:    tt.fields.Shape,
// 				Location: tt.fields.Location,
// 				raml:     tt.fields.raml,
// 			}
// 			if err := dt.UnmarshalYAML(tt.args.value); (err != nil) != tt.wantErr {
// 				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeDataType(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		path string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   func(got *DataTypeFragment)
// 	}{
// 		{
// 			name:   "positive case",
// 			fields: fields{},
// 			args: args{
// 				path: "testdata/fragment.raml",
// 			},
// 			want: func(got *DataTypeFragment) {
// 				if got.Location != "testdata/fragment.raml" {
// 					t.Errorf("MakeDataType() got = %v, want %v", got.Location, "testdata/fragment.raml")
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
// 				ctx:                     tt.fields.ctx,
// 			}
// 			got := r.MakeDataTypeFragment(tt.args.path)
// 			if tt.want != nil {
// 				tt.want(got)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeJSONDataType(t *testing.T) {
// 	type fields struct {
// 		fragmentsCache          map[string]Fragment
// 		fragmentTypes           map[string]map[string]*BaseShape
// 		fragmentAnnotationTypes map[string]map[string]*BaseShape
// 		entryPoint              Fragment
// 		domainExtensions        []*DomainExtension
// 		shapes                  []*BaseShape
// 		unresolvedShapes        list.List
// 		ctx                     context.Context
// 	}
// 	type args struct {
// 		value []byte
// 		path  string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    func(got *DataTypeFragment)
// 		wantErr bool
// 	}{
// 		{
// 			name: "positive case",
// 			fields: fields{
// 				fragmentsCache:          make(map[string]Fragment),
// 				fragmentTypes:           make(map[string]map[string]*BaseShape),
// 				fragmentAnnotationTypes: make(map[string]map[string]*BaseShape),
// 				domainExtensions:        []*DomainExtension{},
// 				shapes:                  []*BaseShape{},
// 				unresolvedShapes:        list.List{},
// 				ctx:                     context.Background(),
// 			},
// 			args: args{
// 				value: []byte(`{"type": "string"}`),
// 				path:  "testdata/fragment.json",
// 			},
// 			want: func(got *DataTypeFragment) {
// 				if got.Location != "testdata/fragment.json" {
// 					t.Errorf("MakeJSONDataType() got = %v, want %v", got.Location, "testdata/fragment.json")
// 				}
// 				if got.Shape == nil {
// 					t.Errorf("MakeJSONDataType() got = %v, want %v", got.Shape, "not nil")
// 				}
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "negative case: decode fragment error",
// 			fields: fields{
// 				fragmentsCache: make(map[string]Fragment),
// 			},
// 			args: args{
// 				value: []byte(`{`),
// 				path:  "testdata/fragment.json",
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
// 				ctx:                     tt.fields.ctx,
// 			}
// 			got, err := r.MakeJSONDataType(tt.args.value, tt.args.path)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("MakeJSONDataType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != nil {
// 				tt.want(got)
// 			}
// 		})
// 	}
// }

// func TestNamedExample_GetReferenceAnnotationType(t *testing.T) {
// 	type fields struct {
// 		ID       string
// 		Map      *orderedmap.OrderedMap[string, *Example]
// 		Location string
// 		raml     *RAML
// 	}
// 	type args struct {
// 		in0 string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    *BaseShape
// 		wantErr bool
// 	}{
// 		{
// 			name:   "negative case",
// 			fields: fields{},
// 			args: args{
// 				in0: "identifier",
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ne := &NamedExample{
// 				ID:       tt.fields.ID,
// 				Map:      tt.fields.Map,
// 				Location: tt.fields.Location,
// 				raml:     tt.fields.raml,
// 			}
// 			got, err := ne.GetReferenceAnnotationType(tt.args.in0)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetReferenceAnnotationType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetReferenceAnnotationType() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestNamedExample_GetReferenceType(t *testing.T) {
// 	type fields struct {
// 		ID       string
// 		Map      *orderedmap.OrderedMap[string, *Example]
// 		Location string
// 		raml     *RAML
// 	}
// 	type args struct {
// 		in0 string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    *BaseShape
// 		wantErr bool
// 	}{
// 		{
// 			name:   "negative case",
// 			fields: fields{},
// 			args: args{
// 				in0: "identifier",
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ne := &NamedExample{
// 				ID:       tt.fields.ID,
// 				Map:      tt.fields.Map,
// 				Location: tt.fields.Location,
// 				raml:     tt.fields.raml,
// 			}
// 			got, err := ne.GetReferenceType(tt.args.in0)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetReferenceType() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetReferenceType() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestNamedExample_GetLocation(t *testing.T) {
// 	type fields struct {
// 		ID       string
// 		Map      *orderedmap.OrderedMap[string, *Example]
// 		Location string
// 		raml     *RAML
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		want   string
// 	}{
// 		{
// 			name: "positive case",
// 			fields: fields{
// 				Location: "location",
// 			},
// 			want: "location",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ne := &NamedExample{
// 				ID:       tt.fields.ID,
// 				Map:      tt.fields.Map,
// 				Location: tt.fields.Location,
// 				raml:     tt.fields.raml,
// 			}
// 			if got := ne.GetLocation(); got != tt.want {
// 				t.Errorf("GetLocation() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestRAML_MakeNamedExample(t *testing.T) {
// 	type fields struct {
// 	}
// 	type args struct {
// 		path string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   func(tt *testing.T, got *NamedExample)
// 	}{
// 		{
// 			name:   "positive case",
// 			fields: fields{},
// 			args: args{
// 				path: "testdata/fragment.raml",
// 			},
// 			want: func(tt *testing.T, got *NamedExample) {
// 				if got.Location != "testdata/fragment.raml" {
// 					tt.Errorf("MakeNamedExample() got = %v, want %v", got.Location, "testdata/fragment.raml")
// 				}
// 				if got.Map != nil {
// 					tt.Errorf("MakeNamedExample() got = %v, want %v", got.Map, "nil")
// 				}
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := &RAML{}
// 			got := r.MakeNamedExample(tt.args.path)
// 			if tt.want != nil {
// 				tt.want(t, got)
// 			}
// 		})
// 	}
// }

// func TestNamedExample_UnmarshalYAML(t *testing.T) {
// 	type fields struct {
// 		ID       string
// 		Map      *orderedmap.OrderedMap[string, *Example]
// 		Location string
// 		raml     *RAML
// 	}
// 	type args struct {
// 		value *yaml.Node
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 		want    func(tt *testing.T, got *NamedExample)
// 	}{
// 		{
// 			name: "positive case: all fields",
// 			fields: fields{
// 				Map:  orderedmap.New[string, *Example](0),
// 				raml: &RAML{},
// 			},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "name",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "value",
// 						},
// 					},
// 				},
// 			},
// 			want: func(tt *testing.T, got *NamedExample) {
// 				if got.Map == nil {
// 					tt.Errorf("UnmarshalYAML() got = %v, want %v", got.Map, "not nil")
// 				}
// 				example, present := got.Map.Get("name")
// 				if !present {
// 					tt.Errorf("UnmarshalYAML() 'name' must be present")
// 				}
// 				if example.Data.Value.(string) != "value" {
// 					tt.Errorf("UnmarshalYAML() 'name' value must be 'value'")
// 				}
// 			},
// 		},
// 		{
// 			name:   "negative case: must be map error",
// 			fields: fields{},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind:  yaml.ScalarNode,
// 					Value: "value",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "negative case: invalid example data",
// 			fields: fields{
// 				raml: &RAML{},
// 			},
// 			args: args{
// 				value: &yaml.Node{
// 					Kind: yaml.MappingNode,
// 					Content: []*yaml.Node{
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "name",
// 						},
// 						{
// 							Kind:  yaml.ScalarNode,
// 							Value: "value",
// 							Tag:   "!!int",
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ne := &NamedExample{
// 				ID:       tt.fields.ID,
// 				Map:      tt.fields.Map,
// 				Location: tt.fields.Location,
// 				raml:     tt.fields.raml,
// 			}
// 			if err := ne.UnmarshalYAML(tt.args.value); (err != nil) != tt.wantErr {
// 				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if tt.want != nil {
// 				tt.want(t, ne)
// 			}
// 		})
// 	}
// }
