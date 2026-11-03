# Ordered float32 reductions

## Fused dot product

DotFMAScalar and DotFMASIMD/2/4 return `(float32,error)` with one correctly rounded
binary32 FMA per update. Scalar proceeds left to right. SIMD assigns full vectors
round-robin to 1/2/4 accumulators, merges lane-wise left to right, folds lanes low
to high from +0, then performs software FMA tail updates. Merge/fold additions
round separately; this is not a single-round whole-dot operation.

The startup gate chooses native FMA or software stripes preserving the same
order. Case.FMA records the actual reference/striped/vector/hybrid/tail FMA path;
Case.Reduction records width, count and fold, and ScalarResultBits retains the
result once after timing. Validation, cached dispatch, updates, merging, extraction
and tails are timed under standalone-dot-fma/v1 or legacy-report-dot-fma/v1.
Current widths allocate zero per call. Logical work is `8*n+4` bytes.

Unequal lengths return +0,ErrLengthMismatch before reading inputs. Equal empty
inputs return +0,nil; inputs stay unchanged and read-only overlap is allowed.
Exceptional values and signed zero follow each order; NaN payloads are unspecified.
The shared -dot-input adds fusion-sensitive alongside the original three recipes.

```powershell
go run ./examples/dotfma
go run ./cmd/simdreport -group dotfma -dot-input fusion-sensitive -bench '^BenchmarkDotFMA$/^n=64$' -count 3 -benchtime 100ms -format md
```

See the [contract](../../docs/features/F004-fused-dot.md) and
[worked guide](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders).
Three-array MulAdd remains pending.

## Two/four accumulator strategies

SumSIMD2/4 and DotSIMD2/4 assign full vectors round-robin to 2/4 independent
accumulators, merge lane-wise left to right, fold lanes low to high from +0,
then process the scalar tail. Dot products round before addition without FMA.
Existing input/error/empty contracts apply; no output slice or new CLI flag.

```powershell
go run ./examples/accumulators
go run ./cmd/simdreport -group dot -bench '^BenchmarkDot$/^n=1024$' -count 3 -benchtime 100ms -format md
```

All four strategies now appear for sum/dot. Algorithmic accumulator counts and
fold identity survive in v1 evidence; original strategy IDs and kernel bodies
stay unchanged. Current widths allocate zero per call. Comparisons include
unrolling, merging and Go's register allocation; more accumulators need not win.
See the [contract](../../docs/features/F004-accumulators.md) and
[Wiki guide](https://github.com/akolybelnikov/go-simd/wiki/16-Accumulator-count-experiments).

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

`-dot-input` selects balanced, cancellation, mixed-magnitude or fusion-sensitive fixtures, with
both input arrays fingerprinted and no output array. Logical work is `8*n+4`.
Typed checked callbacks propagate errors and retain result bits once after
timing under distinct dot harnesses. Scalar is a timing baseline; order changes
can alter rounding and exceptional values. Fused dot is a separate group above; multiple
accumulators are available above. See the [dot contract](../../docs/features/F004-dot.md) and
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
