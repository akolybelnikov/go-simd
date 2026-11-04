package evidence

import (
	"bytes"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/result"
	"strings"
	"testing"
)

func TestReadClmul64Capability(t *testing.T) {
	c := validCapture()
	c.Group = "clmul64"
	o := &c.Run.Observations[0]
	o.Key.Case.OperationID = catalog.Clmul64
	o.Key.Case.TypeID = catalog.Uint64
	o.Key.Case.Vector, _ = result.NewVectorIdentity(4)
	o.Key.Case.Vector.CarrylessCapabilityID = result.CarrylessCapability(false)
	o.ArrayResultSHA256 = strings.Repeat("a", 64)
	if _, e := ReadCapture(bytes.NewReader(encodeCapture(t, c)), DefaultReadLimits()); e != nil {
		t.Fatal(e)
	}
	o.Key.Case.Vector.CarrylessCapabilityID = ""
	if _, e := ReadCapture(bytes.NewReader(encodeCapture(t, c)), DefaultReadLimits()); readCode(e) != "invalid_carryless_evidence" {
		t.Fatal(e)
	}
}
