# Ordered float32 reductions

## Unfused dot product

`DotScalar(x,y []float32) (float32,error)` rounds each product to float32,
then adds left to right from +0. `DotSIMD` uses separate vector Mul/Add with
one accumulator, a low-to-high lane fold and rounded scalar-tail products.
Length validation, multiplication, extraction/fold and tail are timed.
Unequal lengths return +0,ErrLengthMismatch before reading either input;
equal empty/nil inputs return +0,nil. Read-only overlap is allowed.

```powershell
go run ./examples/dot
go run ./cmd/simdreport -group dot -dot-input cancellation -bench '^BenchmarkDot$/^n=64$' -count 3 -benchtime 100ms -format md
```

`-dot-input` selects balanced, cancellation or mixed-magnitude fixtures, with
both input arrays fingerprinted and no output array. Logical work is `8*n+4`.
Typed checked callbacks propagate errors and retain result bits once after
timing under distinct dot harnesses. Scalar is a timing baseline; order changes
can alter rounding and exceptional values. FMA and multiple accumulators remain
future work. See the [dot contract](../../docs/features/F004-dot.md) and
[Wiki exercise](https://github.com/akolybelnikov/go-simd/wiki/15-Dot-product-and-unfused-rounding).

## Sum

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
