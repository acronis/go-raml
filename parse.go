package raml

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReadHead reads, reset file and returns the trimmed first line of a file.
func ReadHead(f io.ReadSeeker) (string, error) {
	r := bufio.NewReader(f)
	head, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read fragment head: %w", err)
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("seek to start: %w", err)
	}

	head = strings.TrimRight(head, "\r\n ")
	return head, nil
}

// IdentifyFragment returns the kind of the fragment by its head.
func IdentifyFragment(head string) (FragmentKind, error) {
	switch head {
	case "#%RAML 1.0 Library":
		return FragmentLibrary, nil
	case "#%RAML 1.0 DataType":
		return FragmentDataType, nil
	case "#%RAML 1.0 NamedExample":
		return FragmentNamedExample, nil
	default:
		return FragmentUnknown, fmt.Errorf("unknown fragment kind: head: %s", head)
	}
}

// ReadRawFile reads a file
func ReadRawFile(path string) (io.ReadCloser, error) {
	f, err := openFragmentFile(path)
	if err != nil {
		return nil, fmt.Errorf("open fragment file: %w", err)
	}

	return f, nil
}

func (r *RAML) decodeDataType(f io.Reader, path string) (*DataType, error) {
	// TODO: This is a temporary workaround for JSON data types.
	if strings.HasSuffix(path, ".json") {
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, NewWrappedError("read file", err, path, WithType(ErrTypeReading))
		}
		dt, err := r.MakeJsonDataType(data, path)
		if err != nil {
			return nil, NewWrappedError("make json data type", err, path, WithType(ErrTypeParsing))
		}
		r.PutFragment(path, dt)
		return dt, nil
	}

	decoder := yaml.NewDecoder(f)

	dt := r.MakeDataType(path)
	if err := decoder.Decode(&dt); err != nil {
		return nil, fmt.Errorf("decode fragment: %w", err)
	}

	r.PutFragment(path, dt)

	baseDir := filepath.Dir(dt.Location)
	for pair := dt.Uses.Oldest(); pair != nil; pair = pair.Next() {
		include := pair.Value
		sublib, err := r.parseLibrary(filepath.Join(baseDir, include.Value))
		if err != nil {
			return nil, fmt.Errorf("parse library: %w", err)
		}
		include.Link = sublib
	}

	return dt, nil
}

func CheckFragmentKind(f *os.File, kind FragmentKind) error {
	// Allow JSON data types.
	if kind == FragmentDataType && strings.HasSuffix(f.Name(), ".json") {
		return nil
	}
	head, err := ReadHead(f)
	if err != nil {
		return fmt.Errorf("read head: %w", err)
	}
	frag, err := IdentifyFragment(head)
	if err != nil {
		return fmt.Errorf("identify fragment: %w", err)
	}
	if frag != kind {
		return fmt.Errorf("unexpected fragment frag != kind: %v != %v", frag, kind)
	}
	return nil
}

func (r *RAML) parseDataType(path string) (*DataType, error) {
	// IMPORTANT: May generate recursive structure.
	// Consumers (resolvers, validators, external clients) must implement recursion detection when traversing links.

	// Fragment paths must be normalized to absolute paths to simplify dependent libraries resolution.

	if dt := r.GetFragment(path); dt != nil {
		// log.Printf("reusing fragment %s", path)
		return dt.(*DataType), nil
	}

	f, err := openFragmentFile(path)
	if err != nil {
		return nil, fmt.Errorf("open fragment file: %w", err)
	}

	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			log.Fatalf("close file error: %v", err)
		}
	}(f)

	if err = CheckFragmentKind(f, FragmentDataType); err != nil {
		return nil, fmt.Errorf("check fragment kind: %w", err)
	}

	dt, err := r.decodeDataType(f, path)
	if err != nil {
		return nil, fmt.Errorf("decode data type: %w", err)
	}

	return dt, nil
}

