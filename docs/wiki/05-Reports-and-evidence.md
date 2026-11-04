# 05 — Reports and retained evidence

`simdreport` uses the compiled backend by default. It summarizes each case's
observations by median and renders one comparison model as TXT, Markdown or CSV.
Installed binaries work from any directory; `-output` is resolved from there.

## Artifact layout

```text
artifacts/
  addition/
    benchmark-<UTC timestamp>-<unique suffix>.measurements.json
    benchmark-<UTC timestamp>-<unique suffix>.md
  computation/
    benchmark-<UTC timestamp>-<unique suffix>.measurements.json
    benchmark-<UTC timestamp>-<unique suffix>.csv
```

Matching stems tie each summary to its evidence. Every invocation measures again;
changing format does not re-render a previous capture. Unique names preserve
earlier successful captures. The developer `go-test` backend retains `.raw`.

## Format reference

| Property | TXT | Markdown | CSV |
| --- | --- | --- | --- |
| Selection | Default or `-format txt` | `-format md` | `-format csv` |
| Saved table | Aligned text | Header, alignment row, data rows | Header and numeric records |
| Timing precision | Two decimals, ns/op | Two decimals, ns/op | Two decimals, ns/op |
| Changes | Signed one decimal and `%` | Signed one decimal and `%` | Numeric percentage points, one decimal |
| Missing value | Em dash | Em dash | Empty field |
| Environment metadata | Below table | Literal text below table | Evidence companion only |
| Terminal comparison | Readable text | Readable text | Readable text |

An illustrative CSV row with invented values:

```csv
ELEMENTS,Scalar ns/op,SIMDPartial ns/op,SIMDPartial change
1024,100.00,80.00,-20.0
```

`-20.0` means 20% less time than scalar, not the ratio `-0.2`. CSV quoting handles
commas, quotes and embedded newlines. Absent scalar is not a measured zero.

## Compiled evidence: go-simd.measurements/v1

| Top-level key | Contents |
| --- | --- |
| `schema` | Exactly `go-simd.measurements/v1` |
| `status`, `error` | `complete` or `failed`; reason on failure |
| `group` | Measured group |
| `config` | Normalized `count`, `benchtime`, `bench`, optional `input` (scalar uint32 bits, pattern and seed) |
| `config.input.SumInput` | Sum fixture name; balanced, cancellation or mixed-magnitude |
| `config.input.DotInput` | Dot fixture name; balanced, cancellation or mixed-magnitude |
| `config.input.FMAInput` | Fused affine fixture: balanced, rounding or mixed-magnitude |
| `machine` | Measuring process's Go/OS/architecture, vector bits/lanes, emulation, GOMAXPROCS, source revision/modified state and embedded Go build information |
| `run` | Typed run and individual observations; model fields retain Go names in v1, e.g. `Observations` and `DurationNS` |

Observations retain complete measurement keys, operation/type/strategy IDs,
semantics, input generator/fingerprint, size, workers, execution mode, repetition,
actual `Iterations`, exact `DurationNS`, unrounded `NanosecondsPerOp`, available
allocation/throughput counters and the legacy selection alias. Direct ns/op is
duration divided by iterations, preserving fractional nanoseconds. Allocation
metrics divide total counters by actual iterations.

Backend is `testing-benchmark/v1`, harness `standalone/v1` for the original kernels and `standalone-patterns/v1`
for new parameterized bindings. Go controls GC and
calibration; calibration calls are not independent samples. Whole adapter
duration is not kernel CPU/op for final N. Preparation, fingerprints and binding
stay outside the callback. Validation, callback/function-pointer dispatch and
per-call workers remain measured. Typed output is retained once after the
repeated loop, without a full checksum.

