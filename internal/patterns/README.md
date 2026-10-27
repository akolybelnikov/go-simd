# Float32 broadcast and predicate patterns

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
