package evidence

import (
	"bytes"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/result"
	"strings"
	"testing"
)

func TestReadCopy64FullBits(t *testing.T) {
	c := validCapture()
	c.Group = "copy64"
	o := &c.Run.Observations[0]
	o.Key.Case.OperationID = catalog.Copy64
	o.Key.Case.TypeID = catalog.Float64
	o.Key.Case.Vector, _ = result.NewVectorIdentity(4)
	o.ArrayResultSHA256 = strings.Repeat("a", 64)
	if _, e := ReadCapture(bytes.NewReader(encodeCapture(t, c)), DefaultReadLimits()); e != nil {
		t.Fatal(e)
	}
	o.ArrayResultSHA256 = ""
	if _, e := ReadCapture(bytes.NewReader(encodeCapture(t, c)), DefaultReadLimits()); readCode(e) != "invalid_copy64_evidence" {
		t.Fatal(e)
	}
}
