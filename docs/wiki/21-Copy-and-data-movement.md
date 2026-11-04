# 21 — Copy data and compare Go's built-in control

Checkpoint history: [three-array MulAdd](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add) is now implemented after the
original F004 closure. The current catalogue has twenty-one operations/fifty-one
implementations; future/stop statements below describe this earlier checkpoint.

The copy group compares four ways of moving float32 data without arithmetic:
Scalar, GoCopy, SIMDPartial and SIMDTail. The copy checkpoint brought the catalogue to fourteen
operations/thirty-eight implementations. The [MulAdd extension](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add) now brings it to sixteen/forty-three.

This experiment asks when explicit vector copying helps, and how it compares
with Go's own copy implementation. It provides a movement reference alongside
the arithmetic and predicate experiments.

## Run a bounded comparison

```powershell
simdreport -list -group copy
simdreport -group copy -bench '^BenchmarkCopy$/^n=64$' -count 3 -benchtime 100ms -format md
simdbench -group copy
```

From a checkout with GOEXPERIMENT=simd, use go run ./cmd/simdreport or
go run ./cmd/simdbench. Make forwards the same flags:

```bash
make list ARGS="-group copy"
make report COUNT=3 BENCHTIME=100ms ARGS="-group copy -bench BenchmarkCopy/n=64 -format md"
make run ARGS="-group copy"
```

No new flag or parameter is needed. Quick comparison uses 1,048,576 elements.
Reports use the standard power-of-two and native-width boundary sweep.
Keep all four labels when comparing the controls:

| Label | Source behavior |
| --- | --- |
| Scalar | Assign each element in an ordinary Go loop |
| GoCopy | Call Go's built-in copy after the common length check |
| SIMDPartial | Partial-vector load/store throughout, including the final short vector |
| SIMDTail | Full vectors, then scalar assignments for remaining elements |

Scalar describes source structure. The compiler may optimize it, and GoCopy
uses compiler/runtime-selected execution. Neither label establishes a particular
machine instruction or vector width. Both are useful controls for explicit SIMD.

The table's percentages remain relative to Scalar. Read GoCopy ns/op as the
additional optimized control; a SIMD improvement over the loop does not by itself
establish improvement over Go copy.

## Exact bits and storage rules

All four kernels preserve every float32 bit, including signed zero, subnormals,
infinities and NaN payloads. They take (x,out []float32) error in
internal/movement:

| Function | Label |
| --- | --- |
| CopyScalar | Scalar |
| CopyBuiltin | GoCopy |
| CopySIMD | SIMDPartial |
| CopySIMDVectorTail | SIMDTail |

Lengths must agree. Wrong lengths return movement.ErrLengthMismatch before any
write; equal empty/nil succeeds. Output is caller-provided, using disjoint
storage or an exact in-place alias. Partial overlap is outside the shared
experiment contract. The Go built-in's broader overlap/short-length behavior
does not change the wrapper's common rules.

Benchmarks always use disjoint arrays. Exact-alias correctness is tested but
has different potential fast paths and is not mixed into the timings.

## What is timed and how much data moves

Validation, call overhead, loop control, reads/writes and tails are timed.
Fixture allocation/filling, binding, input fingerprinting, output retention and
complete output hashing are outside timing. The complete slice is retained after
measurement; a checksum is not added to every copy iteration. Current kernels
allocate zero per call. Go controls calibration and GC.

Logical work is 8*n bytes: four read and four written per element. This is
effective work accounting, not a DRAM or cache-traffic measurement. At the
largest registered n=16,777,216, input and output together occupy 128 MiB.

For a focused larger-working-set experiment:

```powershell
simdreport -group copy -bench '^BenchmarkCopy$/^n=(1024|65536|1048576|16777216)$' -count 3 -benchtime 100ms -format md
```

Results depend on the compiler/runtime, hardware, size and reuse/cache state.
Do not assume a SIMD win, infer request latency from batch averages or subtract
copy timings from arithmetic timings to prove a memory bottleneck.

## Fixture and complete output evidence

The fixed generated-copy-float32-bits-v1 fixture repeats sixteen raw bit values
covering zeros, ordinary finite values, infinities, quiet/signaling NaNs,
subnormals and extreme finite values. No second input or scalar parameter is
allocated. Other operation-specific flags leave this input/case unchanged.

Case.Copy records float32-exact-bits/v1 and actual process SDK NativeLanes for
all four labels. The lane count is SDK context; it does not describe the
compiler/runtime copy's physical instruction width. Machine metadata retains
SDK emulation separately.

ArrayResultSHA256 hashes the complete output's float32 bits in little-endian
index order after timing. It equals x's input hash for a correct copy and safely
retains evidence for NaNs/infinities without JSON float values. It is neither
the raw array nor the last-element field used by fused affine.

```powershell
$path = Get-ChildItem ./artifacts/copy/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
foreach ($row in $capture.run.Observations) {
    $row.Label
    $row.Key.Case.Copy
    $row.ArrayResultSHA256 -eq $row.Key.Case.Input.SHA256
}
```

Harnesses are standalone-copy/v1 and legacy-report-copy/v1. Add -backend go-test
for a developer comparison requiring Go/source. Child SDK width and eight exact
digest-word metrics establish native facts. Input fingerprints are established
only for fresh generated fixtures; historical unknown input stays unknown.

In a new process, GODEBUG=simd=0 exercises portable SDK emulation. GoCopy's
compiler/runtime optimization remains independent, so this setting does not
mean all four implementations use scalar machine instructions.

## Run the bit-preservation example

```powershell
go run ./examples/copy
```

```text
Scalar: [80000000 7f812345 7fc12345 00000001 3f800000]
GoCopy: [80000000 7f812345 7fc12345 00000001 3f800000]
SIMDPartial: [80000000 7f812345 7fc12345 00000001 3f800000]
SIMDTail: [80000000 7f812345 7fc12345 00000001 3f800000]
exact alias preserved first bits: 80000000
```

The [example](https://github.com/akolybelnikov/go-simd/blob/main/examples/copy/main.go)
prints uint32 bits rather than rounding or formatting NaNs. Read the
[kernels](https://github.com/akolybelnikov/go-simd/blob/main/internal/movement/copy.go)
and [contract/evidence record](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-copy.md).
This checkpoint stops after copying. Clamp, reordering, three-array MulAdd,
standalone conversions, other input types and M4 remain separate work.

The implementation record links the signed code and retained normal-SDK and
SDK-emulated captures with reproduction commands and source provenance. Tiny
fixed-iteration captures verify plumbing without performance claims.
