# F004, slice 1 — Broadcasts, predicates and mask composition

Implemented 2026-10-27 after architecture M0-M3. This slice adds five operations,
each with scalar and partial-vector implementations. The catalogue now contains
seven operations and sixteen implementations. Existing addition/computation
arithmetic, fixtures, labels, and timed call paths are retained.
Implementation commit: `dc80a04` (`feat(ops): add broadcast and predicate catalogue experiments`).

## Admitted operations

| Group | Stable operation ID | Contract | Scalar parameters | Logical bytes |
| --- | --- | --- | --- | --- |
| scale | `scale-f32` | `alpha*x` | alpha | `8*n` |
| affine | `affine-f32` | `float32(alpha*x)+beta` | alpha, beta | `8*n` |
| threshold | `threshold-zero-f32` | `x > threshold ? x : +0` | threshold | `8*n` |
| selection | `select-gt-f32` | `x > threshold ? x : y` | threshold | `12*n` |
| rangefilter | `range-zero-f32` | `lower < x && x < upper ? x : +0` | lower, upper | `8*n` |

Each has `Scalar` / `scalar` and `SIMDPartial` / `simd-partial`, serial worker
policy and float32 array output. SIMD uses portable BroadcastFloat32s,
Greater/Less, Masked, IfElse and Mask32s.And, verified against the local Go 1.27.3
SDK. No architecture-specific ISA or feature-check bypass is required.

Kernels live in `internal/patterns`. Scalar/SIMD length checks run once per call;
SIMD constructs scalar broadcasts once per call, inside timing. Scalar argument
extraction and closure binding happen during preparation. Partial loads/stores
cover every chunk, including the tail; no reduction counts inactive lanes.

Scale/affine have one array input; selection has two. Unary fixtures allocate no
unused second input. Byte accounting follows declared arrays/output and excludes
scalar setup/intermediates. In selection, scalar source branches may avoid some
y loads while SIMD loads both arrays. `12*n` remains logical work, not observed
memory traffic. These are complete kernels, not instruction-latency studies.

## Numerical, alias and boundary contract

- Affine explicitly rounds the product to float32 before adding beta. It is an
  unfused experiment; no MulAdd/FMA variant is admitted. A cancellation fixture
  distinguishes fused from separately rounded results.
- Arithmetic comparisons use exact float32 results for non-NaNs, including
  signed zero. Arithmetic NaNs need only remain NaN; payload preservation is not
  promised for multiplication/addition. Overflow and infinities are valid outputs.
- Predicates use ordered strict comparison. Equality/unordered NaN comparisons
  take the false branch. Threshold/range false results are positive zero.
- Selected x/y values preserve their bits, including negative zero and y NaN
  payloads. A selected negative zero is retained; false-to-zero remains +0.
- Range has exclusive bounds. Reversed, empty or NaN bounds produce all-false
  results in direct kernel calls. Benchmark fixture validation is separate.
- Disjoint output and exact aliasing either declared array input are supported.
  Partial overlap is unsupported. Length failures preserve output. Nil/empty
  equal slices succeed. Tests include output canaries, but do not claim to prove
  the absence of all possible SDK memory overreads.

## Parameters and fixture identity

Both commands share `-alpha 1.5`, `-beta -0.5`, `-threshold 0`, `-lower -1`,
`-upper 1`, `-pattern alternating`, `-seed 42`. Flags parse directly to float32;
out-of-range numeric text is rejected rather than silently rounded through an
implicit float64 input contract. NaN/Inf literals are accepted. Parameters apply
only to operations declaring those names; unrelated settings are ignored by case
identity. Relevant values are encoded in descriptor order as exact IEEE-754 bits,
e.g. `alpha=float32:0x3fc00000;beta=float32:0xbf000000`.

Signed-zero and NaN-payload parameter distinctions survive identity. Evidence
configuration stores uint32 scalar bits, not non-finite JSON numbers. The v1
capture gains optional `config.input`; all existing v1 fields remain unchanged.

