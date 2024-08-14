package main_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	goraml "github.com/acronis/go-raml/pkg"
	"github.com/stretchr/testify/require"
)

func Test_main(t *testing.T) {
	start := time.Now()
	lib, err := goraml.ParseLibrary(`./tests/library.raml`)
	require.NoError(t, err)
	elapsed := time.Since(start)
	t.Logf("ParseLibrary took %d ms", elapsed.Milliseconds())
	fmt.Printf("Library location: %s\n", lib.Location)

	vals := goraml.GetRegistry().GetAllShapes()
	fmt.Printf("Total shapes: %d\n", len(vals))
	fmt.Printf("Unresolved: %d\n", len(goraml.GetRegistry().UnresolvedShapes))

	require.NoError(t, goraml.ResolveShapes())

	fmt.Printf("Resolved: %d\n", len(goraml.GetRegistry().ResolvedShapes))

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
