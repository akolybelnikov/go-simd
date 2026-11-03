# F004, slice 2a — Ordered float32 sum

Implemented 2026-11-03 after the broadcast/predicate slice. This checkpoint adds
sum and genuine scalar-result support, then stops before dot product. The lab
now contains eight operations and eighteen implementations.

Signed code checkpoint:
[72700a6](https://github.com/akolybelnikov/go-simd/commit/72700a67ec74a1996c79bae75ce4b37f0a3cbc4e).
The [retained cancellation smoke](evidence/sum-cancellation.measurements.json)
and [provenance/reproduction notes](evidence/README.md) accompany this record.

## Operation and numerical contract

Group `sum`, operation `sum-f32`, benchmark `BenchmarkSum`, input `x []float32`,
output one returned `float32`. Arithmetic family is
`float32-sum-ordered-folds-v1`. There is no caller output array or length-error
sentinel for this one-input reduction. Kernels do not modify x.

| Label / strategy | Specified order |
| --- | --- |
| Scalar / scalar | Start at +0; add x left to right, rounding every addition to float32 |
| SIMDStriped / simd-striped | One zero-initialized native-width vector accumulator; add full vectors by increasing index; extract and fold lanes low to high from +0; add remaining scalar tail left to right |

Extraction, final lane fold and scalar tail are included in kernel timing.
No padded input lanes are loaded or reduced. Current widths use stack scratch;
a larger future width uses a declared dynamic scratch allocation, reflected in
allocation counters. Current hardware/emulated kernels allocate zero per call.
Empty input returns +0, including nil. All-negative-zero inputs also return +0
under these initial-value/order rules. NaN payloads are not fixed. Inf and
overflow follow each specified IEEE float32 fold: different orders can even
produce different exceptional-value classifications from finite inputs.

Scalar is the timing baseline, not an exact mathematical oracle. Reassociation
is not bitwise equivalent. For a native width w, make x of length 2*w:

- x[0]=2^24, x[1]=1, x[w]=-2^24: scalar returns 0; SIMD returns 1.
- x[0]=2^24, x[1]=-2^24, x[w]=1: scalar returns 1; SIMD returns 0.

All other entries are zero. Neither strategy is universally more accurate.
Correctness tests use an independent exact-rational addition oracle, rounded
to float32 after each specified step, and compare result bits except NaN payloads.
For eligible finite nonoverflowing cases they also check the absolute forward
bound `gamma(k)*sum(abs(x))`, where `u=2^-24`, `gamma(k)=k*u/(1-k*u)` and the
conservative k is n+w+1. This applies only when k*u<1; there is no fixed relative
tolerance near cancellation or universal bound promised for the largest sweep.
These are numerical test rules, not benchmark timing thresholds or an automatic
accuracy verdict in reports.

## Typed execution and retention

The catalogue's existing scalar output shape now has executable support through
`BindScalar` callbacks returning float32. Prepared execution has distinct array
and scalar paths. Sum never allocates or passes a one-element output slice.
Its callback has no per-call error check because a single input slice and returned
scalar have no recoverable length mismatch. Existing array dispatch stays intact.

The benchmark retains the last complete returned value once after stopping the
timer. A typed float32 destination and sink prevent dead-code removal; there is
no checksum or per-iteration output encoding. Direct evidence stores optional
`Observation.ScalarResultBits` as a uint32, preserving NaN/Inf safely in JSON.
Arithmetic values remain separate from timing summaries and are never averaged.

New harness IDs are `standalone-reduction/v1` and
`legacy-report-reduction/v1`; existing harnesses/kernels are unchanged. Native
benchmarks report `result-bits` and `simd-lanes` metrics outside timing. In this
Go toolchain integer metrics up to uint32 are rendered exactly; the importer
requires integral/in-range values and rejects missing sum metrics.

`Case.Reduction` records OrderID, observed NativeLanes, VectorAccumulators and
TailID. The scalar order has zero vector accumulators; SIMD has one. NativeLanes
is the process's SIMD width for both strategies, not the scalar accumulator count.
It is part of comparison context, so mixed widths cannot silently combine.
Different fold orders are deliberate strategy dimensions under this arithmetic
family; duplicate strategy/size cases remain ambiguous and rejected.

These evidence fields and `config.input.SumInput` are additive to v1. Zero
reduction identity and absent scalar result bits are omitted for old array
records. Historical array evidence remains readable. Old option records with no
SumInput field resolve to the original default, balanced.

## Fixtures, bytes and commands

Both CLIs accept `-sum-input`, default balanced. Other pattern/scalar controls do
not change sum's input or case identity.

| Fixture | Values |
| --- | --- |
| balanced | x[i]=float32(i%33-16)/4 |
| cancellation | Every 64 entries: index 0 is 2^24, 1 is 1, 32 is -2^24, others zero |
| mixed-magnitude | Repeat [1e-20,1,1e20,-1e20,-1,3,-3,1e-20] as float32 |

Input identity is `generated-sum-float32-v1/<fixture>` plus SHA-256 of x's bits.
Both allocated and streamed fingerprints agree. No second array or output array
is allocated during fixture preparation. Logical work is `4*n + 4` bytes per
call, including a scalar result, not measured DRAM reads/writes. Zero sizes remain
correctness cases rather than throughput points.

```powershell
go run ./cmd/simdreport -list -group sum
go run ./cmd/simdreport -group sum -sum-input cancellation -bench '^BenchmarkSum$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group sum -sum-input mixed-magnitude -bench '^BenchmarkSum$/^n=1024$' -count 3 -benchtime 100ms -format csv
go run ./examples/sum
```

Enable GOEXPERIMENT=simd when building from source. Installed binaries use the
same flags without Go/source; Make forwards them via ARGS. Explicit go-test
reports use the same requested fixture and preserve native result/fold facts;
direct native commands default to balanced unless their internal configuration
environment is supplied. Prefer the report CLI for recorded custom settings.

## Verification and checkpoint

- Rational per-addition oracles across empty, vector-boundary, tail, cancellation,
  subnormal, random finite, Inf/NaN, signed-zero and overflowing cases.
- Both rounding-loss counterexamples and the finite absolute error bound where
  its assumptions apply; input immutability and zero current-width allocations.
- Genuine typed scalar preparation, raw-bit retention including NaN, native
  result/identity parity, strict native metrics, and mixed-width rejection.
- All eight groups in real TXT/MD/CSV exports and no-Go/no-source binary tests;
  quick scalar callbacks, descriptor traffic and fixture fingerprints.
- Hardware, forced portable emulation, full tests/vet and pure-package tests;
  CI adds sum to emulation and native smoke checks and retains Linux race checks.
- Runnable example, compact/detailed catalogue, README/Wiki updates and retained
  smoke evidence. Timings remain observations without performance thresholds.

Stop here. Dot product, multiple vector accumulators, FMA, new numeric types and
other operation families remain separate future work. M4 workers/manifests are
not activated by this checkpoint.
