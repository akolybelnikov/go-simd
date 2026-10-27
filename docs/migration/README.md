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
| M1 | Extract typed results, aggregation, and rendering; implement TXT/MD/CSV against the extracted model | Complete, 2026-10-27 |
| M2 | Share runtime operation descriptors between commands and developer benchmarks | Complete, 2026-10-27 |
| M3 | Add a compiled `testing.Benchmark` backend; validate outside the source checkout without Go on PATH | Complete, 2026-10-27 |
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

## M1 deliverables

The [M1 implementation record](m1-results-and-reports.md) documents the typed
model, comparison compatibility checks, format conventions, and verification.
`internal/result`, `internal/report`, and `internal/input` now implement these
boundaries. TXT/MD/CSV consume one comparison model; native benchmark execution
and kernels retain their existing timing boundaries.

## M1 checkpoint scope

Introduce `internal/result` and `internal/report` as they become used. Keep the
`go test` subprocess while separating legacy parsing, observations, summaries,
and renderers. Retain run, environment, case, and per-case repetition identity;
preserve source timing precision. Keep compatibility fixtures passing while
adding cross-format fixtures for Feature 1. Terminal output stays readable TXT
regardless of the saved format.

M1 does not replace the measurement engine or rewrite kernels. M0's native
timings are observational evidence, never pass/fail thresholds.

## M2 deliverables

The [M2 implementation record](m2-runtime-catalog.md) describes the shared
metadata, six typed kernel bindings, descriptor-derived traffic, runtime listing,
compatibility checks, and retained before/after observations. The quick CLI,
native benchmark trees, report group selection, and legacy identity importer now
consume the catalogue. Kernel bodies and existing report formats are unchanged.

## M3 deliverables

The [M3 implementation record](m3-standalone.md) describes the compiled default
report engine, exact typed observations, versioned JSON evidence, validation and
metadata boundaries. Both commands share selectors/help/listing; Make, Go and
installed launchers use consistent flags. `-backend go-test` is an explicit
developer comparison path, never a fallback. Source and Go are no longer
runtime requirements for default installed reporting.

## M3 stopping point

M3 is complete. M4 worker isolation, cancellation recovery and artifact manifests
remain deferred pending an observed laboratory need. No M4 work or new F004
kernels are included in this checkpoint. The existing commands remain available;
there is no unified executable migration or public workload platform.

## Related operation backlog

[F004: operation and experiment patterns](../features/F004-operation-patterns.md)
reviews missing broadcast, predicate/selection, accumulation, movement and
representation experiments. It informs descriptor seams without expanding M2
into new kernels or changing the checkpoint order.