Sum uses `standalone-reduction/v1`, with a genuine float32-returning callback.
The timer includes vector extraction/final fold/tail; the last returned value
is retained once after stopping it. Optional `ScalarResultBits` stores uint32
bits safely even for NaN/Inf. `Key.Case.Reduction` records OrderID, observed
NativeLanes, VectorAccumulators and TailID. Width belongs to comparison context;
fold order distinguishes strategies. Values are never averaged in timing tables.
See [the sum guide](https://github.com/akolybelnikov/go-simd/wiki/14-Sum-and-reduction-order)
for decoding and numerical examples. These additive v1 fields are omitted from
array observations when absent; historical captures remain readable.

Dot uses `standalone-dot/v1` and checked float32/error callbacks. It rounds
products before accumulation, includes length validation and final fold in
timing, and retains the last successful result once after timing. It uses the
same optional bit/fold fields with distinct arithmetic/order identities. Errors
propagate, keeping prior observations without appending a failed value. See
[the dot guide](https://github.com/akolybelnikov/go-simd/wiki/15-Dot-product-and-unfused-rounding).

Sum/dot also have SIMDStriped2/4. Their distinct OrderIDs and VectorAccumulators
2/4 use the same v1 fields and existing harnesses. Count describes algorithmic
accumulators, not observed physical registers; actual NativeLanes is separate.
Count/order distinguish strategies, so tables can compare 1/2/4 in that order
after scalar while continuing to reject mixed widths. Input/byte accounting and
old fold identities stay unchanged. See [accumulator experiments](https://github.com/akolybelnikov/go-simd/wiki/16-Accumulator-count-experiments).

CPU model detection is explicitly `unknown`. Missing embedded VCS information
also stays `unknown`; `go run` and copied builds can lack it. Build settings come
from the executable, without runtime `go env`. This is a minimal versioned
evidence document, not M4's proposed worker/manifest bundle.

```powershell
$path = Get-ChildItem ./artifacts/addition/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
$capture.machine
$capture.run.Observations | Select-Object Label, Repetition, Iterations, DurationNS, NanosecondsPerOp
```

Fused affine uses `standalone-affine-fma/v1`, with a software fused reference.
`machine.fma32_path` retains the process gate; `Key.Case.FMA` retains RoundingID,
actual strategy PathID, process VectorPathID and NativeLanes. Mixed process
paths/widths cannot pool. Optional `ArrayResultLastBits` is only the last output
element's uint32 bits, including NaN/Inf safely; it is not an array checksum or
a timing average. Startup probing is outside timing; software correction or
vector broadcasts/arithmetic are inside. See [the FMA guide](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding).

Fused dot uses `standalone-dot-fma/v1`, retaining both `Case.FMA` and
`Case.Reduction`, plus `ScalarResultBits` for every observation. RoundingID is
binary32-fma-updates-add-merge-rne/v1: updates fuse, merges/folds round normally.
Actual PathID distinguishes software reference, software stripes, full vector,
vector with software tail and tail-only software. A process passing the vector
gate can still perform all FMA updates for a short case in software. Logical work is
`8*n+4`; see [the fused-dot guide](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders).

Predicate export uses `standalone-predicate-export/v1`. Case.Predicate records
EncodingID=uint8-zero-one/v1 and actual NativeLanes; mixed widths/encodings
cannot pool. ByteResultSHA256 hashes every output byte in index order once after
timing, separately from the input hash. It is output evidence rather than the
mask data itself. Conversion, extraction and byte stores are timed; hashing is
outside timing. Logical work is `5*n`. See [the guide](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks).

Packed predicates use `standalone-packed-predicate/v1`. Case.Predicate records
uint8-packed-lsb0/v1 and actual NativeLanes; ByteResultSHA256 covers the entire
packed sequence including zeroed final padding. Physical output length is
`ceil(n/8)`, and logical work is `4*n+ceil(n/8)`. Conversion/packing/stores are
timed, hashing afterward. Byte-mask and packed series retain distinct output,
operation, arithmetic and harness identities. See [the guide](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks).

Copy uses `standalone-copy/v1`, with Case.Copy retaining exact-bit policy and
actual process SDK lanes. ArrayResultSHA256 hashes every output float32's bits
in little-endian order after timing; it equals x's input hash for correct copy.
It is distinct from byte-mask output and last-element fused-affine evidence.
Mixed bit policies/SDK widths cannot pool. Source labels do not prove physical
instruction width; see [the copy guide](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement).

## Explicit developer backend

`-backend go-test` invokes `go env` and repository-relative `go test` deliberately.
It needs the checkout, matching Go on PATH and `GOEXPERIMENT=simd`; it never
falls back to compiled execution. Its `.raw` companion retains native text,
backend `go-test-text/v1`, harness `legacy-report/v1` for original cases or `legacy-report-patterns/v1`
for parameterized cases. New raw headers record input configuration; absent
historical parameter/fixture context remains unknown.

Sum uses `legacy-report-reduction/v1`. Its native `result-bits` and `simd-lanes`
metrics retain exact scalar bits and the child process's width outside timing;
missing, fractional or out-of-range metrics are rejected.

Dot uses `legacy-report-dot/v1` with the same strict metric requirements. The
reported width comes from the child process. The report selects only dot's
validated leaves from its native package, which also contains sum.

Legacy text retains tool/environment output, repetitions, iterations, displayed
timing, allocations and throughput. It cannot recover exact duration from rounded
ns/op and lacks complete build/emulation identity. Parent metadata does not
establish child capabilities. Historical fingerprints stay unknown; old evidence
is never silently converted. Both backends feed the same summary renderers, but
their results must remain separate series because harness/precision differ.
Throughput is descriptor-specific logical work: `8*n` for unary patterns and
`12*n` for binary array-output operations, `4*n+4` for sum and `8*n+4` for dot,
not measured DRAM traffic.

Fused affine uses `legacy-report-affine-fma/v1`. Strict child `fma-vector`,
`fma-path`, `simd-lanes` and `last-result-bits` metrics establish the actual path,
width and retained output cell; parent capabilities cannot establish them.
Fused/unfused arithmetic stays separate. Additive v1 FMA fields are omitted
from older unrelated cases.

Fused dot uses `legacy-report-dot-fma/v1`. Strict child result-bits,
simd-lanes, fma-vector and fma-path metrics establish the scalar bits, width,
process gate and actual case path. Its five path codes (0..4) are specific to
this operation; affine's existing 0/1 code contract stays unchanged. Missing,
fractional, out-of-range or gate/size-inconsistent metrics fail import. Native
and compiled cases agree on numerical/fixture/fold/path identity; their distinct
harnesses keep timing series separate.

Predicate export uses `legacy-report-predicate-export/v1`. Exact child
simd-lanes and eight output-sha256-0..7 uint32 metrics reconstruct the complete
output digest in big-endian word order. Missing, duplicate, fractional,
non-finite, negative or oversized digest words and invalid widths fail import.
Parent facts and expected output never replace child evidence. Existing v1
captures omit the optional Predicate/ByteResultSHA256 fields when unrelated.

Packed predicates use `legacy-report-packed-predicate/v1` with the same strict
child width/digest-word requirements, retaining the packed encoding. Compiled
and native selectors share additional byte boundaries 7,9,63,65. Expected
output or parent capabilities never substitute for child facts.

Copy uses `legacy-report-copy/v1`. Strict child SDK width and eight exact output
digest words retain full copied-bit evidence. Fresh fixed-input imports establish
the copy generator/hash; historical unknown input stays unknown. Expected output
does not replace a child digest. Percentages use the Scalar source-loop baseline;
GoCopy remains an additional control. SDK emulation does not disable independent
compiler/runtime copy optimization.

## Failure and sharing

Invalid format, group, count, duration/iteration mode, regex, backend and
positionals fail before artifacts. An empty compiled selection or unexpected
measured failure retains a failed capture, including completed earlier samples
when available, and no successful summary for that group. With multiple selected groups, one
can complete before the next fails; inspect the final exit status.

Disk errors or abrupt process termination can leave incomplete files. Atomic
bundles, cancellation recovery and worker isolation are deferred. A saved-run
analyzer/offline re-render CLI is not implemented. Preserve evidence companions,
source context, the experiment question and limitations when sharing summaries.
The lab does not upload data or provision infrastructure automatically.


## Clamp evidence

Clamp records Case.Clamp with ordered-clamp-preserve-bits/v1 and actual SDK lanes.
Exact lower/upper parameter bits, recipe/seed and input fingerprint are retained.
ArrayResultSHA256 covers the complete little-endian float32 output after timing;
it need not equal the input hash. The default standalone-clamp/v1 and developer
legacy-report-clamp/v1 boundaries include all validation/broadcasts/repairs/stores.
Native import requires positive integral child lanes and eight complete digest
words. Different policies/widths cannot pool silently. See [clamp](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts).


## MulAdd evidence

The [MulAdd guide](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add) covers three real operand arrays, a full output digest,
16*n traffic, actual FMA gate/path/width and matched timed work. Use
`go run ./examples/muladd` for the module-local four-slice API.
