# F004, slice 2b — Unfused float32 dot product

Implemented 2026-11-03 after ordered sum. This checkpoint adds dot product and
checked scalar-returning execution, then stops before multiple accumulators or
FMA. The catalogue contains nine operations and twenty implementations.

Signed code checkpoint:
[4847a8f](https://github.com/akolybelnikov/go-simd/commit/4847a8fbe8a42c5a47bb80bb109110e040b4db18).
The [retained dot smoke](evidence/dot-cancellation.measurements.json) and
[reproduction/provenance notes](evidence/README.md) accompany this record.

## Kernel and numerical contract

Group `dot`, operation `dot-f32`, benchmark `BenchmarkDot`, arithmetic
`float32-dot-unfused-ordered-folds-v1`. Inputs are x and y float32 slices;
`DotScalar(x,y)` and `DotSIMD(x,y)` return `(float32,error)`.

- Unequal lengths return +0 and `reduction.ErrLengthMismatch` before reading
  either input. Callers can use `errors.Is`; there is no output array.
- Equal empty/nil slices return +0,nil. Inputs remain unchanged; exact aliases
  and different-start read-only overlaps are supported.
- Each product rounds to float32 **before** addition; every addition also rounds
  to float32. Scalar explicit conversions and SIMD separate Mul/Add operations
  prohibit fusion under this contract.
- Scalar starts at +0 and adds rounded products left to right.
- SIMDStriped uses one zero-initialized native-width vector accumulator. Process
  full vectors in increasing input order, store/fold lanes low to high from +0,
  then add rounded scalar-tail products left to right. No padded input lanes.
- Validation, multiplication, extraction, final lane fold and tail are timed.
  Current widths use stack scratch and allocate zero per call; widths beyond
  16 lanes use dynamic scratch visible in counters.
- NaN/Inf/product overflow and underflow follow each specified float32 order.
  NaN payloads are unspecified. All-negative-zero products sum to +0 under the
  initial-value rule. Finite inputs may yield different exceptional classes
  under the two folds; scalar is a timing baseline, not an exact oracle.

With y all ones, dot has the same two cancellation counterexamples as sum. For
x length 2*w and other entries zero:

| Nonzero x entries | Scalar | SIMDStriped | Exact dot |
| --- | --- | --- | --- |
| x[0]=2^24, x[1]=1, x[w]=-2^24 | 0 | 1 | 1 |
| x[0]=2^24, x[1]=-2^24, x[w]=1 | 1 | 0 | 1 |

Fusion is a separate numerical change: x=[-1,1+2^-23], y=[1,1-2^-23]
returns +0 under rounded-product scalar dot. Fusing the second multiply/add
would retain -2^-46. Tests place the corresponding operations in the same
vector lane and across the scalar-tail boundary to check both paths.

Tests use exact-rational finite multiplication with one float32 rounding,
followed by the independent per-addition rounded fold oracle. IEEE special
values are handled explicitly; result bits agree except unspecified NaN payloads.
For finite normal-range random fixtures without overflow/underflow and k*u<1,
tests also check `abs(got-exactDot) <= gamma(k)*sum(abs(exact x*y))`,
u=2^-24, gamma(k)=k*u/(1-k*u), conservative k=n+w+2. This includes multiplication
and accumulation rounding. No universal bound is promised for the largest sweep,
exceptional inputs or near-zero relative error; there are no timing thresholds.

## Typed execution, identity and evidence

`BindCheckedScalar(x,y)` returns a typed `func()(float32,error)` closure.
The engine's ReduceChecked path, quick callback and native wrapper propagate
errors without fake one-element output arrays or discarded failures. Successful
measurements retain the final returned value once after stopping the timer.
Failed measurements retain prior observations and never append a failed result.
Sum's unchecked scalar callback and original array callbacks stay unchanged.

New harnesses `standalone-dot/v1` and `legacy-report-dot/v1` distinguish dot's
checked scalar dispatch from sum. V1's optional ScalarResultBits uint32 field
preserves exact returned NaN/Inf bits safely in JSON. Values are separate from
timing medians and never averaged. Native result-bits and simd-lanes metrics
are required to be integral and in range; actual child width is retained.

Case.Reduction stores NativeLanes for both strategies, VectorAccumulators 0/1,
and these identities:

| Strategy | OrderID | TailID |
| --- | --- | --- |
| scalar | left-to-right-unfused-dot-f32/v1 | none |
| simd-striped | striped-lane-unfused-dot-f32/v1 | scalar-products-after-lanes/v1 |

Native width belongs to comparison context; fold order distinguishes strategies.
The timing baseline policy is `specified-unfused-float32-dot-folds/v1`.
Dot's arithmetic ID prevents pooling with sum or later fused variants. Shared
native-package selection admits only BenchmarkDot leaves for group dot.
Historical option records with absent DotInput normalize to balanced; sum and
array evidence remains readable without changing old identities.

## Fixtures and logical work

Both CLIs share `-dot-input`, default balanced. Only this option affects dot's
generated input. Sum/predicate/scalar-parameter settings do not alter dot identity.

| Fixture | Recipe |
| --- | --- |
| balanced | x[i]=float32(i%33-16)/4; y[i]=float32(i%17-8)/8 |
| cancellation | x repeats 64 entries: 2^24 at 0, 1 at 1, -2^24 at 32, otherwise zero; y=1 |
| mixed-magnitude | x repeats [1e-20,1,1e20,-1e20,-1,3,-3,1e-20]; y repeats [1e20,1e-20,1e-20,1e-20,1e20,1e-20,1e-20,1e20], both float32 |

Only x/y are allocated; no output slice. Identity is
`generated-dot-float32-v1/<fixture>` plus SHA-256 of both arrays' bits. Streamed
and allocated fingerprints agree; changing either input changes the fingerprint.
Logical bytes per call are `8*n+4`, including the returned scalar. These are
effective work units, not measured memory transactions or register spills.
Zero length remains a correctness case, not a throughput point.

## Commands and acceptance

```powershell
go run ./cmd/simdreport -list -group dot
go run ./cmd/simdreport -group dot -dot-input cancellation -bench '^BenchmarkDot$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group dot -dot-input mixed-magnitude -backend go-test -bench '^BenchmarkDot$/^n=1024$' -count 3 -benchtime 100ms -format csv
go run ./examples/dot
```

Enable GOEXPERIMENT=simd for source builds. Installed binaries use identical
flags without Go/source at runtime; Make forwards them through ARGS. Direct
native benchmarks use balanced defaults; the developer report backend records
custom fixture options. See the [Wiki exercise](../wiki/15-Dot-product-and-unfused-rounding.md).

- Independent rounded-product/fold oracles; finite error-bound checks under
  stated assumptions; both rounding-loss directions and explicit nonfusion cases.
- Empty, mismatched lengths, vector/tail boundaries, subnormals, NaN/Inf,
  overflow, signed zero, input immutability and overlapping read-only inputs.
- Zero current-width allocations, typed checked bindings, error propagation,
  partial captures, result-bit retention and mixed-width rejection.
- All three fixtures through compiled/native reports with exact identity/result
  parity; strict metrics and sibling-package selection; invalid options leave
  no artifacts. All nine groups participate in exports and standalone tests.
- Full tests/vet, pure-package tests, forced emulation, CLI/Make exercises,
  consistent command docs and retained smoke evidence accompany this checkpoint.

Stop after unfused one-accumulator dot. Multiple accumulators, fused arithmetic,
new types and M4 remain separate future work.
