package raml

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_main(t *testing.T) {
	start := time.Now()
	rml, err := ParseFromPath(`./tests/library.raml`, OptWithUnwrap())
	if vErr, ok := UnwrapError(err); ok {
		t.Logf("Validation error: %s", vErr.Error())
		err = vErr
	}
	require.NoError(t, err)
	elapsed := time.Since(start)
	t.Logf("ParseFromPath took %d ms\n", elapsed.Milliseconds())
	fmt.Printf("Library location: %s\n", rml.entryPoint.GetLocation())

	shapesAll := rml.GetShapes()
	fmt.Printf("Total shapes: %d\n", len(shapesAll))

	for _, shape := range shapesAll {
		s, unresolved := shape.(*UnknownShape)
		if unresolved {
			t.Errorf("Unknown shape found %s", s.Name)
		}
		fmt.Printf("Shape: %s: resolved: %v: unwrapped: %v\n", shape, !unresolved, shape.Base().unwrapped)
	}

	fmt.Printf("Resolved: %d\n", resolved)
	fmt.Printf("Unresolved: %d\n", unresolved)

	printMemUsage(t)
}

func printMemUsage(t *testing.T) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Alloc = %v MiB", m.Alloc/1024/1024)
	t.Logf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	t.Logf("\tSys = %v MiB", m.Sys/1024/1024)
	t.Logf("\tNumGC = %v\n", m.NumGC)
}
