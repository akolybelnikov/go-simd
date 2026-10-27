# 08 — Using the kernels in Go

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

## Bring your own slices carefully

Use `go run ./cmd/simdbench -list` to see implemented descriptors before choosing
a kernel. The catalogue and typed bindings are internal extension points, not a
public plugin API. Broadcast parameters, scalar-returning reductions and new
output types still need explicit bindings and correctness contracts.

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
