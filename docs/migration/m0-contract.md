# M0: behavior freeze and migration contract

Decision date: **2026-10-27**. Baseline source:
`3ce769ca33eec166aa389c33c1de1ed315deca8a`.

This record defines preserved behavior and explicit version boundaries. Decisions
about future internal records guide subsequent milestones; they are not APIs or
flags implemented by M0.

## 1. Command compatibility

| Surface | Frozen behavior |
| --- | --- |
| `simdbench` defaults | Both groups; 1,048,576 generated `float32` elements; one `testing.Benchmark` result per variant |
| Group selection | `-group addition\|computation\|all`; quick CLI also retains `-function add\|compute\|all`, `-addition`, and `-computation` |
| Explicit aliases | Matching selections accepted; conflicts rejected; false boolean aliases do not select a group |
| `simdreport` defaults | `-group all -count 5 -benchtime 100ms -bench . -output artifacts` |
| Timing selection | Duration or `Nx` remains Go benchmark syntax; count means repeated measurements |
| Benchmark selection | Go's slash-separated regex semantics, including anchors and unanchored matching |
| Help | `-h` succeeds, writes usage to stderr, and does not benchmark; fixtures retain flags/defaults |
| Errors | Report main exits 1; quick main exits 2; command-prefixed diagnostics; invalid groups/counts/positionals fail before reports are created |
| Headless behavior | Neither command requires a TTY or implicitly launches a frontend |

Reports still require the repository root and Go executable. Standalone behavior
is an M3 target. The quick command retains `testing.Benchmark`'s default duration;
it does not expose report count/benchtime options.

Existing command paths and selectors remain throughout the standalone transition.
A replacement must support legacy `-bench` aliases with equivalent Go matching
before switching wrappers. Unsupported expressions must fail clearly; substring
matching or automatic fallback to source execution is not compatible behavior.
Native `go test -bench` remains a contributor workflow. Removing compatibility
commands/selectors requires deprecation in at least two tagged releases, followed
by a documented breaking release. M0 does not rename installed commands.

## 2. Inputs, kernels, and timing boundaries

Generated fixture ID: `generated-float32-v1`. For zero-based index `i`:

```go
a[i] = float32(i%100) / 10
b[i] = float32(i%50-25) / 10
```

Two inputs and a disjoint reusable output are prepared outside the timed inner
loop. Exact in-place aliasing remains separately tested kernel behavior; partial
overlap is unsupported. Length errors leave output unchanged. Timed calls retain
validation, closure/function-call overhead, and error checks. Output retention
is outside the repeated kernel loop. Go controls calibration and GC; M0 adds no
warmup, cache flush, worker pool, CPU affinity, or frequency control.

The sweep is sorted unique powers of two from 1 through `1 << 24`, plus positive
`width-1`, `width`, `width+1`, `2*width-1`, `2*width`, and `2*width+1` sizes.
The byte count stays `12 * elements`: two `float32` reads and one write. Throughput
is an effective rate over reused buffers, not measured memory traffic.

Serial strategies use one calling goroutine. Parallel addition starts
`min(effective GOMAXPROCS, elements)` goroutines for nonempty input, none for
empty input. Contiguous chunks differ by at most one element; the first workers
receive the extras. Startup, synchronization, and kernel work remain timed.
This comparison still needs a matched parallel scalar control in future work.

Computation remains `x*x + y*y + x*y`: the existing scalar expression and SIMD
`Mul`/`Add` sequence. Its legacy semantic ID describes these implementations;
it does not promise fused/unfused bit equality across toolchains. Resolve and
version floating-point acceptance rules before changing arithmetic or adding FMA.

## 3. Stable identifiers and case identity

Machine keys are lowercase. Display labels and native benchmark names remain
presentation/selection aliases.

| Legacy group / benchmark | Operation ID | Type ID | Arithmetic semantics ID |
| --- | --- | --- | --- |
| `addition` / `BenchmarkAddition` | `add` | `float32` | `go-float32-add-v1` |
| `computation` / `BenchmarkComputation` | `compute-expression` | `float32` | `go-float32-quadratic-v1` |

| Operation | Legacy label | Strategy ID |
| --- | --- | --- |
| Both | `Scalar` | `scalar` |
| `add` | `SIMDPartial` | `simd-partial` |
| `add` | `SIMDTail` | `simd-scalar-tail` |
| `add` | `SIMDParallel` | `simd-parallel-scalar-tail` |
| `compute-expression` | `SIMD` | `simd-partial` |

Logical case identity is `(operation, type, strategy, elements, arithmetic
semantics, parameters, input identity, actual workers, execution mode)`. Current
parameters are empty; execution mode is `kernel-only`. Labels, titles, sample
numbers, and timestamps are not part of logical case identity.

Input identity includes generator ID/version, length, and a SHA-256 digest of
input bits. M1's encoding is all `a` elements followed by all `b` elements, each
`math.Float32bits` encoded as four little-endian bytes; hash outside timing.
Imported inputs require an explicit descriptor. Absent legacy fingerprints are
unknown, not an assumed match.

