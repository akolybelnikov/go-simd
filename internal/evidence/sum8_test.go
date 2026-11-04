package evidence

import (
	"bytes"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/result"
	"strings"
	"testing"
)

func TestReadSum8PhysicalExtent(t *testing.T) {
	c := validCapture()
	c.Group = "sum8"
	o := &c.Run.Observations[0]
	o.Key.Case.OperationID = catalog.Sum8
	o.Key.Case.OutputElements = result.OptionalInt{Value: 1, Known: true}
	o.ArrayResultSHA256 = strings.Repeat("a", 64)
	if _, e := ReadCapture(bytes.NewReader(encodeCapture(t, c)), DefaultReadLimits()); e != nil {
		t.Fatal(e)
	}
	for _, extent := range []result.OptionalInt{{}, {Value: 2, Known: true}} {
		o.Key.Case.OutputElements = extent
		if _, e := ReadCapture(bytes.NewReader(encodeCapture(t, c)), DefaultReadLimits()); readCode(e) != "invalid_output_extent" {
			t.Fatal(e)
		}
	}
}
