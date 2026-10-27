# Changelog

All notable changes are recorded here. Migration milestones are separate from
product roadmap phases; M0 is the architecture behavior-freeze milestone.

## Unreleased

### Architecture migration M1 / Feature 1 — 2026-10-27

#### Added

- `simdreport -format txt|md|csv`; TXT remains the default and terminal output
  remains a readable text comparison for every saved format.
- Markdown tables with literal environment metadata, and encoded CSV containing
  numeric timing/percentage values with empty missing cells and metadata in the
  raw companion file. Unsupported formats fail before execution or artifacts.
- Typed runs, environments, complete case identities, individual observations,
  repetitions, iteration counts, optional counters, and source provenance in
  `internal/result`, with pure median aggregation and scalar comparisons.
- `internal/input` for shared deterministic preparation and streaming SHA-256
  input fingerprints; historical input identity stays explicitly unknown.
- Pure TXT/MD/CSV renderers in `internal/report`, plus identity, encoding,
  cross-format fixture, escaping, writer-failure, and real CLI export tests.

#### Changed

- Removed parsing/aggregation from `cmd/simdreport` and table formatting from the
  testing-dependent `benchutil` surface. The CLI now orchestrates these packages.
- Preserved M0 table arithmetic/order, filtered selections, raw output and
  failed-run evidence. Mixed or ambiguous measurement contexts now fail rather
  than silently pooling incompatible cases. Hyphenated legacy strategy names
  are preserved when stripping numeric GOMAXPROCS suffixes.
- Updated CLI help, README, Make help, and the tracked migration record.

#### Validation and scope

- Formatting, full uncached tests, and vet pass on Windows/amd64. Core result,
  report, and input packages also pass without `GOEXPERIMENT=simd`.
- Real CLI tests cover both groups, default/explicit formats, selected raw
  companions, readable terminal output, empty selections, and benchmark failures.
- Local full race execution retains the M0 `runtime/cgo` limitation; Linux CI
  remains the full race check. Kernel/timed-loop behavior is unchanged.
- M1 is complete; M2 catalog and M3 standalone-backend work remain pending.

### Architecture migration M0 — 2026-10-27

#### Added

- Compatibility fixtures for both command help screens and multi-size reports,
  covering odd/even medians, ordered variants, positive/negative/zero changes,
  fractional timings, filtered variants, and missing/zero scalar baselines.
- Regression coverage for quick-comparison size, variant bindings, allocation
  columns, aliases, the full legacy size sweep, generated-input wraparound,
  and report validation before artifact creation.
- Tracked M0 migration contract defining operation/type/strategy identity,
  arithmetic and harness version boundaries, input fingerprint encoding,
  selector compatibility, and preservation of legacy raw artifacts.
- Native Windows/amd64 baseline with 162 observations and complete command,
  toolchain, hardware, source-revision, and measurement-boundary context.
- Actual two-group report smoke artifacts and reproduction instructions.

#### Validation

- Formatting, uncached tests, vet, and filtered report execution passed on
  Windows/amd64 with Go 1.27.3 and `GOEXPERIMENT=simd`.
- Local race tests could not build `runtime/cgo` (`cgo.exe` exits 2). This also
  affects the unchanged addition package; Linux CI remains the configured race
  check. No local Linux/arm64 execution is claimed.

#### Compatibility

- M0 characterizes existing behavior. Kernel implementations, timed loops,
  command defaults, report formats, and measurement algorithms are unchanged.
- M0 is complete. M1 result/rendering extraction and Feature 1 remain pending.
