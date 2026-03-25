package raml

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func writeToDiskHelper(t *testing.T, tt struct {
	name string
	typeName string
	outSchema *JSONSchemaRAML
}) {
	outputDir := "./test_output"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		t.Errorf("Failed to create output directory: %v", err)
	}
	outputPath := filepath.Join(outputDir, tt.name+"_"+tt.typeName+".json")
	var jsonData []byte
	jsonData, err := json.MarshalIndent(tt.outSchema, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal JSON Schema: %v", err)
	}
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		t.Errorf("Failed to write JSON Schema to file: %v", err)
	}
	t.Logf("Successfully wrote JSON Schema to %s", outputPath)
}


func Test_ParseFixturesIntegration(t *testing.T) {
	// Define test cases for valid fixtures that should parse successfully
	validTests := []struct {
		name string
		path string
	}{
		{
			name: "library_recursive_trait.raml",
			path: "./fixtures/library_recursive_trait.raml",
		},
		{
			name: "library.raml",
			path: "./fixtures/library.raml",
		},
		{
			name: "dtype.raml",
			path: "./fixtures/dtype.raml",
		},
		{
			name: "named_example.raml",
			path: "./fixtures/named_example.raml",
		},
		{
			name: "other_lib.raml",
			path: "./fixtures/other_lib.raml",
		},
		{
			name: "recursive_type.raml",
			path: "./fixtures/recursive_type.raml",
		},
		{
			name: "common.raml",
			path: "./fixtures/common.raml",
		},
	}

	for _, tt := range validTests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			rml, err := ParseFromPath(tt.path, OptWithUnwrap(), OptWithValidate())
			require.NoError(t, err, "Failed to parse %s", tt.path)
			elapsed := time.Since(start)

			require.NotNil(t, rml, "RAML should not be nil for %s", tt.path)
			require.NotNil(t, rml.entryPoint, "Entry point should not be nil for %s", tt.path)

			shapesAll := rml.GetShapes()
			t.Logf("ParseFromPath(%s) took %d ms, location %s, total shapes %d",
				tt.name, elapsed.Milliseconds(), rml.entryPoint.GetLocation(), len(shapesAll))

			// Verify that no UnknownShape instances exist
			for _, base := range shapesAll {
				shape := base.Shape
				require.NotNil(t, shape, "Shape should not be nil in %s", tt.path)
				_, ok := shape.(*UnknownShape)
				require.False(t, ok, "Found UnknownShape in %s, this indicates parsing issues", tt.path)
			}

			// Verify JSON Schema conversion
			conv, err := NewJSONSchemaConverter(WithWrapper(JSONSchemaWrapper))
			require.NoError(t, err, "Failed to create JSON Schema converter for %s", tt.path)

			convertedCount := 0
			for _, frag := range rml.fragmentsCache {
				switch f := frag.(type) {
				case *Library:
					for pair := f.AnnotationTypes.Oldest(); pair != nil; pair = pair.Next() {
						s := pair.Value
						typeName := pair.Key
						outSchema, err := conv.Convert(s.Shape)
						writeToDiskHelper(t, struct{name string; typeName string; outSchema *JSONSchemaRAML}{name: tt.name, typeName: typeName, outSchema: outSchema})
						require.NoError(t, err, "Failed to convert annotation type shape in %s: %v", tt.path, err)
						convertedCount++
					}
					for pair := f.Types.Oldest(); pair != nil; pair = pair.Next() {
						s := pair.Value
						typeName := pair.Key
						outSchema, err := conv.Convert(s.Shape)
						writeToDiskHelper(t, struct{name string; typeName string; outSchema *JSONSchemaRAML}{name: tt.name, typeName: typeName, outSchema: outSchema})
						require.NoError(t, err, "Failed to convert type shape in %s: %v", tt.path, err)
						convertedCount++
					}
				case *DataType:
					typeName := f.Shape.Name
					outSchema, err := conv.Convert(f.Shape.Shape)
					writeToDiskHelper(t, struct{name string; typeName string; outSchema *JSONSchemaRAML}{name: tt.name, typeName: typeName, outSchema: outSchema})
					require.NoError(t, err, "Failed to convert data type shape in %s: %v", tt.path, err)
					convertedCount++
				}
			}
			t.Logf("Successfully converted %d shapes to JSON Schema for %s", convertedCount, tt.name)
		})
	}

	// Define test cases for invalid fixtures that should fail parsing
	invalidTests := []struct {
		name string
		path string
	}{
		{
			name: "dtype_invalid_decode.raml",
			path: "./fixtures/dtype_invalid_decode.raml",
		},
		{
			name: "dtype_invalid_header.raml",
			path: "./fixtures/dtype_invalid_header.raml",
		},
		{
			name: "invalid_decode.raml",
			path: "./fixtures/invalid_decode.raml",
		},
		{
			name: "library_invalid.raml",
			path: "./fixtures/library_invalid.raml",
		},
		{
			name: "library_invalid_decode.raml",
			path: "./fixtures/library_invalid_decode.raml",
		},
		{
			name: "library_invalid_dot_import.raml",
			path: "./fixtures/library_invalid_dot_import.raml",
		},
		{
			name: "library_invalid_inheritance.raml",
			path: "./fixtures/library_invalid_inheritance.raml",
		},
		{
			name: "library_invalid_unwrap.raml",
			path: "./fixtures/library_invalid_unwrap.raml",
		},
		{
			name: "library_invalid_recursive_trait.raml",
			path: "./fixtures/library_invalid_inherited_recursive_trait.raml",
		},
		{
			name: "library_invalid_trait_example.raml",
			path: "./fixtures/library_invalid_trait_example.raml",
		},
	}

	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFromPath(tt.path, OptWithUnwrap(), OptWithValidate())
			require.Error(t, err, "Expected error when parsing %s", tt.path)
			t.Logf("ParseFromPath(%s) correctly failed with error: %v", tt.name, err)
		})
	}
}

