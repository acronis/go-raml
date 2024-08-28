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
func ReadHead(f *os.File) (string, error) {
	r := bufio.NewReader(f)
	head, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read fragment head: %w", err)
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("seek to start: %w", err)
	}

	r.Reset(f)

	head = strings.TrimRight(head, "\r\n ")
	return head, nil
}

// IdentifyFragment returns the kind of the fragment by its head.
func IdentifyFragment(head string) (FragmentKind, error) {
	if head == "#%RAML 1.0 Library" {
		return FragmentLibrary, nil
	} else if head == "#%RAML 1.0 DataType" {
		return FragmentDataType, nil
	} else if head == "#%RAML 1.0 NamedExample" {
		return FragmentNamedExample, nil
	} else {
		return FragmentUnknown, fmt.Errorf("unknown fragment kind: head: %s", head)
	}
}

// ReadRawFile reads a file
func ReadRawFile(path string) (io.ReadCloser, error) {
	// Library paths must be normalized to simplify dependent libraries resolution.
	// Convert rel to abs relative to current workdir if necessary.
	// TODO: Add support for URI
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

func (r *RAML) decodeDataType(f *os.File) (*DataType, error) {
	path := f.Name()

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

	dt := r.MakeDataType(f.Name())
	if err := decoder.Decode(&dt); err != nil {
		return nil, fmt.Errorf("decode fragment: %w", err)
	}

	r.PutFragment(f.Name(), dt)

	baseDir := filepath.Dir(dt.Location)
	for _, include := range dt.Uses {
		sublib, err := r.parseLibrary(filepath.Join(baseDir, include.Value))
		if err != nil {
			return nil, fmt.Errorf("parse library: %w", err)
		}
		include.Link = sublib
	}

	return dt, nil
}

func IsFragmentKind(f *os.File, kind FragmentKind) error {
	if kind == FragmentDataType {
		// Allow JSON data types.
		if strings.HasSuffix(f.Name(), ".json") {
			return nil
		}
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

	if err = IsFragmentKind(f, FragmentDataType); err != nil {
		return nil, fmt.Errorf("check fragment kind: %w", err)
	}

	dt, err := r.decodeDataType(f)
	if err != nil {
		return nil, fmt.Errorf("decode data type: %w", err)
	}

	return dt, nil
}

func openFragmentFile(path string) (*os.File, error) {
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

func (r *RAML) decodeLibrary(f *os.File) (*Library, error) {
	decoder := yaml.NewDecoder(f)

	lib := r.MakeLibrary(f.Name())
	if err := decoder.Decode(&lib); err != nil {
		return nil, fmt.Errorf("decode fragment: %w", err)
	}

	r.PutFragment(f.Name(), lib)

	// Resolve included libraries in a separate stage.
	baseDir := filepath.Dir(lib.Location)
	for _, include := range lib.Uses {
		sublib, err := r.parseLibrary(filepath.Join(baseDir, include.Value))
		if err != nil {
			return nil, NewWrappedError("parse uses library", err, sublib.Location)
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

	if err = IsFragmentKind(f, FragmentLibrary); err != nil {
		return nil, fmt.Errorf("check fragment kind: %w", err)
	}

	lib, err := r.decodeLibrary(f)
	if err != nil {
		return nil, fmt.Errorf("decode library: %w", err)
	}

	return lib, nil
}

func (r *RAML) decodeNamedExample(f *os.File) (*NamedExample, error) {
	path := f.Name()
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

	if err = IsFragmentKind(f, FragmentNamedExample); err != nil {
		return nil, fmt.Errorf("check fragment kind: %w", err)
	}

	ne, err := r.decodeNamedExample(f)
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
		lib, err := r.decodeLibrary(f)
		if err != nil {
			return fmt.Errorf("parse library: %w", err)
		}
		r.SetEntryPoint(lib)
	case FragmentDataType:
		dt, err := r.decodeDataType(f)
		if err != nil {
			return fmt.Errorf("parse data type: %w", err)
		}
		r.SetEntryPoint(dt)
	case FragmentNamedExample:
		ne, err := r.decodeNamedExample(f)
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

type parserOptions struct {
	withUnwrapOpt bool
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
