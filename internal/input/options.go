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
	FMAInput                             string
	DotInput                             string
	SumInput                             string
	Alpha, Beta, Threshold, Lower, Upper uint32
	Pattern                              string
	Seed                                 uint64
}

func DefaultOptions() Options {
	return Options{Alpha: math.Float32bits(1.5), Beta: math.Float32bits(-0.5), Lower: math.Float32bits(-1), Upper: math.Float32bits(1), Pattern: "alternating", Seed: 42, SumInput: "balanced", DotInput: "balanced", FMAInput: "balanced"}
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
	return op.ID == catalog.ThresholdZero || op.ID == catalog.SelectGT || op.ID == catalog.RangeZero || op.ID == catalog.PredicateGT
}

func (o Options) Validate() error {
	switch o.FMAFixture() {
	case "balanced", "rounding", "mixed-magnitude":
	default:
		return fmt.Errorf("unknown fma-input %q (choose balanced, rounding, or mixed-magnitude)", o.FMAInput)
	}
	switch o.DotFixture() {
	case "balanced", "cancellation", "mixed-magnitude", "fusion-sensitive":
	default:
		return fmt.Errorf("unknown dot-input %q (choose balanced, cancellation, mixed-magnitude, or fusion-sensitive)", o.DotInput)
	}
	switch o.SumFixture() {
	case "balanced", "cancellation", "mixed-magnitude":
	default:
		return fmt.Errorf("unknown sum-input %q (choose balanced, cancellation, or mixed-magnitude)", o.SumInput)
	}
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
	if op.ID == catalog.AffineFMA {
		return "generated-affine-fma-float32-v1/" + o.FMAFixture()
	}
	if op.ID == catalog.Dot {
		return "generated-dot-float32-v1/" + o.DotFixture()
	}
	if op.ID == catalog.DotFMA {
		return "generated-dot-fma-float32-v1/" + o.DotFixture()
	}
	if op.ID == catalog.Sum {
		return "generated-sum-float32-v1/" + o.SumFixture()
	}
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
	if op.ID == catalog.AffineFMA {
		switch o.FMAFixture() {
		case "rounding":
			v := [8]float32{1 - 0x1p-23, 1 + 0x1p-23, 1, -1, math.SmallestNonzeroFloat32, math.MaxFloat32, math.Float32frombits(0x80000000), 1 - 0x1p-23}
			return v[i%8], 0
		case "mixed-magnitude":
			v := [8]float32{1e-20, 1, 1e20, -1e20, -1, 3, -3, 1e-20}
			return v[i%8], 0
		default:
			return float32(i%33-16) / 4, 0
		}
	}
	if op.ID == catalog.Dot || op.ID == catalog.DotFMA {
		switch o.DotFixture() {
		case "fusion-sensitive":
			switch i % 64 {
			case 0:
				return -1, 1
			case 32:
				return 1 + 0x1p-23, 1 - 0x1p-23
			}
			return 0, 0
		case "cancellation":
			switch i % 64 {
			case 0:
				return 16777216, 1
			case 1:
				return 1, 1
			case 32:
				return -16777216, 1
			}
			return 0, 1
		case "mixed-magnitude":
			x := [8]float32{1e-20, 1, 1e20, -1e20, -1, 3, -3, 1e-20}
			y := [8]float32{1e20, 1e-20, 1e-20, 1e-20, 1e20, 1e-20, 1e-20, 1e20}
			return x[i%8], y[i%8]
		default:
			return float32(i%33-16) / 4, float32(i%17-8) / 8
		}
	}
	if op.ID == catalog.Sum {
		switch o.SumFixture() {
		case "cancellation":
			switch i % 64 {
			case 0:
				return 16777216, 0
			case 1:
				return 1, 0
			case 32:
				return -16777216, 0
			}
			return 0, 0
		case "mixed-magnitude":
			values := [8]float32{1e-20, 1, 1e20, -1e20, -1, 3, -3, 1e-20}
			return values[i%8], 0
		default:
			return float32(i%33-16) / 4, 0
		}
	}
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

// MakeForOperation allocates the declared input arrays and an array output only
// when required. Scalar returns and byte outputs have no float32 output slice;
// callers allocate the typed byte output separately outside measurement.
func MakeForOperation(op catalog.Operation, n int, o Options) (a, b, out []float32) {
	if op.ByteResult() {
		a = make([]float32, n)
		for i := range a {
			a[i], _ = o.pair(op, i, n)
		}
		return
	}
	if op.ScalarResult() {
		a = make([]float32, n)
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
	if op.ScalarResult() {
		return fingerprintGenerated(n, len(op.Inputs) == 2, func(i int) (float32, float32) { return o.pair(op, i, n) })
	}
	if !op.Parameterized() {
		return GeneratedFloat32Fingerprint(n)
	}
	return fingerprintGenerated(n, len(op.Inputs) == 2, func(i int) (float32, float32) { return o.pair(op, i, n) })
}

// Empty means the original default when decoding older option records.
func (o Options) SumFixture() string {
	if o.SumInput == "" {
		return "balanced"
	}
	return o.SumInput
}

// DotFixture supplies the default for historical option records.
func (o Options) DotFixture() string {
	if o.DotInput == "" {
		return "balanced"
	}
	return o.DotInput
}

func (o Options) FMAFixture() string {
	if o.FMAInput == "" {
		return "balanced"
	}
	return o.FMAInput
}
