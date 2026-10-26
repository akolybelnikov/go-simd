# Architecture migration

This is the tracked implementation record for the selective migration to a
standalone Go SIMD laboratory. Architecture milestones (`M0`, `M1`, ...) are
distinct from product roadmap phases. Architecture foundations precede new
report features. Each milestone is tested, documented, committed, pushed, and
reviewed at a stopping point before the next milestone starts.

## Scope and sequence

| Milestone | Scope | Status |
| --- | --- | --- |
| M0 | Freeze behavior, retain native observations, decide identity/versioning/compatibility contracts | Complete, 2026-10-27 |
| M1 | Extract typed results, aggregation, and rendering; implement TXT/MD/CSV against the extracted model | Pending |
| M2 | Share runtime operation descriptors between commands and developer benchmarks | Pending |
| M3 | Add a compiled `testing.Benchmark` backend; validate outside the source checkout without Go on PATH | Pending |
| Later, conditional | Worker isolation, resource instrumentation, workload adapters, advanced analysis, optional TUI | Requires an observed laboratory or pilot need |

The committed product is an open-source laboratory measuring its implemented Go
kernels. A general workload platform, cloud provisioning, and a default TUI are
outside the initial scope. Offline HTML follows structured results; it must not
depend on resource instrumentation or worker isolation.

## M0 deliverables

- [Behavior and migration contract](m0-contract.md): stable identifiers,
  measurement boundaries, versioning, compatibility, and acceptance.
- [Baseline evidence](baselines/README.md): revision, environment, reproduction
  commands, complete native benchmark output, and report smoke artifacts.
- Command help fixtures under `cmd/simdbench/testdata` and
  `cmd/simdreport/testdata`.
- Synthetic report fixtures covering ordering, odd/even medians, filtered
  variants, missing values, and percentage arithmetic.
- [Changelog](../../CHANGELOG.md): changes and validation limitations.

M0 changes tests and documentation. Existing kernels, source-based report runner,
quick-comparison runner, and timing boundaries remain the migration baseline.

## Next checkpoint: M1

Introduce `internal/result` and `internal/report` as they become used. Keep the
`go test` subprocess while separating legacy parsing, observations, summaries,
and renderers. Retain run, environment, case, and per-case repetition identity;
preserve source timing precision. Keep compatibility fixtures passing while
adding cross-format fixtures for Feature 1. Terminal output stays readable TXT
regardless of the saved format.

M1 does not replace the measurement engine or rewrite kernels. M0's native
timings are observational evidence, never pass/fail thresholds.
