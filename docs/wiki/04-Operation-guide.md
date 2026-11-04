# 04 — Operation and kernel guide

The lab contains seventeen operations and forty-five implementations: sixteen operations on `float32`
inputs and one float64 addition operation. Float32 inputs provide ten arithmetic/selection/clipping float32 outputs, one exact-copy output, two uint8 predicate representations and three reductions. Array kernels have
one, two or three declared inputs, scalar parameters where needed and a caller-provided
output slice of their declared type. Predicate export writes canonical uint8 0/1.
Sum accepts one input and returns one float32; dot/dotfma accept two
inputs and returns a float32 plus a length-validation error.

## Complete kernel reference

| Package | Function | Benchmark label | What it does |
| --- | --- | --- | --- |
| `internal/addition` | `Add64Scalar` / `Add64SIMD(a,b,out []float64) error` | Scalar / SIMDPartial | Binary64 addition, true arrays, bounded actual-width loads/stores |
| `internal/addition` | `AddScalar(a, b, out []float32) error` | `Scalar` | Ordinary loop computing `a[i] + b[i]` |
| `internal/addition` | `AddSIMD(a, b, out []float32) error` | `SIMDPartial` | Partial-vector loads/stores throughout the slice |
| `internal/addition` | `AddSIMDVectorTail(a, b, out []float32) error` | `SIMDTail` | Full SIMD vectors followed by a scalar remainder |
| `internal/addition` | `AddSIMDParallel(a, b, out []float32) error` | `SIMDParallel` | Contiguous worker chunks, each using full vectors and a scalar tail |
| `internal/computation` | `ComputeHeavyScalar(a, b, out []float32) error` | `Scalar` | Ordinary loop computing `a*a + b*b + a*b` |
| `internal/computation` | `ComputeHeavySIMD(a, b, out []float32) error` | `SIMD` | Partial-vector loads, explicit vector `Mul`/`Add`, partial stores |
| `internal/patterns` | `ScaleScalar` / `ScaleSIMD(x,out []float32,alpha float32) error` | Scalar / SIMDPartial | Broadcast scale |
| `internal/patterns` | `AffineScalar` / `AffineSIMD(x,out []float32,alpha,beta float32) error` | Scalar / SIMDPartial | Explicit unfused float32 multiply/add |
| `internal/patterns` | `AffineFMAScalar` / `AffineFMASIMD(x,out []float32,alpha,beta float32) error` | Scalar / SIMDPartial | One-rounding binary32 affine; actual software/vector path retained |
| `internal/patterns` | `ThresholdZeroScalar` / `ThresholdZeroSIMD(x,out []float32,t float32) error` | Scalar / SIMDPartial | Strict predicate, false to +0 |
| `internal/patterns` | `PredicateGTScalar` / `PredicateGTSIMD(x []float32,out []uint8,t float32) error` | Scalar / SIMDPartial | Ordered comparison exported as canonical byte 0/1; conversion/extraction/stores included |
| `internal/patterns` | `PredicateGTPackedScalar` / `PredicateGTPackedSIMD(x []float32,out []uint8,t float32) error` | Scalar / SIMDPartial | Ordered comparison packed LSB first into ceil(n/8) bytes; unused high bits zero; conversion/packing/stores included |
| `internal/patterns` | `SelectGTScalar` / `SelectGTSIMD(x,y,out []float32,t float32) error` | Scalar / SIMDPartial | Bit-preserving selection |
| `internal/patterns` | `RangeZeroScalar` / `RangeZeroSIMD(x,out []float32,lo,hi float32) error` | Scalar / SIMDPartial | Strict range masks combined with And |
| `internal/patterns` | `MulAddScalar` / `MulAddSIMD(a,b,c,out []float32) error` | Scalar / SIMDPartial | Three streamed arrays; one-rounding binary32 FMA; actual software/vector path |
| `internal/patterns` | `ClampScalar` / `ClampSIMD` / `ClampSIMDMinMax(x,out []float32,lower,upper float32) error` | Scalar / SIMDPartial / SIMDMinMax | Ordered inclusive clipping; preserve original tie/NaN bits; Min/Max repairs timed |
| `internal/reduction` | `SumScalar(x []float32) float32` | Scalar | Left-to-right float32 sum from +0 |
| `internal/reduction` | `SumSIMD(x []float32) float32` | SIMDStriped | One vector accumulator, low-to-high lane fold, scalar tail |
| `internal/reduction` | `DotScalar(x,y []float32) (float32,error)` | Scalar | Rounded products, left-to-right float32 accumulation |
| `internal/reduction` | `DotSIMD(x,y []float32) (float32,error)` | SIMDStriped | Unfused vector products, one accumulator, lane fold/tail |
| `internal/reduction` | `SumSIMD2(x)` / `SumSIMD4(x)` returning float32 | SIMDStriped2 / SIMDStriped4 | Round-robin vector accumulation, left-to-right lane-wise merge/fold/tail |
| `internal/reduction` | `DotSIMD2(x,y)` / `DotSIMD4(x,y)` returning (float32,error) | SIMDStriped2 / SIMDStriped4 | Unfused products with 2/4 independent accumulators, merge/fold/tail |
| `internal/reduction` | `DotFMAScalar(x,y []float32) (float32,error)` | Scalar | Corrected binary32 FMA updates left to right |
| `internal/reduction` | `DotFMASIMD(x,y []float32) (float32,error)` | SIMDStriped | One accumulator; FMA updates, rounded lane fold, software FMA tail |
| `internal/reduction` | `DotFMASIMD2(x,y)` / `DotFMASIMD4(x,y)` returning (float32,error) | SIMDStriped2 / SIMDStriped4 | 2/4 accumulators; actual vector/software stripes, rounded merge/fold, software FMA tail |
| `internal/movement` | `CopyScalar(x,out []float32) error` | Scalar | Source-loop exact bit copy |
| `internal/movement` | `CopyBuiltin(x,out []float32) error` | GoCopy | Strict-length wrapper around Go copy; compiler/runtime-selected execution |
| `internal/movement` | `CopySIMD(x,out []float32) error` | SIMDPartial | Partial-vector loads/stores, preserving every bit |
| `internal/movement` | `CopySIMDVectorTail(x,out []float32) error` | SIMDTail | Full vectors followed by scalar assignments |

