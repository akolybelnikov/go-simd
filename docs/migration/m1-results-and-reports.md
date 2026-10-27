# M1: structured results and report formats

Completed **2026-10-27**. This milestone extracts the architecture boundaries
defined by [M0](m0-contract.md) and implements Feature 1 against them. Runtime
catalog work remains M2; standalone execution remains M3.

## Dependency boundaries

```text
cmd/simdreport -> existing go test subprocess -> result.ReadLegacy
              -> result.Summarize -> result.Compare -> report.Write

internal/benchutil -> internal/input (preparation outside timing)
internal/result   -> internal/input (fresh-input fingerprints outside timing)
internal/report   -> internal/result
```

`result`, `report`, and `input` production code have no SIMD, `testing`, or process
execution dependency. They can be tested without enabling the SIMD experiment.
Kernels and timed benchmark loops are unchanged. The report command orchestrates
execution and artifact persistence; renderers only consume typed comparisons.

## Internal records

| Record | Retained information |
| --- | --- |
| `Run` | Unique artifact-stem run ID, environment, backend/harness IDs, source format, readable metadata, individual observations |
| `Environment` | Referenced ID and child-reported OS, architecture, CPU; conflicting headers are rejected |
| `Case` | Operation/type/strategy IDs, arithmetic mode, parameters, size, input identity, worker availability/count, execution mode |
| `MeasurementKey` | Complete case plus run/environment/backend/harness identity and observed GOMAXPROCS suffix |
| `Observation` | Key, one-based per-case repetition, iterations, unrounded parsed ns/op, optional allocation/throughput metrics, source benchmark name |
| `Summary` | Complete measurement key, label, median, sample count |
| `Comparison` | Ordered columns and size rows with precomputed timings and percentage changes; explicit legacy serial-scalar baseline policy |

These are internal Go records, not a published JSON/NDJSON artifact schema.
Legacy `.raw` files retain their existing meaning and original observations;
TXT/MD/CSV are summary exports. The current export does not serialize all typed
identity fields. That storage contract belongs to a later versioned artifact.

Fresh report runs identify `generated-float32-v1` inputs and compute SHA-256 over
the M0 little-endian input-bit encoding outside measurement. Hashing streams
generated values without allocating another large input/output set, and caches
the digest per size. Shared generation logic keeps hashing and preparation
consistent. Historical import defaults to unknown input identity; text alone
cannot prove a fingerprint or source revision. Exact total duration remains
unavailable rather than reconstructed from rounded ns/op.

The importer retains only facts supplied by child headers/suffixes. Existing
readable metadata still includes parent runtime context, with the limitations
documented by M0. Worker counts follow the known legacy implementation and
observed concurrency; unknown metadata stays unavailable.

## Aggregation and comparison policy

Samples aggregate only by their complete measurement key. Changing operation,
type, arithmetic semantics, parameters, input identity, workers, execution mode,
run, environment, backend, harness, or GOMAXPROCS creates a distinct series.

A comparison accepts one compatible operation/context and input-generator
family. Per-size input fingerprints must match across strategies. Ambiguous
duplicates, such as two worker counts for the same strategy and size, fail
instead of overwriting a table cell. Mixed contexts require separate tables.

Legacy report behavior is preserved: ascending sizes, existing variant order,
odd/even medians, ratios of unrounded medians, and unavailable changes without a
positive scalar baseline. Parallel SIMD still compares against serial scalar,
as explicitly recorded by `legacy-serial-scalar/v1`; M1 adds no matched parallel
control and makes no vectorization-only claim.

Malformed negative/non-finite measurements and invalid iteration counts fail
clearly with a source line. Unknown benchmark/strategy names retain a legacy
namespace. Only numeric CPU-count suffixes are stripped, so hyphenated unknown
strategy names remain intact.

## Feature 1 conventions

```powershell
go run ./cmd/simdreport -group addition -format md
go run ./cmd/simdreport -group computation -format csv
```

| Format | Saved file | Missing values | Percentages | Metadata |
| --- | --- | --- | --- | --- |
| `txt` (default) | `.txt` | Em dash | Signed, one decimal, `%` | Below table |
| `md` | `.md` | Em dash | Signed, one decimal, `%` | Indented literal text below table |
| `csv` | `.csv` | Empty field | Numeric percentage points, one decimal | Companion raw only |

All formats use identical sizes, timing precision (two decimals), ordering, and
computed comparisons. CSV uses `encoding/csv`, with a header and one row per
size; embedded commas, quotes, and newlines are encoded correctly. Markdown has
a header/alignment row and escaped cell content. Terminal output always uses
TXT. `-format` affects only saved reports and is validated before any benchmark
or artifact creation. `simdbench` retains its existing interface and output.

## Verification and checkpoint

- M0 addition/computation/filtered fixtures pass through the extracted packages.
  The report help fixture adds only the intentional `-format` flag/default.
- A shared structured fixture has independently specified TXT, Markdown, and
  CSV outputs, including fractional values, missing cells, and zero baselines.
- Model tests verify identity separation, per-case repetitions, retained counters,
  bit-level input encoding, no fabricated durations, and no input mutation.
- Real subprocess tests exercise default TXT and explicit TXT/MD/CSV for both
  groups at sizes 8/16, two repetitions, `1x`. Each saved/terminal table is checked
  against its own raw observations, without performance thresholds.
- Empty selections and failed benchmarks retain raw evidence, return exit 1,
  and produce no successful summary file. Writer failures propagate.
- Formatting, full uncached tests, and vet pass on Windows/amd64. Local full
  race execution remains blocked by the M0 `runtime/cgo` toolchain issue;
  Linux CI supplies the configured full race check.

M1 is the stopping point. M2 catalog work has not started.

Portability follow-up during Wiki publication (2026-10-27): select universal
power-of-two sizes 8/16 in the export smoke test. The earlier 8/9 selection
assumed a width-dependent boundary; size 9 is absent from a width-16 sweep.
This corrects test coverage without changing production selection or kernels.
