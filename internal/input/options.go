package input

import (
	"fmt"
	"math"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

const NativeOptionsEnv = "GO_SIMD_INPUT_OPTIONS"

// Scalars are stored as exact IEEE-754 bits; JSON never contains NaN/Inf numbers.
type Options struct {
	Alpha, Beta, Threshold, Lower, Upper uint32
	Pattern                              string
	Seed                                 uint64
}

func DefaultOptions() Options {
	return Options{math.Float32bits(1.5), math.Float32bits(-0.5), 0, math.Float32bits(-1), math.Float32bits(1), "alternating", 42}
}

func (o Options) Bits(name string) uint32 {
	switch name {
	case "alpha":
		return o.Alpha
	case "beta":
		return o.Beta
	case "threshold":
		return o.Threshold
	case "lower":
		return o.Lower
	case "upper":
		return o.Upper
	}
	panic("unknown parameter " + name)
}

func (o Options) Float(name string) float32 { return math.Float32frombits(o.Bits(name)) }

func (o Options) Identity(op catalog.Operation) string {
	var values []string
	for _, p := range op.Parameters {
		values = append(values, fmt.Sprintf("%s=float32:0x%08x", p.Name, o.Bits(p.Name)))
	}
	return strings.Join(values, ";")
}

func IsPredicate(op catalog.Operation) bool {
	return op.ID == catalog.ThresholdZero || op.ID == catalog.SelectGT || op.ID == catalog.RangeZero
}

func (o Options) Validate() error {
	switch o.Pattern {
	case "none", "all", "alternating", "clustered", "shuffled":
		return nil
	}
	return fmt.Errorf("unknown pattern %q (choose none, all, alternating, clustered, or shuffled)", o.Pattern)
}

func (o Options) ValidateFor(op catalog.Operation) error {
	if err := o.Validate(); err != nil {
		return err
	}
	if IsPredicate(op) && o.Pattern != "none" {
		_, value := o.predicateValues(op)
		if !o.matches(op, value) {
			return fmt.Errorf("%s cannot generate true predicates with these bounds; use -pattern none or change the parameters", op.Group)
		}
	}
	return nil
}

func (o Options) matches(op catalog.Operation, v float32) bool {
	if op.ID == catalog.RangeZero {
		return v > o.Float("lower") && v < o.Float("upper")
	}
	return v > o.Float("threshold")
}

func (o Options) predicateValues(op catalog.Operation) (float32, float32) {
	bound := o.Float("threshold")
	if op.ID == catalog.RangeZero {
		bound = o.Float("lower")
	}
	value := bound + 1
	if !o.matches(op, value) {
		value = math.Nextafter32(bound, float32(math.Inf(1)))
	}
	return bound, value
}

func (o Options) trueAt(i, n int) bool {
	switch o.Pattern {
	case "none":
		return false
	case "all":
		return true
	case "alternating":
		return i%2 == 0
	case "clustered":
		return i < (n+1)/2
	case "shuffled":
		// SplitMix64: deterministic independent index mixing, approximately 50%.
		v := uint64(i) + o.Seed + 0x9e3779b97f4a7c15
		v = (v ^ (v >> 30)) * 0xbf58476d1ce4e5b9
		v = (v ^ (v >> 27)) * 0x94d049bb133111eb
		return (v^(v>>31))>>63 != 0
	}
	return false
}

func (o Options) Generator(op catalog.Operation) string {
	if !op.Parameterized() {
		return GeneratedFloat32ID
	}
	if IsPredicate(op) {
		if o.Pattern == "shuffled" {
			return fmt.Sprintf("predicate-float32-v1/shuffled/seed=%d", o.Seed)
		}
		return "predicate-float32-v1/" + o.Pattern
	}
	return "generated-unary-float32-v1"
}

func (o Options) pair(op catalog.Operation, i, n int) (float32, float32) {
	if !op.Parameterized() {
		return generatedPair(i)
	}
	if IsPredicate(op) {
		no, yes := o.predicateValues(op)
		if o.trueAt(i, n) {
			no = yes
		}
		return no, float32(i%17-8) / 4
	}
	return float32(i%33-16) / 4, 0
}

// MakeForOperation allocates only the declared arrays and one output.
func MakeForOperation(op catalog.Operation, n int, o Options) (a, b, out []float32) {
	if !op.Parameterized() {
		return MakeFloat32Slices(n)
	}
	a, out = make([]float32, n), make([]float32, n)
	if len(op.Inputs) == 2 {
		b = make([]float32, n)
	}
	for i := range a {
		x, y := o.pair(op, i, n)
		a[i] = x
		if b != nil {
			b[i] = y
		}
	}
	return
}

func FingerprintForOperation(op catalog.Operation, n int, o Options) string {
	if !op.Parameterized() {
		return GeneratedFloat32Fingerprint(n)
	}
	return fingerprintGenerated(n, len(op.Inputs) == 2, func(i int) (float32, float32) { return o.pair(op, i, n) })
}
