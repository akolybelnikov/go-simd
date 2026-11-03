# Float32 broadcast, predicate and fused-affine patterns

The affinefma group adds AffineFMAScalar/AffineFMASIMD with one binary32
rounding. Its corrected reference avoids widened-FMA double rounding. A startup
gate selects verified vector MulAdd or an explicitly recorded software path.
Old affine/dot remain unfused. Array alias/length rules and 8*n accounting match
unary affine. Run `go run ./examples/fma`; see the
[contract](../../docs/features/F004-fma.md) and [Wiki guide](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding).

## Predicate byte export

PredicateGTScalar/PredicateGTSIMD accept
`(x []float32,out []uint8,threshold float32) error`. They export ordered strict
comparison as canonical 0/1. Equality and NaN are false. Length errors precede
writes; empty succeeds; input/output storage must be disjoint.

SIMDPartial broadcasts threshold/one/zero, loads partial vectors, selects exact
float32 0/1, converts/stores int32 scratch and writes only valid output bytes.
No raw SDK mask encoding escapes, including padded true lanes under negative
thresholds. Conversion/extraction/stores are timed. Logical work is `5*n`;
current widths allocate zero per call. Typed byte bindings and separate harnesses
retain full output SHA-256 after timing alongside width/encoding/input identity.

```powershell
go run ./examples/predicate
go run ./cmd/simdreport -group predicate '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicate$/^n=64$' -count 3 -benchtime 100ms -format md
```

Shared threshold/pattern/seed controls work through installed CLI, Go and Make.
See the [contract](../../docs/features/F004-predicate-export.md) and
[guide](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks). Packed bits and other representations remain separate work.

## Original broadcast/predicate slice

Five operations each have Scalar and SIMDPartial implementations:

| Functions | Rule |
| --- | --- |
| ScaleScalar / ScaleSIMD | `alpha*x` |
| AffineScalar / AffineSIMD | `float32(alpha*x)+beta`, unfused |
| ThresholdZeroScalar / ThresholdZeroSIMD | `x > threshold ? x : +0` |
| SelectGTScalar / SelectGTSIMD | `x > threshold ? x : y` |
| RangeZeroScalar / RangeZeroSIMD | `lower < x && x < upper ? x : +0` |

SIMD broadcasts constants once per call and uses partial loads/stores. Masked
false lanes become +0; selection preserves chosen operand bits. Strict NaN
comparisons are false. Exact output aliases are supported; partial overlap is
unsupported. Length mismatches leave output unchanged. Arithmetic NaN payloads
are not fixed, while finite results/signed zeros follow float32 arithmetic.

Run `go run ./examples/patterns` for small values, or inspect/report an experiment:

```powershell
go run ./cmd/simdreport -list -group selection
go run ./cmd/simdreport -group selection -threshold 0 -pattern shuffled -seed 42 -bench '^BenchmarkSelection$/^n=1024$' -count 3 -benchtime 100ms -format md
```

Set GOEXPERIMENT=simd for source builds. Installed commands use the same flags
without requiring Go/source. New groups are scale, affine, threshold, selection,
rangefilter. For recorded custom parameters through native Go benchmarks, use
the same report command with `-backend go-test`. Direct native tests/benchmarks
use the defaults unless the internal options environment is supplied.

See [the F004 implementation contract](../../docs/features/F004-broadcast-predicates.md)
for bit-level parameters, fixture distributions, traffic, timing and evidence.
