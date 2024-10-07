package raml

import (
	"log/slog"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_ParseFromPath(t *testing.T) {
	start := time.Now()
	rml, err := ParseFromPath(`./fixtures/library.raml`, OptWithUnwrap(), OptWithValidate())
	require.NoError(t, err)
	elapsed := time.Since(start)
	shapesAll := rml.GetShapes()
	slog.Info("ParseFromPath", "took ms", elapsed.Milliseconds(), "location",
		rml.entryPoint.GetLocation(), "total shapes", len(shapesAll))

	for _, base := range shapesAll {
		shape := base.Shape
		require.NotNil(t, shape)
		_, ok := shape.(*UnknownShape)
		require.False(t, ok)
	}

	conv := NewJSONSchemaConverter()
	for _, frag := range rml.fragmentsCache {
		switch f := frag.(type) {
		case *Library:
			for pair := f.AnnotationTypes.Oldest(); pair != nil; pair = pair.Next() {
				s := pair.Value
				conv.Convert(s.Shape)
				// b, err := json.MarshalIndent(schema, "", "  ")
				// if err != nil {
				// 	t.Errorf("StackTrace marshalling schema: %s", err)
				// }
				// os.WriteFile(fmt.Sprintf("./out/%s_%s.json", s.Base().Name, s.Base().ID), b, 0644)
				// fmt.Println(string(b))
			}
			for pair := f.Types.Oldest(); pair != nil; pair = pair.Next() {
				s := pair.Value
				conv.Convert(s.Shape)
				// if err != nil {
				// 	t.Errorf("StackTrace converting shape: %s", err)
				// }
				// b, err := json.MarshalIndent(schema, "", "  ")
				// if err != nil {
				// 	t.Errorf("StackTrace marshalling schema: %s", err)
				// }
				// os.WriteFile(fmt.Sprintf("./out/%s_%d.json", s.Name, s.ID), b, 0644)
				// fmt.Println(string(b))
			}
		case *DataType:
			conv.Convert(f.Shape.Shape)
			// b, err := json.MarshalIndent(schema, "", "  ")
			// if err != nil {
			// 	t.Errorf("StackTrace marshalling schema: %s", err)
			// }
			// os.WriteFile(fmt.Sprintf("./out/%s_%s.json", s.Base().Name, s.Base().ID), b, 0644)
			// fmt.Println(string(b))
		}
	}

	printMemUsage(t)
}

func printMemUsage(t *testing.T) {
	var m runtime.MemStats
	t.Helper()
	runtime.ReadMemStats(&m)
	slog.Info("Memory usage", "alloc MiB", m.Alloc/1024/1024, "total alloc MiB",
		m.TotalAlloc/1024/1024, "sys MiB", m.Sys/1024/1024, "num GC", m.NumGC)
}
