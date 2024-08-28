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
	t.Logf("ParseFromPath took %d ms", elapsed.Milliseconds())
	fmt.Printf("Library location: %s\n", rml.entryPoint.GetLocation())

	shapesAll := rml.GetAllShapes()
	fmt.Printf("Total shapes: %d\n", len(shapesAll))
	//fmt.Printf("Unresolved: %d\n", len(rml.unresolvedShapes))
	//fmt.Printf("Resolved: %d\n", len(rml.shapes))

	resolved := 0
	unresolved := 0
	for _, shape := range shapesAll {
		if shape.Base().resolved {
			resolved++
		} else {
			unresolved++
		}
		fmt.Printf("Shape: %s: resolved: %v: unwrapped: %v\n", shape, shape.Base().resolved, shape.Base().unwrapped)
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
