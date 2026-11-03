# Ordered float32 sum

`SumScalar(x []float32) float32` accumulates from +0 left to right.
`SumSIMD(x []float32) float32` uses one native-width vector accumulator, extracts
and folds lanes low to high from +0, then adds the scalar tail. Each addition
rounds to float32. Extraction/final fold are included in whole-kernel timing.

Inputs are unchanged. Empty/nil input returns +0. NaN/Inf/overflow behavior is
order-dependent and NaN payloads are not fixed. Different folds need not agree
bitwise; scalar is the timing baseline, not an exact mathematical oracle.

```powershell
go run ./examples/sum
go run ./cmd/simdreport -group sum -sum-input cancellation -bench '^BenchmarkSum$/^n=64$' -count 3 -benchtime 100ms -format md
```

Set GOEXPERIMENT=simd for source builds. Installed commands use the same flags.
`-sum-input` selects balanced, cancellation or mixed-magnitude fixtures. For
recorded native comparisons, add `-backend go-test`; plain native benchmarks
default to balanced. Logical accounting is one input and one scalar result,
`4*n + 4` bytes/call, rather than an output array.

The JSON companion records scalar result bits and complete fold/width identity
separately from timing statistics. See the
[implementation contract](../../docs/features/F004-sum.md) and
[worked Wiki guide](https://github.com/akolybelnikov/go-simd/wiki/14-Sum-and-reduction-order).
