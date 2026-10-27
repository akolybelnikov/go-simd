package result

import (
	"strings"
	"testing"
)

func TestHistoricalPatternParametersStayUnknown(t *testing.T) {
	run, err := ReadLegacy(strings.NewReader("BenchmarkScale/n=8/Scalar-4 1 20 ns/op"), LegacyOptions{RunID: "historical", GeneratedInputs: true})
	if err != nil {
		t.Fatal(err)
	}
	c := run.Observations[0].Key.Case
	if c.Parameters != "unknown" || c.Input.SHA256 != "" || c.Input.GeneratorID != "" {
		t.Fatal("invented historical configuration")
	}
}