func openFragmentFile(path string) (*os.File, error) {
	// TODO: Maybe fragments should be loaded against specified base URI.
	// If base URI is not specified, use current workdir.
	if !filepath.IsAbs(path) {
		workdir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get workdir: %w", err)
		}
		path = filepath.Join(workdir, path)
	}
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

func (r *RAML) decodeLibrary(f io.Reader, path string) (*Library, error) {
	decoder := yaml.NewDecoder(f)

	lib := r.MakeLibrary(path)
	if err := decoder.Decode(&lib); err != nil {
		return nil, fmt.Errorf("decode fragment: %w", err)
	}

	r.PutFragment(path, lib)

	// Resolve included libraries in a separate stage.
	baseDir := filepath.Dir(lib.Location)
	for pair := lib.Uses.Oldest(); pair != nil; pair = pair.Next() {
		include := pair.Value

		slibloc := filepath.Join(baseDir, include.Value)
		sublib, err := r.parseLibrary(slibloc)
		if err != nil {
			return nil, NewWrappedError("parse uses library", err, slibloc)
		}
		include.Link = sublib
	}
	return lib, nil
}

func (r *RAML) parseLibrary(path string) (*Library, error) {
	// IMPORTANT: May generate recursive structure.
	// Consumers (resolvers, validators, external clients) must implement recursion detection when traversing links.

	// Fragment paths must be normalized to absolute paths to simplify dependent libraries resolution.
	// TODO: Add support for URI
	var err error

	if lib := r.GetFragment(path); lib != nil {
		slog.Debug("reusing fragment", slog.String("path", path))
		return lib.(*Library), nil
	}

	f, err := openFragmentFile(path)
	if err != nil {
		return nil, fmt.Errorf("open fragment file: %w", err)
	}

	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			log.Fatalf("close file error: %v", err)
		}
	}(f)

	if err = CheckFragmentKind(f, FragmentLibrary); err != nil {
		return nil, fmt.Errorf("check fragment kind: %w", err)
	}

	lib, err := r.decodeLibrary(f, path)
	if err != nil {
		return nil, fmt.Errorf("decode library: %w", err)
	}

	return lib, nil
}

func (r *RAML) decodeNamedExample(f io.Reader, path string) (*NamedExample, error) {
	decoder := yaml.NewDecoder(f)

	ne := r.MakeNamedExample(path)
	if err := decoder.Decode(&ne); err != nil {
		return nil, fmt.Errorf("decode fragment: %w", err)
	}

	r.PutFragment(path, ne)

	return ne, nil
}

func (r *RAML) parseNamedExample(path string) (*NamedExample, error) {
	// Library paths must be normalized to simplify dependent libraries resolution.
	// Convert rel to abs relative to current workdir if necessary.

	if lib := r.GetFragment(path); lib != nil {
		slog.Debug("reusing fragment", slog.String("path", path))
		return lib.(*NamedExample), nil
	}

	f, err := openFragmentFile(path)
	if err != nil {
		return nil, fmt.Errorf("open fragment file: %w", err)
	}

	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			log.Fatalf("close file error: %v", err)
		}
	}(f)

	if err = CheckFragmentKind(f, FragmentNamedExample); err != nil {
		return nil, fmt.Errorf("check fragment kind: %w", err)
	}

	ne, err := r.decodeNamedExample(f, path)
	if err != nil {
		return nil, fmt.Errorf("decode named example: %w", err)
	}

	return ne, nil
}

func (r *RAML) ParseFromPath(path string, opts ...ParseOpt) error {
	// Library paths must be normalized to simplify dependent libraries resolution.
	// Convert rel to abs relative to current workdir if necessary.

	pOpts := &parserOptions{}
	for _, opt := range opts {
		opt.Apply(pOpts)
	}

	f, err := openFragmentFile(path)
	if err != nil {
		return fmt.Errorf("open fragment file: %w", err)
	}

	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			log.Fatalf("close file error: %v", err)
		}
	}(f)

	return r.parseFragment(f, f.Name(), pOpts)
}

