# F004, slice 4 — Fused dot with specified fold orders

Implemented 2026-11-03 after fused affine. Group dotfma adds Scalar,
SIMDStriped, SIMDStriped2 and SIMDStriped4. The catalogue contains eleven
operations and thirty implementations. Existing dot, affine, fused-affine and
accumulator kernels remain unchanged. Stop after this checkpoint; three-array
MulAdd, other operation families and architecture M4 remain separate work.

Implementation: signed commit [9d96687](https://github.com/akolybelnikov/go-simd/commit/9d966877f9c8776c974c9a9c4edbc35c4627898d).
Retained [hardware capture](evidence/dotfma-vector.measurements.json),
[software-striped capture](evidence/dotfma-software.measurements.json) and
[commands/provenance](evidence/README.md) accompany this checkpoint. They are
tiny fixed-iteration plumbing observations, not performance rankings.

## Per-update arithmetic, assignment and final fold

Operation dot-fma-f32, benchmark BenchmarkDotFMA, arithmetic
float32-dot-fused-ordered-folds-v1. Inputs are x/y float32 slices; output is a
returned float32 plus error. Each update rounds exact x*y+accumulator once to
binary32, nearest ties-to-even. This is not one exact sum rounded at the end.

- Scalar: start at +0 and apply FMA32 left to right using the corrected software
  primitive from the fused-affine checkpoint.
- SIMDStriped/2/4: assign full-vector ordinal v to accumulator v modulo k, for
  k=1/2/4; process each accumulator's vectors in input order with fused updates.
- Native loops unroll k full vectors. Remaining full vectors go to accumulators
  0,1,2 as applicable. No padded input lanes are loaded or reduced.
- Merge accumulators lane-wise left to right with separately rounded float32
  additions: a0, a0+a1, or ((a0+a1)+a2)+a3. Fold merged lanes low to high from +0.
- After the lane fold, apply software FMA32 to the remaining scalar tail in
  input order. The final fold does not fuse or reassociate its additions.
- On an unavailable vector gate, software striped execution retains the same
  assignment, merge and tail rules at the actual process width. It never
  substitutes the scalar order for a SIMD-labelled strategy.

DotFMAScalar, DotFMASIMD, DotFMASIMD2 and DotFMASIMD4 return (float32,error).
Unequal lengths return +0,ErrLengthMismatch before reading either input. Equal
empty/nil slices return +0,nil. Inputs remain unchanged; exact aliases and
different-start read-only overlaps are supported. No output array is allocated.
NaN payloads are unspecified; infinities, overflow, subnormals and signed zero
follow the specified operations. Cross-strategy bit equality is not required.

## Numerical examples and acceptance rule

For n=2*w, all other entries zero:

| Nonzero values | Scalar / SIMDStriped | SIMDStriped2/4 |
| --- | --- | --- |
| `x[0]=-1,y[0]=1,x[w]=1+2^-23,y[w]=1-2^-23` | -2^-46 | +0 |
| `x[0]=-max,y[0]=1,x[w]=max,y[w]=2` | max finite | +Inf |

The original unfused scalar dot returns +0 for the first layout and +Inf for
the second. Fusing updates can preserve information lost by a rounded product,
but placing products in different accumulators changes the rounded result again.
Neither fusion nor accumulator count establishes a universal accuracy ranking.

For n=w, x[i]=smallest positive subnormal and y[i]=-0.25, scalar fused updates
return -0. Full-vector lane results are also -0, but the low-to-high lane fold
from +0 returns +0. With a scalar tail, a later fused update can return -0 again.
This signed-zero distinction is part of the contract, not a tolerance exception.

Independent tests use exact-rational multiplication/addition with one binary32
rounding for each FMA, explicit special/zero rules and independently rounded
merge/fold additions. NaN classification is checked without fixing payloads.
Cases include vector/unroll/tail boundaries, random bits, finite random values,
underflow, exceptional values, cancellation and overflow rescue.
Eligible finite normal-range cases also check the conservative absolute bound
gamma(m)*sum(abs(exact x*y)), u=2^-24, gamma(m)=m*u/(1-m*u), m=n+w+k+1
(k=0 for scalar). This requires m*u<1 and no overflow/underflow invalidating the
relative-error assumptions. Exceptional/subnormal cases use the exact-order
oracle. No universal relative tolerance or benchmark timing threshold is used.

## Process gate versus actual case execution

Reuse the shared internal/fmacap gate and corrected internal/fma32 primitive.
The selected SDK's emulated MulAdd is not guaranteed fused; software execution
therefore performs fused scalar updates within each declared stripe. The startup
gate remains outside timing. A passing gate does not mean a short case executes
vector FMA: when n<w all updates are scalar-tail updates.

Case.FMA retains process VectorPathID/NativeLanes separately from actual PathID:

| Native fma-path code | PathID | Condition |
| --- | --- | --- |
| 0 | software-fma32/v1 | Scalar reference |
| 1 | vector-fma32/v1 | Vector gate passes; full vectors and no tail |
| 2 | software-striped-fma32/v1 | SIMD-labelled strategy with software gate |
| 3 | vector-fma32-with-software-tail/v1 | Vector gate passes; full vectors plus tail |
| 4 | software-tail-fma32/v1 | Vector gate passes; n<w |

These codes apply to fused dot; fused affine retains its prior 0/1 encoding.
RoundingID is binary32-fma-updates-add-merge-rne/v1. Case.Reduction stores:

| Strategy | OrderID | VectorAccumulators |
| --- | --- | --- |
| Scalar | left-to-right-fma-dot-f32/v1 | 0 |
| SIMDStriped | striped-lane-fma-dot-f32/v1 | 1 |
| SIMDStriped2 | striped-2-accumulators-fma-dot-f32/v1 | 2 |
| SIMDStriped4 | striped-4-accumulators-fma-dot-f32/v1 | 4 |

Scalar TailID is none; SIMD TailID is fma-products-after-lanes/v1. Counts are
algorithmic, not physical register measurements. Both FMA/fold identities retain
actual process lanes; process gates/widths cannot pool. Case paths may legitimately
change across sizes in one sweep, while strategy/order and input identity remain
explicit. Harnesses are standalone-dot-fma/v1 and legacy-report-dot-fma/v1;
baseline policy is specified-fused-float32-dot-folds/v1.

Reuse v1 ScalarResultBits for actual returned bits, including NaN/Inf safely.
Values are not averaged with timings. Native child fma-vector (0/1), fma-path
(integer 0..4), simd-lanes and result-bits metrics are mandatory. The importer
validates the reported case path against strategy, size and child width/gate.
Parent facts cannot stand in for child capabilities. Old operation IDs, fold/
rounding identities, strict affine metric rules and evidence remain unchanged.

## Fixtures, logical bytes and measurement boundary

Both dot and dotfma use -dot-input: balanced, cancellation, mixed-magnitude and
new fusion-sensitive. Existing recipes/fingerprints are unchanged. New recipe:
repeat 64 entries, x[0]=-1,y[0]=1; x[32]=1+2^-23,y[32]=1-2^-23; otherwise x=y=0.
At n=64 the fused scalar returns -2^-46 and the original unfused scalar returns
+0. Other fused strategies depend on actual width/assignment; four accumulators
can return +0 at width 16. Preserve those facts when sharing results.

Fused generator ID is generated-dot-fma-float32-v1/fixture; the two groups use
identical array values but retain separate generator/arithmetic identities.
Both arrays participate in SHA-256. Other operation-specific options do not
change fused-dot identity. No new flag is added. Logical work stays 8*n+4 bytes
per call, excluding scratch/intermediate traffic; it is not measured DRAM traffic.

Timing includes length checks, cached-path dispatch, all FMA updates, unrolling,
merge, extraction/fold and scalar tail. Typed checked callbacks retain the final
value once after stopping timing; errors preserve earlier successful samples.
Native reporting is isolated from existing reduction/pattern callbacks.
Current widths/counts allocate zero per call. Native extraction has 16 stack
slots; software stripes have 64. Larger widths may require dynamic scratch,
whose actual allocations appear in counters. This compares complete kernels,
including Go's register allocation/spills and software correction overhead.

```powershell
go run ./cmd/simdreport -list -group dotfma
go run ./cmd/simdreport -group dotfma -dot-input fusion-sensitive -bench '^BenchmarkDotFMA$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./examples/dotfma
```

Acceptance includes independent numerical oracles, exact binding/count checks,
allocation/input/length/overlap rules, path-code validation, partial captures,
four-fixture compiled/native parity across tail-only/full/hybrid cases, all-group
exports and no-Go/source standalone execution. Full tests/vet, pure packages,
normal/forced-emulation paths, CLI/Make checks and retained captures accompany
this checkpoint. Stop before three-array MulAdd or any architecture M4 work.
