# F004, slice 7 — Exact float32 copy and data-movement baseline

Historical checkpoint record. The final [F004 acceptance record](F004-completion.md)
now closes the agreed scope after clamp; counts and stop decisions below describe
this earlier slice.

Implemented 2026-11-04 after packed predicates. The copy group adds Scalar,
GoCopy, SIMDPartial and SIMDTail: fourteen operations/thirty-eight implementations.
The question is when explicit SIMD helps copying and how it compares with Go's
compiler/runtime-selected copy. Stop after this movement checkpoint; clamp,
reordering, three-array MulAdd, standalone conversions, other input types and
architecture M4 remain separate work.

Retained [normal-SDK capture](evidence/copy-vector.measurements.json),
[SDK-emulated capture](evidence/copy-sdk-emulated.measurements.json) and
[commands/provenance](evidence/README.md) preserve observed exact-copy results.
These fixed-iteration captures verify plumbing without performance claims.

Implementation: signed commit [3080a89](https://github.com/akolybelnikov/go-simd/commit/3080a89f5c2a75efac9ed7fd73c67d8f806f0b83).

## Contract and controls

Operation copy-f32; benchmark BenchmarkCopy; arithmetic float32-copy-bits-v1.
One float32 input x, one caller-provided float32 output and no scalar parameters.
The four internal/movement functions accept `(x,out []float32) error`.

| Function | Label / strategy | Implementation |
| --- | --- | --- |
| CopyScalar | Scalar / scalar | Ordinary source loop assigning each element |
| CopyBuiltin | GoCopy / go-copy | Length-checked Go built-in copy(out,x) |
| CopySIMD | SIMDPartial / simd-partial | Partial-vector loads/stores, including the last short vector |
| CopySIMDVectorTail | SIMDTail / simd-scalar-tail | Full vector loads/stores followed by scalar assignments |

Every input bit is preserved: signed zeros, subnormals, infinities and quiet or
signaling NaN payloads are copied without arithmetic. Equal lengths are required;
a mismatch returns movement.ErrLengthMismatch before writes. Equal empty/nil
slices succeed. Disjoint storage and exact in-place aliases are supported.
Partial overlap is outside the shared contract, even though the Go built-in
itself has broader overlap behavior. This wrapper also rejects unequal lengths
rather than silently copying the shorter length.

The measured fixture uses disjoint arrays. Exact-alias timing could favor
different fast paths and is not mixed into this series. Scalar is a source-loop
label, not proof of scalar machine instructions. GoCopy delegates optimization
to the selected compiler/runtime; its label makes no fixed ISA/width claim.

## Preparation and measurement

A typed BindUnary seam prepares direct calls to the four kernels. No unused
second input or scalar parameter is fabricated. Dedicated compiled, native and
quick array callbacks retain the complete output slice after stopping timing;
existing numerical/predicate callback bodies retain their measurement boundaries.

Validation, call overhead, loop control, loads/stores and tail handling are timed.
Fixture generation, allocation, binding, input fingerprinting, output retention
and full-output hashing are outside timing. Every iteration performs a complete
kernel call; reported iterations follow Go's calibration rules. Current kernels
allocate zero per call. There is no manual warmup or timing threshold.

Logical work is `8*n` bytes: one float32 read plus one float32 write per element.
It excludes cache/coherency/runtime traffic and is not a measured DRAM counter.
At n=16,777,216 the two arrays occupy 128 MiB together, excluding runtime overhead.
The standard powers-of-two/native-width boundary sweep is unchanged.

Reports retain the Scalar source loop as the timing baseline, policy
copy-source-loop/v1. GoCopy is an additional optimized control; inspect its ns/op
alongside SIMD rather than interpreting a source-loop gain as beating Go copy.
Existing percentage columns remain changes versus Scalar. Data movement and
arithmetic are separate operations; do not subtract timings or equate working
sets to infer a universal memory bottleneck.

## Fixed raw-bit input and evidence

Generator generated-copy-float32-bits-v1 repeats these sixteen uint32 bit values:

```text
00000000 80000000 3f800000 bf800000 7f800000 ff800000 7fc12345 7f812345
00000001 80000001 007fffff 00800000 7f7fffff ff7fffff 3eaaaaab deadbeef
```

Only x and output are allocated. Existing operation-specific parameter/fixture
options do not alter copy input or case parameters. No new flag is added.
Input identity hashes x's float32 bits in little-endian index order.

Case.Copy records BitPolicyID=float32-exact-bits/v1 and actual process SDK
NativeLanes for every strategy. This is the SDK float32 context, not an observed
instruction width for Scalar or GoCopy. Different widths/bit policies cannot pool.
ArrayResultSHA256 hashes every output float32's bits, little endian in index
order, after each reported sample. For this exact copy it equals the input hash.
It is distinct from a byte-mask digest or the fused-affine last-element field;
older unrelated v1 captures omit the new optional fields.

Harnesses are standalone-copy/v1 and legacy-report-copy/v1. Native output retains
exact positive integer simd-lanes and eight uint32 output-sha256-0..7 words,
using big-endian word order to encode the digest bytes. The importer keeps
actual child facts and rejects missing/fractional/invalid width or digest metrics.
Only fresh generated-input imports establish this fixture hash; historical input
fingerprints remain unknown. Expected output never substitutes for child output.

GODEBUG=simd=0 exercises portable SDK emulation in a new process. It does not
disable optimizations in Go's independently selected built-in copy. Preserve
machine emulation, SDK lanes, compiler/runtime and harness context when comparing.

## Commands and acceptance

```powershell
go run ./cmd/simdreport -list -group copy
go run ./cmd/simdreport -group copy -bench '^BenchmarkCopy$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group copy -bench '^BenchmarkCopy$/^n=(1024|65536|1048576|16777216)$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -backend go-test -group copy -bench '^BenchmarkCopy$/^n=(8|16)$' -count 2 -benchtime 3x -format md
go run ./examples/copy
```

Installed CLI and Make use the same flags. Select larger working sets explicitly;
there is no arbitrary-size or alias-mode flag.

Tests compare exact uint32 bits over special values and randomized raw input,
at empty/native/tail/large nonmultiple lengths and offset subslices. They verify
both disjoint output and exact alias, unchanged inputs, output canaries, strict
length failures before writes and zero per-call allocations. Typed bindings,
single-input preparation/streaming fingerprint identity, irrelevant options,
strict native facts, mixed-context rejection, timing restoration and partial
capture retention are covered. All four labels pass compiled/native case/output
parity, including larger inputs and SDK emulation. Export tests include GoCopy
in all formats; compact listing permits one row per operation. All-group/no-Go
standalone execution, pure packages, full tests/vet, checked Go/installed/Make
examples and Windows/Linux CI complete this checkpoint.
