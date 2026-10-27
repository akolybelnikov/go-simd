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

## Explicit developer backend

`-backend go-test` invokes `go env` and repository-relative `go test` deliberately.
It needs the checkout, matching Go on PATH and `GOEXPERIMENT=simd`; it never
falls back to compiled execution. Its `.raw` companion retains native text,
backend `go-test-text/v1`, harness `legacy-report/v1` for original cases or `legacy-report-patterns/v1`
for parameterized cases. New raw headers record input configuration; absent
historical parameter/fixture context remains unknown.

Legacy text retains tool/environment output, repetitions, iterations, displayed
timing, allocations and throughput. It cannot recover exact duration from rounded
ns/op and lacks complete build/emulation identity. Parent metadata does not
establish child capabilities. Historical fingerprints stay unknown; old evidence
is never silently converted. Both backends feed the same summary renderers, but
their results must remain separate series because harness/precision differ.
Throughput is descriptor-specific logical work: `8*n` for unary patterns and
`12*n` for the binary operations, not measured DRAM traffic.

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