M1 represents the tuple with typed fields, not slash-delimited labels. Public
serialized encoding is deferred until a versioned artifact ships; consumers
must not depend on undocumented concatenation or hashing. Observations also
reference a run ID and one-based per-case repetition ordinal. Compatibility
requires matching environment, source/build identity, backend, harness, and
timing configuration, recorded separately from the logical case key.

## 4. Backend and harness versions

| Path | Backend ID | Harness ID | Timing source |
| --- | --- | --- | --- |
| Existing report | `go-test-text/v1` | `legacy-report/v1` | Native text at its available precision |
| Existing quick comparison | `testing-benchmark/v1` | `legacy-quick/v1` | Existing `BenchmarkResult.NsPerOp()` presentation |
| Future standalone adapter | `testing-benchmark/v1` | `standalone/v1` | Actual `BenchmarkResult.T` and `N` |

Version the harness when timed dispatch, validation/error handling, setup,
calibration, retention, worker lifecycle, or accounting boundaries change.
Version the backend when the measurement algorithm changes and semantic IDs
when the arithmetic contract changes. Record source revision independently;
never pool observations across versions automatically.

The text importer may retain parsed ns/op and iteration counts, but cannot
recover exact total duration by multiplying rounded text by `N`. Missing duration,
resources, build identity, and worker metadata stay unavailable. Future direct
results retain actual duration/iterations and defer rounding to rendering.
Calibration calls are not independent samples. Whole-call CPU time divided by
final `N` is not kernel CPU/op.

## 5. Aggregation and rendering compatibility

- Aggregate each case independently. Sort sizes ascending and variants in order:
  `Scalar`, `SIMD`, `SIMDPartial`, `SIMDTail`, `SIMDParallel`; unknown labels follow
  lexically. Only observed variants appear.
- Keep the median: odd counts use the middle value, even counts average the two
  middle values. Preserve parsed precision until display.
- Change is `(median_variant - median_scalar) / median_scalar * 100`, a ratio of
  medians rather than an aggregate of per-repeat percentage ratios.
- Keep one row per size, two decimals for ns/op, signed one-decimal percentages,
  and an em dash for missing measurements or missing/nonpositive scalar baselines.
  Ties render `+0.0%`.
- Keep quick-table allocation columns and the scalar `baseline` label. Quick
  timing currently uses integer `NsPerOp()`; a precision change must be deliberate.
- Keep readable terminal and table-first TXT reports followed by metadata.
  Alignment padding can change; headings, row/column order, precision, signs,
  blank lines, and missing markers are fixture-protected.

Synthetic fixtures under `cmd/simdreport/testdata` check arithmetic independently
of timing. Native observations are evidence, never speedup test thresholds.

## 6. Artifacts and metadata

Keep unique `artifacts/<group>/benchmark-<UTC timestamp>-*.raw` and companion TXT
files. Raw files retain metadata, individual benchmark lines, allocations, and
throughput. Failed benchmark output remains in raw, its path is reported, and
no successful summary is produced. Empty selections report `no benchmark
measurements found` and retain raw output.

Structured observations require a new explicitly versioned format; do not reuse
`.raw` for a different schema silently. TXT/MD/CSV exports are independent of raw
format. Do not delete, convert, or overwrite legacy evidence automatically. Keep
the legacy reader through the transition until a documented offline importer
replaces it. Original bytes remain evidence; no retention expiry is introduced.

M1 carries run/environment references and source-format provenance. Subsequent
metadata capture must include source revision/dirty state (or explicit unknown),
toolchain/build settings, CPU/OS/architecture, vector lanes/bits, emulation,
effective GOMAXPROCS, actual workers, fixture identity, exact selection/count/
timing, and harness/backend versions. The execution process is authoritative;
parent-only metadata does not establish child capabilities. Label unsupported
detection unknown. M0 records existing metadata limitations without changing it.

## 7. Validation and exit record

| Gate | Evidence |
| --- | --- |
| Command/help/report freeze | Checked-in help and synthetic tables; quick-table/registration regression tests |
| Native baseline | 162 observations: 9 sizes, 6 operation/strategy pairs, 3 repetitions, 100 ms target |
| Environment/provenance | Source revision, generator formula, OS/CPU/toolchain settings, lanes, GOMAXPROCS, commands |
| Real report workflow | Both groups, sizes 8/9, two repetitions, `1x`; companion TXT/raw evidence |
| Formatting, tests, vet | Passed on local Windows/amd64 with `GOEXPERIMENT=simd` |
| Race detector | Local `runtime/cgo` exits 2 before tests build; existing Linux CI race job remains configured |
| Other targets | Linux/amd64 CI configured; no local Linux or arm64 execution claimed |

Reviewers can identify preserved behavior and versioned changes from this
contract and fixtures. M0 is complete; stop here before M1. The race-toolchain
limitation is recorded; M0 changes no kernel or timed code.
