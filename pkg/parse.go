package goraml

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
		return LIBRARY
	} else if head == "#%RAML 1.0 DataType" {
		return DATATYPE
	} else if head == "#%RAML 1.0 NamedExample" {
		return NAMED_EXAMPLE
	} else {
		return UNKNOWN
	}
}

func NewFragmentDecoder(f *os.File, kind FragmentKind) (*yaml.Decoder, error) {
	r := bufio.NewReader(f)
	head, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	frag := IdentifyFragment(strings.TrimRight(head, "\r\n"))
	if frag != kind {
		return nil, fmt.Errorf("unexpected RAML fragment kind")
	}

	return yaml.NewDecoder(r), nil
}

func ParseDataType(path string) (*DataType, error) {
	// Library paths must be normalized to simplify dependent libraries resolution.
	// Convert rel to abs relative to current workdir if necessary.
	// TODO: Add support for URI
	if !filepath.IsAbs(path) {
		workdir, err := os.Getwd()
		if err != nil {
			return nil, err
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
		return nil, err
	}
	defer f.Close()

	// TODO: This is a temporary workaround for JSON data types.
	if strings.HasSuffix(path, ".json") {
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		dt := MakeDataType(path)
		if err := dt.UnmarshalJSON(data); err != nil {
			return nil, err
		}
		GetRegistry().PutFragment(path, dt)
		return dt, nil
	}

	decoder, err := NewFragmentDecoder(f, DATATYPE)
	if err != nil {
		return nil, err
	}

	dt := MakeDataType(path)
	if err := decoder.Decode(&dt); err != nil {
		return nil, err
	}
	GetRegistry().PutFragment(path, dt)

	return dt, nil
}

func ParseLibrary(path string) (*Library, error) {
	// Library paths must be normalized to simplify dependent libraries resolution.
	// Convert rel to abs relative to current workdir if necessary.
	// TODO: Add support for URI
	if !filepath.IsAbs(path) {
		workdir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(workdir, path)
	}

	if lib := GetRegistry().GetFragment(path); lib != nil {
		// log.Printf("reusing fragment %s", path)
		return lib.(*Library), nil
	}

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return nil, err
	}
	defer f.Close()

	decoder, err := NewFragmentDecoder(f, LIBRARY)
	if err != nil {
		return nil, err
	}

	lib := MakeLibrary(path)
	if err := decoder.Decode(&lib); err != nil {
		return nil, err
	}
	GetRegistry().PutFragment(path, lib)

	return lib, nil
}