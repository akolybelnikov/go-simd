# 08 — Using the kernels in Go

Fused dot adds reduction.DotFMAScalar and reduction.DotFMASIMD/2/4, accepting
`(x,y []float32) (float32,error)`. They use corrected binary32 FMA updates,
ordinary rounded accumulator merges/lane folds and software FMA tails. Unequal
lengths return +0,reduction.ErrLengthMismatch before reading input; equal empty
inputs return +0,nil. Inputs remain unchanged and read-only overlap is supported.
There is no output slice; logical work is `8*n+4` bytes.

Run `go run ./examples/dotfma` for fusion, accumulator-count and signed-zero
examples. The [fused-dot guide](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders)
explains the exact order and actual vector/software/tail paths. Software fallback
preserves each striped order; outputs need not equal the scalar fold.

The current packages are under `internal/`, so Go restricts imports to this
module's permitted source tree. A separate application cannot simply import
`github.com/akolybelnikov/go-simd/internal/addition`. Use the checked-out lab and
its examples to experiment. There is no stable public library/SDK contract yet.

## Run the complete checked-in example

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/basic
```

[Read the example source](https://github.com/akolybelnikov/go-simd/blob/main/examples/basic/main.go).
For these ordinary exactly representable inputs, expected output is:

```text
addition/Scalar: [5 7 9]
addition/SIMDPartial: [5 7 9]
addition/SIMDTail: [5 7 9]
addition/SIMDParallel: [5 7 9]
computation/Scalar: [21 39 63]
computation/SIMD: [21 39 63]
addition/output aliases a: [5 7 9]
addition/output aliases b: [5 7 9]
addition/length mismatch: true; output=[123]
addition/empty: ok
```

This example demonstrates values and contracts. It does not time functions or
choose an optimal implementation.

## Minimal addition call

Inside a Go source file in this module:

```go
import "github.com/akolybelnikov/go-simd/internal/addition"

a := []float32{1, 2, 3}
b := []float32{4, 5, 6}
out := make([]float32, len(a))
if err := addition.AddSIMD(a, b, out); err != nil {
    // Handle invalid slice lengths.
    panic(err)
}
// out is [5 7 9].
```

The import is a declaration and the statements belong inside your function;
the complete runnable file linked above shows the required layout. Select
`AddScalar`, `AddSIMDVectorTail`, or `AddSIMDParallel` explicitly to explore
those strategies. No automatic best-strategy dispatcher is implemented.

## Compute expression

```go
if err := computation.ComputeHeavySIMD(a, b, out); err != nil {
    panic(err)
}
// With a=[1 2 3], b=[4 5 6], out is [21 39 63].
```

Import `internal/computation` in the same module, as the full example does.
Use `ComputeHeavyScalar` for the scalar reference. These functions implement the
fixed expression; there is no arbitrary-expression parser or callback workload
adapter in the CLI.

## Exact in-place output

```go
if err := addition.AddSIMD(a, b, a); err != nil {
    panic(err)
}
```

Output now replaces `a`. Exact aliasing with `b` is also supported. Repeatedly
calling this in a hand-written benchmark changes the next iteration's inputs;
the built-in lab uses a disjoint reusable output to keep the input fixture
unchanged. Preserve that distinction when comparing measurements.

Different-start overlapping ranges, such as using `buffer[:n]` as input and
`buffer[1:n+1]` as output, are unsupported even when lengths match.

## Handle a length mismatch

```go
out := []float32{123}
err := addition.AddSIMD([]float32{1, 2}, []float32{3, 4}, out)
if errors.Is(err, addition.ErrLengthMismatch) {
    // out is still [123].
}
```

Import the standard `errors` package. Use
`computation.ErrLengthMismatch` for computation calls; the two packages have
distinct sentinels. Zero-length equal slices succeed, including nil slices.

## Fused affine

`patterns.AffineFMAScalar` and `patterns.AffineFMASIMD` round alpha*x+beta once
to binary32. Signatures and array alias/length rules match existing affine;
old affine/dot APIs remain unfused. Run `go run ./examples/fma` and inspect its
actual vector/software path. See [the FMA contract](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding).

## Bring your own slices carefully

Use `go run ./cmd/simdbench -list` to see implemented descriptors before choosing
a kernel. The catalogue and typed bindings are internal extension points, not a
public plugin API. F004 adds configured scalar-parameter bindings for float32
arrays; run `go run ./examples/patterns` and follow the
[broadcast/predicate guide](https://github.com/akolybelnikov/go-simd/wiki/13-Broadcast-and-predicate-experiments).
Sum adds `reduction.SumScalar(x)` and `reduction.SumSIMD(x)`, returning float32
directly without an output slice or error sentinel. Run `go run ./examples/sum`
and follow [its fold and rounding contract](https://github.com/akolybelnikov/go-simd/wiki/14-Sum-and-reduction-order).
Additional reductions and output types need their own explicit contracts.

SumSIMD2/4 and DotSIMD2/4 add two/four-accumulator strategies with the same
signatures/error contracts as their one-accumulator counterparts. Run
`go run ./examples/accumulators` and follow
[the exact assignment/merge order](https://github.com/akolybelnikov/go-simd/wiki/16-Accumulator-count-experiments).

Dot now adds `reduction.DotScalar(x,y)` and `reduction.DotSIMD(x,y)`, returning
`(float32,error)`. Handle `reduction.ErrLengthMismatch` for unequal lengths;
no output array is used. Input aliases/overlaps are read-only and supported.
Run `go run ./examples/dot` and follow
[dot and unfused rounding](https://github.com/akolybelnikov/go-simd/wiki/15-Dot-product-and-unfused-rounding).

The functions accept caller-supplied `[]float32` inside the module, so you can
write a local program that prepares different values and checks results. The
report CLI itself still uses generated fixtures; it has no CSV/JSON input-file
flag. Preparing data in a local example does not make it part of the built-in
measurement/export workflow automatically.

Define the correctness rule before measuring unusual floating-point values.
Compare rounding, NaNs, infinities, signed zero, and arithmetic semantics
deliberately; equality on a few simple integers is not a full floating-point
contract. Keep data preparation and complete-output validation outside the
timed inner loop when your objective is kernel-only timing.