Read the [addition source](https://github.com/akolybelnikov/go-simd/blob/main/internal/addition/add.go)
and [computation source](https://github.com/akolybelnikov/go-simd/blob/main/internal/computation/compute.go).
These are internal experimental APIs; external modules cannot directly import
them. See [using the kernels](https://github.com/akolybelnikov/go-simd/wiki/08-Using-the-kernels)
for a module-local example.

## Scalar addition

For `a = [1, 2, 3]` and `b = [4, 5, 6]`, every addition kernel should produce
`[5, 7, 9]`. The scalar reference walks the slice one element at a time in source.
The benchmark still includes the function call, length check, and loop overhead;
it is not measuring an isolated arithmetic instruction.

## Partial vectors

`AddSIMD` uses `simd.LoadFloat32sPart` and `StorePart`. Each loop iteration
processes up to one vector's worth of valid values. The final short portion is
handled through the partial-vector mechanism instead of a separate scalar loop.

This gives a compact loop that handles arbitrary lengths. It does not guarantee
lower overhead or better performance than another strategy at every size.

## Full vectors with a scalar tail

`AddSIMDVectorTail` queries the current number of float32 lanes, processes full
vectors with full loads/stores, then adds remaining elements in a scalar loop.

If the width is eight lanes, a nine-element input uses one full vector and one
scalar element. A seven-element input uses only the scalar tail after setup.
Eight is an example width, not a universal project constant; use the report's
observed lane count for your machine/toolchain.

## Parallel addition

For nonempty input, worker count is:

```text
min(effective GOMAXPROCS, number of elements)
```

The input is divided into balanced contiguous chunks. Each call creates
goroutines, performs the work, and waits for them to finish. Startup, scheduling,
synchronization, and allocations are inside the timed call. There is no
persistent worker pool or payload threshold that bypasses concurrency.

Each chunk has its own vector boundary and tail. A whole input divisible by the
vector width can still yield scalar worker tails after partitioning. With small
chunks, a parallel "SIMD" strategy may do mostly or entirely scalar-tail work.

There is currently no parallel scalar control. The displayed comparison against
serial scalar combines concurrency and vector strategy effects. It cannot
attribute the full gain to SIMD.

## Compute expression

Both computation kernels evaluate:

```text
out[i] = a[i]*a[i] + b[i]*b[i] + a[i]*b[i]
```

For `[1, 2, 3]` and `[4, 5, 6]`, the result is `[21, 39, 63]`.
It has more arithmetic per element than addition while using the same two inputs
and output shape. "Compute-heavy" is the existing function name, not proof that
every measured size is compute-bound.

The computation group's scalar expression and explicit SIMD Mul/Add source
are unchanged. A separate
`affinefma` group now provides one-rounding affine with a corrected software
reference and gated vector/software paths; see [fused affine](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding). Floating-point rounding, overflow,
infinities, NaNs, and toolchain lowering need explicit semantics before expanding
that coverage. Do not claim bit-for-bit fused/unfused equivalence from ordinary
integer-valued test examples.

## Original addition/computation input/output contract

- All three slices must have equal lengths.
- A mismatch returns that package's `ErrLengthMismatch` before modifying output.
- Empty inputs, including three nil slices, succeed.
- Output can be disjoint or exactly alias either input.
- Partial overlapping ranges are unsupported, including overlapping subslices
  with different starting elements. Use disjoint storage or an exact alias.
- Kernels do not resize slices or allocate an output slice for you.

The two packages have separate error sentinels; use `errors.Is` with the sentinel
from the package whose operation you called.

## Original addition/computation data and accounting

The generated fixture is `generated-float32-v1`:

```go
a[i] = float32(i%100) / 10
b[i] = float32(i%50-25) / 10
```

The current benchmark allocates and fills the buffers before timing and reuses
the output. It reports `12 * elements` effective bytes per operation: two
float32 reads and one write. At 16,777,216 elements, each slice is 64 MiB and the
three slices total 192 MiB, excluding other runtime overhead.

New unary scale/affine/threshold/range account for `8*n`; two-array selection
accounts for `12*n`. Numerical, alias and fixture rules are in
[broadcast and predicate experiments](https://github.com/akolybelnikov/go-simd/wiki/13-Broadcast-and-predicate-experiments).

Predicate export requires disjoint input/output storage and returns a length
error before writes. Equal empty/nil slices succeed. Logical work is `5*n`;
input stays unchanged, and inactive tail lanes never escape. See
[canonical byte masks](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks) for full output evidence and examples.

Packed predicate output requires exactly `ceil(n/8)` bytes; invalid lengths fail
before writes. Input i maps to bit i%8 of byte i/8, with unused high bits zero.
All output bytes are overwritten. Logical work is `4*n+ceil(n/8)`.
See [packed predicates](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks) for decoding, boundaries and evidence.

Copy takes one input/output float32 array with equal lengths; mismatches return
movement.ErrLengthMismatch before writes. Empty succeeds. Disjoint output and
exact alias are supported; partial overlap is outside the shared contract.
Benchmarks use disjoint buffers, logical work `8*n`, and full output-bit digests
after timing. See [copy and data movement](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement) for compiler/runtime controls.

## Internal machine identifiers

Sum has no output slice or length-error sentinel. It leaves input unchanged,
returns +0 for empty input and rounds every addition to float32. Its specified
orders can return different values, including different exceptional classes.
See [sum and reduction order](https://github.com/akolybelnikov/go-simd/wiki/14-Sum-and-reduction-order)
for examples, input recipes, result retention and `4*n+4` logical byte accounting.

Dot rounds products before addition, propagates ErrLengthMismatch before reading
unequal-length inputs, and allocates no output slice. Empty equal inputs return
+0,nil; read-only input overlaps are allowed. See
[dot and unfused rounding](https://github.com/akolybelnikov/go-simd/wiki/15-Dot-product-and-unfused-rounding)
for its complete numerical/error contract and `8*n+4` accounting.

DotFMA functions use the same checked scalar shape and input/error/empty rules,
but fuse each multiply-accumulate. Ordinary merge/fold additions still round
separately. Software fallback preserves the exact specified stripes and merges.
See [fused dot](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders) for order-sensitive examples, `8*n+4` accounting and
per-case execution-path evidence.

The structured result model separates display names from identity:

| Display group / label | Internal identifier |
| --- | --- |
| addition | Operation `add` |
| computation | Operation `compute-expression` |
| Scalar | Strategy `scalar` |
| SIMDPartial / computation SIMD | Strategy `simd-partial` |
| SIMDTail | Strategy `simd-scalar-tail` |
| SIMDParallel | Strategy `simd-parallel-scalar-tail` |
| sum | Operation `sum-f32` |
| dot | Operation `dot-f32` |
| dotfma | Operation `dot-fma-f32` |
| predicate | Operation `predicate-gt-f32`; output encoding `uint8-zero-one/v1` |
| predicatepack | Operation `predicate-gt-packed-f32`; output encoding `uint8-packed-lsb0/v1` |
| copy / GoCopy | Operation `copy-f32`; strategy `go-copy` |
| SIMDStriped | Strategy `simd-striped` |
| SIMDStriped2 | Strategy `simd-striped-2` |
| SIMDStriped4 | Strategy `simd-striped-4` |

These identifiers are internal model fields, not new command flags. Select
current CLI cases with group names and benchmark labels from the registry.

M2 defines these identities once in
[`internal/catalog`](https://github.com/akolybelnikov/go-simd/tree/main/internal/catalog).
Quick comparisons, native benchmark registration and the legacy importer consume
that metadata. Run `go run ./cmd/simdbench -list` to inspect the current descriptor
and runtime context. The original six bindings preserve their kernel/call paths;
Twelve float32-output configured bindings cover broadcasts, predicates and fused affine;
four additional BindUint8 bindings export byte and packed predicates, and four
BindUnary bindings implement exact copying. Sum, dot and dotfma each have four scalar-returning
bindings; [the accumulator guide](https://github.com/akolybelnikov/go-simd/wiki/16-Accumulator-count-experiments)
specifies the 1/2/4-vector strategies and their numerical differences.

## Pattern gaps and reading credit

[SIMD and Go, by Jesús Espino](https://internals-for-interns.com/posts/go-simd/)
is credited as the reading prompt for our
[SIMD pattern review](https://github.com/akolybelnikov/go-simd/wiki/12-SIMD-patterns-and-reading).
That review distinguishes implemented broadcasts, predicates and accumulation
from implemented copy and implemented three-array FMA and planned reordering experiments and checks the portable/`archsimd` boundary.


## Clamp

[Clamp and numerical contracts](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts) explains inclusive bounds, exact NaN/zero behavior and timed Min/Max repairs. Run `go run ./examples/clamp`.