func printMemUsage(t *testing.T) {
	var m runtime.MemStats
	t.Helper()
	runtime.ReadMemStats(&m)
	t.Logf("Memory usage: alloc MiB %d, total alloc MiB %d, sys MiB %d, num GC %d",
		m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

type mockReadSeeker struct {
	P       []byte
	ReadErr error
	SeekErr error
}

func (m *mockReadSeeker) Read(p []byte) (n int, err error) {
	if m.ReadErr != nil {
		return 0, m.ReadErr
	}
	if len(m.P) == 0 {
		return 0, io.EOF
	}
	n = copy(p, m.P)
	return n, io.EOF
}

func (m *mockReadSeeker) Seek(offset int64, whence int) (int64, error) {
	if m.SeekErr != nil {
		return 0, m.SeekErr
	}
	return 0, nil
}

func TestReadHead(t *testing.T) {
	type args struct {
		f io.ReadSeeker
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "positive: read head",
			args: args{
				f: &mockReadSeeker{
					P: []byte("#%RAML 1.0 Library\r\n"),
				},
			},
			want: "#%RAML 1.0 Library",
		},
		{
			name: "negative: read head",
			args: args{
				f: &mockReadSeeker{
					P:       []byte(""),
					ReadErr: errors.New("read error"),
				},
			},
			wantErr: true,
		},
		{
			name: "negative: seek error",
			args: args{
				f: &mockReadSeeker{
					P:       []byte("#%RAML 1.0 Library\n"),
					SeekErr: errors.New("seek error"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadHead(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadHead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadHead() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifyFragment(t *testing.T) {
	type args struct {
		head string
	}
	tests := []struct {
		name    string
		args    args
		want    FragmentKind
		wantErr bool
	}{
		{
			name: "positive: identify library",
			args: args{
				head: "#%RAML 1.0 Library",
			},
			want: FragmentLibrary,
		},
		{
			name: "positive: identify datatype",
			args: args{
				head: "#%RAML 1.0 DataType",
			},
			want: FragmentDataType,
		},
		{
			name: "positive: identify resource type",
			args: args{
				head: "#%RAML 1.0 ResourceType",
			},
			want:    FragmentUnknown,
			wantErr: true,
		},
		{
			name: "negative: identify named example",
			args: args{
				head: "#%RAML 1.0 NamedExample",
			},
			want: FragmentNamedExample,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IdentifyFragment(tt.args.head)
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentifyFragment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IdentifyFragment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadRawFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    func(closer io.ReadCloser)
		wantErr bool
	}{
		{
			name: "positive: read raw file",
			args: args{
				path: "./fixtures/library.raml",
			},
			want: func(closer io.ReadCloser) {
				if closer == nil {
					t.Errorf("ReadRawFile() got = %v, want %v", closer, nil)
				}
				err := closer.Close()
				if err != nil {
					t.Errorf("ReadRawFile() got = %v, want %v", err, nil)
				}
			},
		},
		{
			name: "negative: read raw file",
			args: args{
				path: "./fixtures/not-found.raml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadRawFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadRawFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(got)
			}
		})
	}
}

func TestRAML_decodeDataType(t *testing.T) {
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
		f    io.Reader
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(tt *testing.T, got *DataType)
		wantErr bool
	}{
		{
			name: "positive: decode data type",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:    &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\ntype: string")},
				path: "./fixtures/test.raml",
			},
			want: func(tt *testing.T, got *DataType) {
				require.NotNil(tt, got)
				require.Equal(tt, "string", got.Shape.Type)
			},
		},
		{
			name: "parse data type: json schema",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/test.json",
				f:    &mockReadSeeker{P: []byte("{\"type\": \"string\"}")},
			},
			want: func(tt *testing.T, got *DataType) {
				require.NotNil(tt, got)
				require.Equal(tt, "json", got.Shape.Type)
			},
		},
		{
			name: "positive: data type with uses",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
			},
			args: args{
				f:    &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\nuses:\n  common: common.raml\ntype: common.A")},
				path: "./fixtures/test.raml",
			},
			want: func(tt *testing.T, got *DataType) {
				require.NotNil(tt, got)
				require.NotNil(tt, got.Uses)
			},
		},
		{
			name: "negative: json file read",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:    &mockReadSeeker{ReadErr: errors.New("read error")},
				path: "./fixtures/test.json",
			},
			wantErr: true,
		},
		{
			name: "negative: make json data type error",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:    &mockReadSeeker{P: []byte("{invalid json")},
				path: "./fixtures/test.json",
			},
			wantErr: true,
		},
		{
			name: "negative: data type error (bad yaml)",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:    &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\ninvalid")},
				path: "./fixtures/test.raml",
			},
			wantErr: true,
		},
		{
			name: "negative: invalid uses",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:    &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\nuses:\n  common: invalid_decode.raml")},
				path: "./fixtures/test.raml",
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
			got, err := r.decodeDataType(tt.args.f, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeDataType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestCheckFragmentKind(t *testing.T) {
	type args struct {
		f    *os.File
		kind FragmentKind
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "positive: check fragment kind: library",
			args: args{
				f: func() *os.File {
					f, _ := os.Open("./fixtures/library.raml")
					return f
				}(),
				kind: FragmentLibrary,
			},
			wantErr: false,
		},
		{
			name: "positive: check fragment kind: data type",
			args: args{
				f: func() *os.File {
					f, _ := os.Open("./fixtures/dtype.raml")
					return f
				}(),
				kind: FragmentDataType,
			},
			wantErr: false,
		},
		{
			name: "positive: json data type",
			args: args{
				f: func() *os.File {
					f, _ := os.Open("./fixtures/dtype.json")
					return f
				}(),
				kind: FragmentDataType,
			},
			wantErr: false,
		},
		{
			name: "negative: check fragment kind",
			args: args{
				f: func() *os.File {
					f, _ := os.Open("./fixtures/library.raml")
					return f
				}(),
				kind: FragmentDataType,
			},
			wantErr: true,
		},
		{
			name: "negative: file not found",
			args: args{
				f: nil,
			},
			wantErr: true,
		},
		{
			name: "negative: invalid fragment kind",
			args: args{
				f: func() *os.File {
					f, _ := os.Open("./fixtures/invalid_decode.raml")
					return f
				}(),
				kind: FragmentLibrary,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckFragmentKind(tt.args.f, tt.args.kind); (err != nil) != tt.wantErr {
				t.Errorf("CheckFragmentKind() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_parseDataType(t *testing.T) {
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
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(tt *testing.T, got *DataType)
		wantErr bool
	}{
		{
			name: "positive: parse data type",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/dtype.raml",
			},
			want: func(tt *testing.T, got *DataType) {
				require.NotNil(tt, got)
				require.Equal(tt, "string", got.Shape.Type)
			},
		},
		{
			name: "positive: get fragment from cache",
			fields: fields{
				fragmentsCache: map[string]Fragment{
					"test": &DataType{
						Shape: &BaseShape{
							Type: "string",
						},
					},
				},
			},
			args: args{
				path: "test",
			},
			want: func(tt *testing.T, got *DataType) {
				require.NotNil(tt, got)
				require.Equal(tt, "string", got.Shape.Type)
			},
		},
		{
			name: "negative: invalid header",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/dtype_invalid_header.raml",
			},
			wantErr: true,
		},
		{
			name: "negative: file not found",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/not-found.raml",
			},
			wantErr: true,
		},
		{
			name: "negative: invalid data type",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/dtype_invalid_decode.raml",
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
			got, err := r.parseDataType(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDataType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_parseLibrary(t *testing.T) {
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
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(tt *testing.T, got *Library)
		wantErr bool
	}{
		{
			name: "positive: parse library",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
			},
			args: args{
				path: "./fixtures/library.raml",
			},
			want: func(tt *testing.T, got *Library) {
				require.NotNil(tt, got)
			},
		},
		{
			name: "negative: file not found",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/not-found.raml",
			},
			wantErr: true,
		},
		{
			name: "negative: invalid kind",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/dtype.raml",
			},
			wantErr: true,
		},
		{
			name: "negative: invalid decode library",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/library_invalid_decode.raml",
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
			got, err := r.parseLibrary(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLibrary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_decodeNamedExample(t *testing.T) {
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
		f    io.Reader
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(tt *testing.T, got *NamedExample)
		wantErr bool
	}{
		{
			name: "positive: decode named example",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:    &mockReadSeeker{P: []byte("#%RAML 1.0 NamedExample\nexample: {\"name\": \"John\"}")},
				path: "./fixtures/named_example.raml",
			},
			want: func(tt *testing.T, got *NamedExample) {
				require.NotNil(tt, got)
				require.NotNil(tt, got.Map)
				require.NotEmpty(tt, got.Map)
			},
		},
		{
			name: "negative: invalid named example",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:    &mockReadSeeker{P: []byte("#%RAML 1.0 NamedExample\ninvalid")},
				path: "./fixtures/named_example.raml",
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
			got, err := r.decodeNamedExample(tt.args.f, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeNamedExample() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_parseNamedExample(t *testing.T) {
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
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(tt *testing.T, got *NamedExample)
		wantErr bool
	}{
		{
			name: "positive: parse named example",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/named_example.raml",
			},
			want: func(tt *testing.T, got *NamedExample) {
				require.NotNil(tt, got)
				require.NotNil(tt, got.Map)
				require.NotEmpty(tt, got.Map)
			},
		},
		{
			name: "positive: get fragment from cache",
			fields: fields{
				fragmentsCache: map[string]Fragment{
					"test": &NamedExample{
						ID: "test",
					},
				},
			},
			args: args{
				path: "test",
			},
			want: func(tt *testing.T, got *NamedExample) {
				require.NotNil(tt, got)
				require.Equal(tt, "test", got.ID)
			},
		},
		{
			name: "negative: file not found",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/not-found.raml",
			},
			wantErr: true,
		},
		{
			name: "negative: invalid fragment kind",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/dtype.raml",
			},
			wantErr: true,
		},
		{
			name: "negative: invalid decode named example",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/named_example_invalid_decode.raml",
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
			got, err := r.parseNamedExample(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNamedExample() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestRAML_ParseFromPath(t *testing.T) {
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
		path string
		opts []ParseOpt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive: parse from path",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/named_example.raml",
				opts: []ParseOpt{OptWithUnwrap(), OptWithValidate()},
			},
		},
		{
			name: "negative: file not found",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				path: "./fixtures/not-found.raml",
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
			if err := r.ParseFromPath(tt.args.path, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("ParseFromPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_ParseFromString(t *testing.T) {
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
		content  string
		fileName string
		baseDir  string
		opts     []ParseOpt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "positive: parse from string",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				content:  "#%RAML 1.0 NamedExample\nexample: {\"name\": \"John\"}",
				fileName: "test.raml",
				baseDir:  "./fixtures",
				opts:     []ParseOpt{OptWithUnwrap(), OptWithValidate()},
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
			if err := r.ParseFromString(tt.args.content, tt.args.fileName, tt.args.baseDir, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("ParseFromString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRAML_parseFragment(t *testing.T) {
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
		f            io.ReadSeeker
		fragmentPath string
		pOpts        *parserOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    func(tt *testing.T, got *RAML)
	}{
		{
			name: "positive: parse fragment: named example",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:            &mockReadSeeker{P: []byte("#%RAML 1.0 NamedExample\nexample: {\"name\": \"John\"}")},
				fragmentPath: "./fixtures/named_example.raml",
				pOpts: &parserOptions{
					withUnwrapOpt:   true,
					withValidateOpt: true,
				},
			},
		},
		{
			name: "positive: parse fragment: data type",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:            &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\ntype: string")},
				fragmentPath: "./fixtures/dtype.raml",
				pOpts: &parserOptions{
					withUnwrapOpt:   true,
					withValidateOpt: true,
				},
			},
		},
		{
			name: "positive: parse fragment: library",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f:            &mockReadSeeker{P: []byte("#%RAML 1.0 Library\nuses:")},
				fragmentPath: "./fixtures/library.raml",
				pOpts: &parserOptions{
					withUnwrapOpt:   true,
					withValidateOpt: true,
				},
			},
		},
		{
			name: "negative: read file",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f: &mockReadSeeker{ReadErr: errors.New("read error")},
			},
			wantErr: true,
		},
		{
			name: "negative: identify fragment kind",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f: &mockReadSeeker{P: []byte("#%RAML 1.0 UnknownKind\n")},
			},
			wantErr: true,
		},
		{
			name: "negative: parse library error",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f: &mockReadSeeker{P: []byte("#%RAML 1.0 Library\ninvalid")},
			},
			wantErr: true,
		},
		{
			name: "negative: parse data type error",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f: &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\ninvalid")},
			},
			wantErr: true,
		},
		{
			name: "negative: parse named example error",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
			},
			args: args{
				f: &mockReadSeeker{P: []byte("#%RAML 1.0 NamedExample\ninvalid")},
			},
			wantErr: true,
		},
		{
			name: "negative: resolve type error",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
			},
			args: args{
				f:            &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\nuses:\n  lib: library_invalid_inheritance.raml\ntype: lib.C")},
				fragmentPath: "./fixtures/dtype.raml",
				pOpts:        &parserOptions{},
			},
			wantErr: true,
		},
		{
			name: "negative: resolve annotation type error",
			fields: fields{
				fragmentsCache:          map[string]Fragment{},
				fragmentTypes:           map[string]map[string]*BaseShape{},
				fragmentAnnotationTypes: map[string]map[string]*BaseShape{},
			},
			args: args{
				f:            &mockReadSeeker{P: []byte("#%RAML 1.0 Library\n(A): B")},
				fragmentPath: "./fixtures/dtype.raml",
				pOpts:        &parserOptions{},
			},
			wantErr: true,
		},
		{
			name: "negative: unwrap error",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
				fragmentTypes:  map[string]map[string]*BaseShape{},
			},
			args: args{
				f:            &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\nuses:\n  lib: library_invalid_unwrap.raml\ntype: lib.B")},
				pOpts:        &parserOptions{withUnwrapOpt: true},
				fragmentPath: "./fixtures/dtype.raml",
			},
			wantErr: true,
		},
		{
			name: "negative: validate error",
			fields: fields{
				fragmentsCache: map[string]Fragment{},
				fragmentTypes:  map[string]map[string]*BaseShape{},
			},
			args: args{
				f: &mockReadSeeker{P: []byte("#%RAML 1.0 DataType\nuses:\n  lib: library_invalid.raml\ntype: lib.A")},
				pOpts: &parserOptions{
					withUnwrapOpt:   true,
					withValidateOpt: true,
				},
				fragmentPath: "./fixtures/dtype.raml",
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
			if err := r.parseFragment(tt.args.f, tt.args.fragmentPath, tt.args.pOpts); (err != nil) != tt.wantErr {
				t.Errorf("parseFragment() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				tt.want(t, r)
			}
		})
	}
}

func TestParseFromPathCtx(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
		opts []ParseOpt
	}
	tests := []struct {
		name    string
		args    args
		want    func(tt *testing.T, got *RAML)
		wantErr bool
	}{
		{
			name: "positive: parse from path",
			args: args{
				ctx:  context.Background(),
				path: "./fixtures/named_example.raml",
			},
			want: func(tt *testing.T, got *RAML) {
				require.NotNil(tt, got)
			},
		},
		{
			name: "negative: ctx is nil",
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFromPathCtx(tt.args.ctx, tt.args.path, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFromPathCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestParseFromStringCtx(t *testing.T) {
	type args struct {
		ctx      context.Context
		content  string
		fileName string
		baseDir  string
		opts     []ParseOpt
	}
	tests := []struct {
		name    string
		args    args
		want    func(tt *testing.T, got *RAML)
		wantErr bool
	}{
		{
			name: "positive: parse from string",
			args: args{
				ctx:      context.Background(),
				content:  "#%RAML 1.0 NamedExample\nexample: {\"name\": \"John\"}",
				fileName: "test.raml",
				baseDir:  "./fixtures",
				opts:     []ParseOpt{OptWithUnwrap(), OptWithValidate()},
			},
			want: func(tt *testing.T, got *RAML) {
				require.NotNil(tt, got)
			},
		},
		{
			name: "negative: ctx is nil",
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFromStringCtx(tt.args.ctx, tt.args.content, tt.args.fileName, tt.args.baseDir, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFromStringCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestParseFromString(t *testing.T) {
	type args struct {
		content  string
		fileName string
		baseDir  string
		opts     []ParseOpt
	}
	tests := []struct {
		name    string
		args    args
		want    func(tt *testing.T, got *RAML)
		wantErr bool
	}{
		{
			name: "positive: parse from string",
			args: args{
				content:  "#%RAML 1.0 NamedExample\nexample: {\"name\": \"John\"}",
				fileName: "test.raml",
				baseDir:  mustAbs("./fixtures"),
				opts:     []ParseOpt{OptWithUnwrap(), OptWithValidate()},
			},
			want: func(tt *testing.T, got *RAML) {
				require.NotNil(tt, got)
			},
		},
		{
			name: "negative: baseDir must be an absolute path",
			args: args{
				content: "#%RAML 1.0 NamedExample\nexample: {\"name\": \"John\"}",
				baseDir: "fixtures",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFromString(tt.args.content, tt.args.fileName, tt.args.baseDir, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}