func (r *RAML) ParseFromString(content string, fileName string, baseDir string, opts ...ParseOpt) error {
	pOpts := &parserOptions{}
	for _, opt := range opts {
		opt.Apply(pOpts)
	}

	f := strings.NewReader(content)

	return r.parseFragment(f, filepath.Join(baseDir, fileName), pOpts)
}

func (r *RAML) parseFragment(f io.ReadSeeker, fragmentPath string, pOpts *parserOptions) error {
	head, err := ReadHead(f)
	if err != nil {
		return fmt.Errorf("read head: %w", err)
	}
	frag, err := IdentifyFragment(head)
	if err != nil {
		return fmt.Errorf("identify fragment: %w", err)
	}
	switch frag {
	case FragmentLibrary:
		lib, err := r.decodeLibrary(f, fragmentPath)
		if err != nil {
			return fmt.Errorf("parse library: %w", err)
		}
		r.SetEntryPoint(lib)
	case FragmentDataType:
		dt, err := r.decodeDataType(f, fragmentPath)
		if err != nil {
			return fmt.Errorf("parse data type: %w", err)
		}
		r.SetEntryPoint(dt)
	case FragmentNamedExample:
		ne, err := r.decodeNamedExample(f, fragmentPath)
		if err != nil {
			return fmt.Errorf("parse named example: %w", err)
		}
		r.SetEntryPoint(ne)
	default:
		return fmt.Errorf("unknown fragment kind: head: %s", head)
	}

	err = r.resolveShapes()
	if err != nil {
		return fmt.Errorf("resolve shapes: %w", err)
	}
	err = r.resolveDomainExtensions()
	if err != nil {
		return fmt.Errorf("resolve domain extensions: %w", err)
	}

	if pOpts.withUnwrapOpt {
		err = r.UnwrapShapes()
		if err != nil {
			return fmt.Errorf("unwrap shapes: %w", err)
		}
	}

	if pOpts.withValidateOpt {
		err = r.ValidateShapes()
		if err != nil {
			return fmt.Errorf("validate shapes: %w", err)
		}
	}

	return nil
}

func ParseFromPathCtx(ctx context.Context, path string, opts ...ParseOpt) (*RAML, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	rml := New(ctx)
	err := rml.ParseFromPath(path, opts...)
	return rml, err
}

func ParseFromPath(path string, opts ...ParseOpt) (*RAML, error) {
	return ParseFromPathCtx(context.Background(), path, opts...)
}

func ParseFromStringCtx(ctx context.Context, content string, fileName string, baseDir string, opts ...ParseOpt) (*RAML, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	rml := New(ctx)
	err := rml.ParseFromString(content, fileName, baseDir, opts...)
	return rml, err
}

func ParseFromString(content string, fileName string, baseDir string, opts ...ParseOpt) (*RAML, error) {
	// TODO: Probably needs to be a bit more flexible. Maybe baseDir must be defined as parser option?
	if !filepath.IsAbs(baseDir) {
		return nil, fmt.Errorf("baseDir must be an absolute path")
	}
	return ParseFromStringCtx(context.Background(), content, fileName, baseDir, opts...)
}

type parserOptions struct {
	withUnwrapOpt   bool
	withValidateOpt bool
}

type ParseOpt interface {
	Apply(*parserOptions)
}

type parseOptWithUnwrap struct{}

func (parseOptWithUnwrap) Apply(opt *parserOptions) {
	opt.withUnwrapOpt = true
}

func OptWithUnwrap() ParseOpt {
	return parseOptWithUnwrap{}
}

type parseOptWithValidate struct{}

func (parseOptWithValidate) Apply(opt *parserOptions) {
	opt.withValidateOpt = true
}

func OptWithValidate() ParseOpt {
	return parseOptWithValidate{}
}
