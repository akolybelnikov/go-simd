# F004, slice 3 — Fused binary32 affine

Historical checkpoint record. The final [F004 acceptance record](F004-completion.md)
now closes the agreed scope after clamp; counts and stop decisions below describe
this earlier slice.

Implemented 2026-11-03 after accumulator comparisons. Group affinefma adds
Scalar and SIMDPartial for one-rounding alpha*x+beta. The catalogue now has ten
operations and twenty-six implementations. Existing affine and dot remain
unfused; the legacy computation source/contract remains unchanged. Fused dot and three-array MulAdd are separate
future contracts. Stop after this checkpoint, without starting M4.

Signed implementation:
[ded45bb](https://github.com/akolybelnikov/go-simd/commit/ded45bb5ace0fb0bcad0c8715a116cdd51f885da).
Retained [vector](evidence/fma-vector.measurements.json) and
[software](evidence/fma-software.measurements.json) captures have
[exact reproduction/provenance notes](evidence/README.md).

## Numerical contract and software reference

Operation affine-fma-f32, arithmetic float32-affine-fused-rne-v1, benchmark
BenchmarkAffineFMA. One x float32 input, float32 alpha/beta parameters, one
caller-owned float32 output. Compute the exact product plus beta, then round
once to binary32, round-to-nearest ties-to-even, including gradual underflow
and overflow. NaN payloads are unspecified. IEEE invalid products/sums yield
NaN; infinities follow IEEE classification. Exact opposite-sign cancellation
returns +0; a negative-zero product plus negative-zero beta returns -0.

Widening math.FMA to float64 and narrowing is not a universal float32 oracle:
the intermediate binary64 rounding can land on a binary32 midpoint. Tests
include explicit failures at normal, subnormal and overflow boundaries.

internal/fma32.FMA uses an exact binary64 product (at most 48 significant bits),
error-free TwoSum for product+beta, and a round-to-odd intermediate before the
final binary32 conversion. If the nearest binary64 sum is even and its residual
is nonzero, choose the adjacent odd binary64 value toward the exact sum. All
finite binary32 products/residuals fit the binary64 exponent range, so the wider
arithmetic does not overflow or underflow. Explicit conversions fix intermediate
rounding. Special nonfinite cases use math.FMA only for IEEE classification/signs.
No big-number arithmetic, allocation or runtime Go tool is used in the kernel.

Independent tests instead multiply/add exact big.Rat values and round once to
float32, with explicit exceptional/signed-zero rules. They cover 30,000 seeded
random bit triples and their sign inversions, special-value combinations and
targeted midpoint failures. This implementation is tested, not formally verified.
Background: [Boldo and Melquiond, rounding-to-odd FMA/sum algorithms](https://guillaume.melquiond.fr/doc/08-tc.pdf)
and the [Flocq rounding-to-odd properties](https://flocq.gitlabpages.inria.fr/flocq/html/Flocq.Prop.Round_odd.html).
Their work explains avoiding double rounding; we specialize the wider-format
approach to binary32 rather than claiming their formal proof verifies this Go code.

## SDK gate and actual execution paths

The installed Go 1.27.3 SDK's Float32s.MulAdd documentation says element-wise
x*y+z, without promising fused rounding. Its simd_emulated.go implementation
uses ordinary x.get(i)*y.get(i)+z.get(i). AMD64 compiler intrinsics/lowering include
VFMADD operations. A method name alone therefore cannot establish our contract.

At process initialization, emulated SIMD selects software-fma32/v1. Otherwise
a bounded vector probe checks cancellation, normal/subnormal/overflow midpoint
cases and negative zero in every lane against the software reference. Passing
selects vector-fma32/v1; failing selects software-fma32/v1. This startup probe is
outside measurements and benchmark-fixture preparation. It is a regression gate
backed by the selected SDK inspection and numerical suite, not an exhaustive proof
over all triples or a hardware acceleration/performance guarantee.
The shared internal/fmacap module owns this gate; platform metadata does not
depend on operation kernels.

Scalar always runs the corrected software reference. SIMDPartial uses partial
loads/stores and per-call alpha/beta broadcasts on the verified vector path.
On the software path it runs the same corrected scalar primitive per valid
element; it never silently substitutes the SDK's unfused emulation. Detailed
listing, quick titles, machine metadata and complete case identities expose the
chosen path. GODEBUG=simd=0 requires a new process and exercises this explicit
software path. No backend silently invokes Go or changes from compiled to native.

## Input, timing and array contract

AffineFMAScalar / AffineFMASIMD accept (x,out []float32,alpha,beta float32) error.
Length mismatch returns patterns.ErrLengthMismatch before writing output. Equal
empty/nil slices succeed. Output can be disjoint or exactly alias x; different-
start partial overlap is unsupported. Valid output bits are checked at vector/
tail boundaries with store canaries; these value tests do not prove SDK memory
access behavior. Inactive partial lanes may evaluate arithmetic but are not stored.

Length checks, cached-path dispatch, software correction or vector arithmetic,
per-call broadcasts when applicable and output writes are timed. The original
array callback/last-output sink shape is retained, including its timed final read.
New evidence bit extraction/native metric encoding happens after measurement.
Allocation tests report zero per call. Startup gate and fixture/hash preparation
are outside timing. This compares complete implementations with a software
correctness baseline, not isolated instruction costs or interchangeable fused/
unfused arithmetic. No performance thresholds are imposed.

Both CLIs share existing alpha/beta and new -fma-input (default balanced):

| Fixture | x recipe |
| --- | --- |
| balanced | float32(i%33-16)/4 |
| rounding | Repeat [1-2^-23,1+2^-23,1,-1,smallest positive subnormal,max finite,-0,1-2^-23] |
| mixed-magnitude | Repeat float32 [1e-20,1,1e20,-1e20,-1,3,-3,1e-20] |

Generator ID is generated-affine-fma-float32-v1/fixture plus x-bit SHA-256;
alpha/beta bits are separate exact parameters. Other fixture flags do not change
this identity. Only x and output are allocated; broadcasts are not extra input
arrays. Logical work is 8*n bytes, excluding intermediate/scratch traffic and
not measuring DRAM transactions. Historical options with no FMAInput normalize
to balanced without changing existing operation fingerprints or identities.

## Evidence and compatibility

Harnesses are standalone-affine-fma/v1 and legacy-report-affine-fma/v1. Baseline
policy is fused-software-reference/v1. Additive v1 Case.FMA contains:

| Field | Meaning |
| --- | --- |
| RoundingID | ieee754-binary32-rne-once/v1 |
| PathID | Actual strategy path: software-fma32/v1 or vector-fma32/v1 |
| VectorPathID | Process vector gate result, recorded for scalar too |
| NativeLanes | Actual measuring-process width |

Rounding, VectorPathID and width are comparison context; strategy PathID differs
legitimately between software Scalar and vector SIMDPartial. Conflicting process
paths/widths cannot pool. Case.FMA is omitted from old unrelated records.
machine.fma32_path exposes the process gate. Observation.ArrayResultLastBits is
optional uint32 for the final output element, including NaN/Inf bits; it is not
a whole-array checksum and is never averaged with timing statistics.

Native reports require exact fma-vector/fma-path (0/1), simd-lanes (positive
integer) and last-result-bits (uint32) metrics reported by the child. The importer
checks strategy/path consistency rather than using parent capability metadata.
Missing, fractional or out-of-range metrics fail. Old text cannot invent FMA
facts. Both backends use the same fixtures and selected exact native leaves.

```powershell
go run ./cmd/simdreport -list -group affinefma
go run ./cmd/simdreport -group affinefma -fma-input rounding '-alpha=1.0000001192092896' '-beta=-1' -bench '^BenchmarkAffineFMA$/^n=8$' -count 3 -benchtime 100ms -format md
go run ./examples/fma
```

## Acceptance and stopping point

- Independent rational scalar oracle, random/special triples, double-rounding
  counterexamples at all three boundaries, signed-zero and nonfinite behavior.
- Fused/unfused distinguishing cases, vector/tail sizes, exact in-place output,
  canaries, length failures, unchanged inputs and zero per-call allocations.
- Verified vector path and forced-emulation software path, with exact result
  bits, complete compiled/native case parity and strict child metrics.
- All ten groups in exports and standalone no-Go/source binary checks; invalid
  fixture options leave no artifacts and nonfinite results stay valid JSON.
- Full tests/vet, pure packages (including fma32), emulation, CLI/Make examples,
  retained vector/software smoke evidence and synchronized README/Wiki guides.

Stop after fused affine. Fused reductions, three-array MulAdd, clamp, predicate
export, new types and M4 require separate work.
