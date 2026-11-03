# F004, slice 2c — Multiple reduction accumulators

Historical checkpoint record. The final [F004 acceptance record](F004-completion.md)
now closes the agreed scope after clamp; counts and stop decisions below describe
this earlier slice.

Implemented 2026-11-03 after unfused dot. Sum and dot each add SIMDStriped2 and
SIMDStriped4 alongside Scalar and the unchanged one-accumulator SIMDStriped.
The catalogue contains nine operations and twenty-four implementations.
Stop after this checkpoint; fused arithmetic and M4 remain separate phases.

Signed code checkpoint:
[3e7778c](https://github.com/akolybelnikov/go-simd/commit/3e7778c18b88aca48b67d189ff1b7baa140dcc42).
Retained [sum](evidence/accumulators-sum.measurements.json) and
[dot](evidence/accumulators-dot.measurements.json) smoke captures include
[provenance and exact reproduction commands](evidence/README.md).

## Experiment and exact order

Measure the effect of two or four independent accumulation chains in a complete
reduction. The new kernels unroll two/four full vectors per main loop. Results
include that loop-control change, register allocation/spills chosen by Go, merge,
extraction, final fold and tail. This is not an isolated instruction-latency or
physical-register-count measurement; more accumulators need not be faster.

For k=2 or 4, native float32 width w, full-vector ordinal v and lane j:

1. Initialize k vector accumulators to +0. Assign vector v to accumulator v%k.
2. Sum adds x[v*w+j]. Dot rounds x[v*w+j]*y[v*w+j] to float32 before adding it.
   Each accumulator receives its assigned vectors in increasing input order.
3. Full vectors after the last unrolled block go to accumulator 0, then 1, then
   2 as applicable. No input is padded or skipped.
4. Merge lane-wise left to right: `(a0+a1)` or `((a0+a1)+a2)+a3`, with one
   float32 rounding per addition. This is explicitly not a balanced merge tree.
5. Store merged lanes; fold low to high from +0. Process the remaining scalar
   input tail left to right, rounding dot products before addition.

Public functions are SumSIMD2/4(x) float32 and DotSIMD2/4(x,y) (float32,error).
Existing empty, signed-zero, NaN/Inf and immutable-input contracts apply. Unequal
dot lengths return +0,ErrLengthMismatch before reading either input. Read-only
dot overlaps are supported. No output array, FMA or parallel worker is introduced.
Current widths use stack scratch and allocate zero per call; widths beyond 16
lanes use dynamic scratch whose allocation is measured. Scalar and one-accumulator
kernel bodies, callback shapes, arithmetic/harness IDs and old fold IDs remain
unchanged. New variants use those existing typed callback/harness boundaries.

## Rounding and numerical acceptance

Changing accumulator count changes arithmetic order. With n=8*w, all other x
entries zero and dot y all ones, these examples have exact mathematical result 1:

| Nonzero x positions | Scalar | SIMDStriped | SIMDStriped2 | SIMDStriped4 |
| --- | --- | --- | --- | --- |
| `x[0]=2^24, x[w]=1, x[2*w]=-2^24` | 0 | 0 | 1 | 0 |
| `x[0]=2^24, x[2*w]=1, x[4*w]=-2^24` | 0 | 0 | 0 | 1 |

These are examples of order dependence, not an accuracy ranking. Special values
and overflow can produce different exceptional classes. NaN payloads remain
unspecified. Tests use an independent software assignment/merge oracle with
exact-rational, individually rounded products/additions. Finite normal-range
random cases also check the absolute forward-error bound gamma(m)*sum(abs(x))
for sum, or gamma(m)*sum(abs(exact x*y)) for dot, with u=2^-24,
gamma(m)=m*u/(1-m*u), conservative m=n+w+k+1 (sum) or n+w+k+2 (dot).
It applies only when m*u<1 and no overflow/underflow invalidates the relative
rounding assumptions. Exceptional/subnormal cases use the exact-order oracle.
There is no universal relative tolerance near cancellation or timing threshold.

## Catalogue, evidence and compatibility

| Label | Strategy ID | VectorAccumulators |
| --- | --- | --- |
| SIMDStriped2 | simd-striped-2 | 2 |
| SIMDStriped4 | simd-striped-4 | 4 |

Sum OrderIDs are striped-2-accumulators-lane-fold-f32/v1 and
striped-4-accumulators-lane-fold-f32/v1. Dot OrderIDs are
striped-2-accumulators-unfused-dot-f32/v1 and
striped-4-accumulators-unfused-dot-f32/v1. TailIDs remain scalar-after-lanes/v1
for sum and scalar-products-after-lanes/v1 for dot. VectorAccumulators describes
the algorithm; it is not an observation of physical register allocation.
NativeLanes retains actual process width for every strategy, including scalar.

Existing v1 Case.Reduction and ScalarResultBits retain these facts/results;
native labels resolve to the same registered fold identity, with required
result-bits and simd-lanes metrics. No new schema or input flag is needed.
Accumulator count/order distinguish strategies within a comparison; mixed widths,
operations or harnesses remain incompatible. Table order is Scalar, SIMDStriped,
SIMDStriped2, SIMDStriped4. Scalar remains the timing baseline.

Fixtures and fingerprints are unchanged. Both CLIs and Go/Make/installed
launchers select strategies using existing -bench names. Logical work remains
4*n+4 bytes for sum and 8*n+4 for dot; merge/scratch operations are timed but do
not change declared input/output byte units or establish measured DRAM traffic.

```powershell
go run ./cmd/simdreport -list -group dot
go run ./cmd/simdreport -group dot -dot-input mixed-magnitude -bench '^BenchmarkDot$/^n=1024$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group sum -bench '^BenchmarkSum$/^n=1024$/^(Scalar|SIMDStriped2|SIMDStriped4)$' -count 3 -benchtime 100ms -format csv
go run ./examples/accumulators
```

## Acceptance and stopping point

- Independent assignment/product/merge/fold oracles across zero, vector and
  unroll boundaries, remaining full vectors, scalar tails and large nonmultiples.
- Random, cancellation, signed zero, subnormal, NaN/Inf/overflow, input
  immutability, dot length failures/read-only overlap and zero current allocations.
- Vector/tail nonfusion checks; both accumulator-count counterexamples;
  binding tests that distinguish the actual algorithms from declared labels.
- Exact fold/count identities, stable old IDs, native import and four-column
  ordering, mixed-width rejection, compiled/native numerical/context parity,
  exports and standalone binary execution without Go/source.
- Full tests/vet, pure packages, forced emulation, runnable examples, actual
  Make/CLI workflows, retained smoke evidence and consistent README/Wiki guides.

FMA, clamp, predicate export, movement and new types remain future work. This
checkpoint does not activate M4 workers or manifests.