Scale/affine use `generated-unary-float32-v1`: `x[i]=float32(i%33-16)/4`.
Predicate fixtures use a false value equal to the lower comparison bound and a
true value initially bound+1. If that is not a representable satisfying value,
the generator tries the next float32 toward +Inf. This avoids a default
subnormal true value at threshold zero while retaining strict-boundary behavior.
Selection y is `float32(i%17-8)/4`.

| Pattern | Predicate arrangement |
| --- | --- |
| none | 0% true |
| all | 100% true |
| alternating | True at even indices; ceil(n/2) true |
| clustered | First ceil(n/2) true, remainder false |
| shuffled | Deterministic SplitMix64 index/seed mixing; approximately 50% true |

Pattern and shuffled seed appear in the generator ID. Only shuffled consumes
seed. SHA-256 hashes exactly the declared input arrays, retaining float bits;
streamed generation and allocated fixtures are tested for equal fingerprints.
Input generator/pattern/seed/fingerprint and relevant parameter bits prevent
pooling different experiments. Alternating and clustered fixtures differ in
ordering even when their true counts agree.

If a requested pattern needs true values but none exist (NaN/+Inf threshold,
an empty/reversed range, or no representable strict interior), preparation fails
before artifacts or benchmark fixtures. `-pattern none` permits those all-false
experiments. Direct numerical kernel calls accept all parameter values.

## Execution and compatibility

New typed configured binders coexist with the six legacy function/buffer
bindings. Existing timed dispatch stays intact. New bound-parameter callbacks
have harness `standalone-patterns/v1`; the explicit native text path uses
`legacy-report-patterns/v1`. Backend algorithms retain their existing IDs.
Do not pool these with legacy harness series. Output retention remains a single
float32 sink after the repeated loop, without a timed whole-array checksum.

All new native benchmarks are in `./internal/patterns`. Catalogue metadata
declares that development package explicitly. The go-test adapter expands the
validated selected leaf set to an exact Go expression, preventing a selected
group from measuring sibling operations in the shared package. Scalar bits and
fixture options pass through an internal JSON environment setting and are
recorded in raw metadata. The legacy importer recovers recorded parameter/input
identities for fresh generated runs; absent historical configuration stays
unknown. Compiled/native parity tests cover all five groups with nondefault
parameters and a seeded fixture. There is no automatic backend fallback.

`-list` now gives a compact operation table; `-list -group <name>` shows detailed
expressions, parameter values/bits, fixtures and strategies. Both CLIs and all
Go/Make/installed launchers share these flags. No group-specific Make targets
are needed: use ARGS for new groups.

## Verification and stopping point

- Independent scalar-formula oracles for all ten configured bindings, at empty,
  small, vector-boundary, tail and 1025-element lengths; exact aliases, invalid
  lengths, output canaries, infinities, subnormals, signed zeros and NaNs.
- An explicit unfused affine counterexample; bit-preserving selection tests.
- None/all/balanced/seeded fixtures, impossible-distribution rejection, and
  allocated/streamed fingerprints with correct unary/binary storage.
- Frozen legacy IDs/bindings/tables, configured inventory/traffic, all summary
  formats and raw/direct parameter/input parity. Standalone no-Go/no-source tests
  now run all seven groups. Native new-operation benchmark smoke passes.
- Hardware and forced portable emulation checks. CI adds `GODEBUG=simd=0`
  pattern/binding checks on Windows and Linux, plus existing Linux race coverage.
- Full tests/vet, SIMD-independent metadata/input/evidence/result/report tests,
  runnable examples and README/Wiki command/link/table/fence validation.

`examples/patterns` demonstrates all five operations. Smoke evidence under
`docs/features/evidence` retains a seeded selection capture with actual parameter
bits/fingerprints; short 2x timings are plumbing evidence, not speedup claims.

This completes the first F004 slice. Sum/dot/reductions, explicit FMA, clamp,
predicate export/packed masks, conversions, new numeric types and archsimd
movement remain separate backlog slices. Worker isolation and infrastructure
features are not activated. Reading credit remains [Jesús Espino, SIMD and Go](https://internals-for-interns.com/posts/go-simd/),
the prompt for this pattern-first expansion.
