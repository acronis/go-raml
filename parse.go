package raml

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TODO: Add NamedExample fragment parsing

func IdentifyFragment(head string) FragmentKind {
	if head == "#%RAML 1.0 Library" {
		return FragmentLibrary
	} else if head == "#%RAML 1.0 DataType" {
		return FragmentDataType
	} else if head == "#%RAML 1.0 NamedExample" {
		return FragmentNamedExample
	} else {
		return FragmentUnknown
	}
}

func NewFragmentDecoder(f *os.File, kind FragmentKind) (*yaml.Decoder, error) {
	r := bufio.NewReader(f)
	head, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read fragment head: %w", err)
	}
	frag := IdentifyFragment(strings.TrimRight(head, "\r\n"))
	if frag != kind {
		return nil, fmt.Errorf("unexpected RAML fragment kind")
	}

	return yaml.NewDecoder(r), nil
}

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
		log.Fatalf("open file error: %v", err)
		return nil, fmt.Errorf("open file: %w", err)
	}

	return f, nil
}

func ParseDataType(path string) (*DataType, error) {
	// NOTE: May generate recursive structure.
	// Fragment paths must be normalized to absolute paths to simplify dependent libraries resolution.
	// TODO: Add support for URI
	if !filepath.IsAbs(path) {
		workdir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get workdir: %w", err)
		}
		path = filepath.Join(workdir, path)
	}

	if dt := GetRegistry().GetFragment(path); dt != nil {
		// log.Printf("reusing fragment %s", path)
		return dt.(*DataType), nil
	}

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			log.Fatal(fmt.Errorf("close file error: %w", err))
		}
	}(f)

	// TODO: This is a temporary workaround for JSON data types.
	if strings.HasSuffix(path, ".json") {
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}
		dt, err := MakeJsonDataType(data, path)
		if err != nil {
			return nil, fmt.Errorf("make json data type: %w", err)
		}
		GetRegistry().PutFragment(path, dt)
		return dt, nil
	}

	decoder, err := NewFragmentDecoder(f, FragmentDataType)
	if err != nil {
		return nil, fmt.Errorf("new fragment decoder: %w", err)
	}

	dt := MakeDataType(path)
	if err := decoder.Decode(&dt); err != nil {
		return nil, fmt.Errorf("decode fragment: %w", err)
	}
	GetRegistry().PutFragment(path, dt)

	baseDir := filepath.Dir(dt.Location)
	for _, include := range dt.Uses {
		// IMPORTANT: Library recursion will NEVER happen inside data types on the parser level since data type is a leaf fragment.
		// However, consumers (resolvers, external clients) must implement recursion detection when traversing links.
		sublib, err := ParseLibrary(filepath.Join(baseDir, include.Value), make([]string, 0))
		if err != nil {
			return nil, fmt.Errorf("parse library: %w", err)
		}
		include.Link = sublib
	}

	return dt, nil
}

func ParseLibrary(path string, history []string) (*Library, error) {
	// NOTE: May generate recursive structure.
	// Fragment paths must be normalized to absolute paths to simplify dependent libraries resolution.
	// TODO: Add support for URI
	if !filepath.IsAbs(path) {
		workdir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get workdir: %w", err)
		}
		path = filepath.Join(workdir, path)
	}

	// NOTE: Library recursion may happen on the parser level in cases when a library includes itself directly or indirectly.
	for i, item := range history {
		if item == path {
			p := strings.Join(history[i:], " -> ")
			return nil, fmt.Errorf("circular reference detected: %s -> %s", p, path)
		}
	}
	// We need to create a new history slice to avoid modifying the original one in subsequent calls.
	newHistory := make([]string, len(history)+1)
	copy(newHistory, history)
	newHistory[len(history)] = path

	if lib := GetRegistry().GetFragment(path); lib != nil {
		// log.Printf("reusing fragment %s", path)
		return lib.(*Library), nil
	}

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			log.Fatal(fmt.Errorf("close file error: %w", err))
		}
	}(f)

	decoder, err := NewFragmentDecoder(f, FragmentLibrary)
	if err != nil {
		return nil, fmt.Errorf("new fragment decoder: %w", err)
	}

	lib := MakeLibrary(path)
	if err := decoder.Decode(&lib); err != nil {
		return nil, fmt.Errorf("decode fragment: %w", err)
	}
	GetRegistry().PutFragment(path, lib)

	// Resolve included libraries in a separate stage.
	baseDir := filepath.Dir(lib.Location)
	for _, include := range lib.Uses {
		sublib, err := ParseLibrary(filepath.Join(baseDir, include.Value), newHistory)
		if err != nil {
			return nil, fmt.Errorf("parse library: %w", err)
		}
		include.Link = sublib
	}

	return lib, nil
}

func ParseNamedExample(path string) (*NamedExample, error) {
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

	if lib := GetRegistry().GetFragment(path); lib != nil {
		// log.Printf("reusing fragment %s", path)
		return lib.(*NamedExample), nil
	}

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			log.Fatal(fmt.Errorf("close file error: %w", err))
		}
	}(f)

	decoder, err := NewFragmentDecoder(f, FragmentNamedExample)
	if err != nil {
		return nil, fmt.Errorf("new fragment decoder: %w", err)
	}

	ne := MakeNamedExample(path)
	if err := decoder.Decode(&ne); err != nil {
		return nil, fmt.Errorf("decode fragment: %w", err)
	}
	GetRegistry().PutFragment(path, ne)

	return ne, nil
}
