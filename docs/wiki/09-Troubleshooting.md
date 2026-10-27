# 09 — Troubleshooting

Inspect the exact command and retain error/evidence output. Installed commands
run anywhere with the default compiled backend. Go/Make launchers and the
explicit `-backend go-test` developer option need Go and the source checkout.

## Setup and build failures

| Symptom | Likely boundary | Next step |
| --- | --- | --- |
| `unknown GOEXPERIMENT simd` or missing `simd` package | Selected Go toolchain does not provide this experiment | Check `go version`, `go env GOTOOLCHAIN`, and `go env GOEXPERIMENT`; use a toolchain satisfying this module and experiment |
| Toolchain download/version failure | Go selection, network, or local installation | Inspect the module's Go directive and your selected installation; changing `go.mod` does not implement SIMD in another compiler |
| `cannot find main module` / missing `go.mod` | Wrong working directory | Change to the checkout root |
| Developer backend cannot find `./internal/addition` or computation | `-backend go-test` runs outside its checkout | Use the default compiled backend, or run the developer option from the repository root |
| `exec: go ... not found` | Explicit developer backend requires Go | Use the default compiled backend, or provide Go/source/experiment for go-test |
| `use of internal package ... not allowed` | Import from an external module | Use the module-local example; there is no public import API yet |
| `make`, `printf`, shell syntax, or quoting failures | Make/POSIX shell unavailable or incompatible | Use direct PowerShell Go commands from the registry |

## Rejected options

- Both commands support `-addition`, `-computation` and `-function` aliases;
  conflicting selections fail consistently.
- `simdbench -format csv` is not supported; report format selection belongs to
  `simdreport`.
- `-workers`, `-input`, `-type`, `--version`, and TUI/advisor subcommands are not
  implemented at this baseline.
- `-format html`, uppercase `CSV`, and blank formats are rejected.
- `-count 0` and negative counts are rejected before measuring.
- Conflicting quick selectors such as `-group all -addition` are rejected.
- Positional arguments are rejected; supply the appropriate named flag.

Use `go run ./cmd/simdbench -h` and `go run ./cmd/simdreport -h` to inspect the
installed source's actual command definitions.

Both commands support `-list` from M3 and produce the same catalogue. Their built
binaries need neither Go nor source; `go run` still does. Only the explicit
developer report backend launches Go. Listing only the current float32 pairs
is expected: it does not advertise planned SDK methods as implemented kernels.

## Pattern input errors

Unknown patterns and unrepresentable float32 parameter text fail before artifacts.
A pattern needing true lanes also fails for a NaN/+Inf threshold or an empty
strict interval. Use `-pattern none` for those deliberate all-false experiments.
Unrelated scalar flags do not affect an operation's case identity. Inspect the
selected context with `-list -group selection -threshold 0 -pattern shuffled`.

## No benchmark measurements found

Either backend can have no matching benchmark case. Verify:

1. Top-level name: `BenchmarkAddition` or `BenchmarkComputation`.
2. Variant name: addition uses `SIMDPartial`/`SIMDTail`/`SIMDParallel`; computation
   uses `SIMD`.
3. The size is actually in the generated sweep.
4. Slash-separated regex components and shell quoting are intact.
5. `-group all` is not selecting additional groups your regex excludes.

Start with this known small selection:

```powershell
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1$/^Scalar$' -count 1 -benchtime 1x
```

Then widen one component at a time. A selection regex cannot add sizes that
the benchmark never registers. Save the retained raw output when diagnosing
an empty selection or runner failure.

### Why size 9 can disappear on another machine

Size 9 is not a power of two. It appears around width 8 (`w+1`), but is not
registered around width 16. Use powers of two, such as 8 and 16, for a portable
two-size smoke check, or derive tail examples from the observed lane count.

## The table has em dashes or empty percentage fields

These can be correct: scalar was filtered out, a variant was not measured at
that size, or the measured scalar time was nonpositive. The tool deliberately
does not replace missing evidence with zero. CSV represents missing cells as
empty strings; TXT/Markdown use an em dash.

## The report seems slow or quiet

`-benchtime` applies per measurement, and `-count` repeats every selected case.
The default covers all seven groups and many sizes up to 16,777,216 elements. That
largest point needs about 192 MiB for its three slices, before other overhead.
The command prints group progress and buffers the child output until that group
finishes. Narrow the benchmark selection for early exploration.

`1x` shortens measurement work but does not eliminate large input allocation.
Use a small selected size when testing setup or running with limited memory.

## Parallel SIMD is slower

That can be a valid observation. The implementation creates and joins workers
on each call, and each chunk can have a scalar tail. Inspect payload size,
GOMAXPROCS, raw variability, and allocation counters. There is no persistent pool
or automatic threshold. Preserve the negative result instead of assuming a bug
from timing alone.

## Race build fails before tests run

A `runtime/cgo` or C-compiler error during `go test -race ./...` can be a host
toolchain issue. Record the exact diagnostic and C compiler, then separate it
from an assertion failure in a project test. The project's Linux CI runs a full
race check; current Windows CI runs tests/vet/formatting and filtered benchmarks.
Race-instrumented timings should not be compared with normal benchmark timings.

## Mixed or ambiguous measurement contexts

The internal model rejects attempts to create one comparison from incompatible
operations, environments, backend/harness identities, or GOMAXPROCS settings.
Conflicting per-size input fingerprints and duplicate strategy/size entries
with different worker configurations also fail. Keep those experiments in
separate tables instead of merging names and losing identity.

## Report a useful issue

Include the exact command, Go/toolchain settings, OS/architecture/CPU context,
source revision/local changes, the error, and relevant raw observations. State
what you expected and whether the problem reproduces with a tiny `1x` case.
Review machine metadata before choosing to post your own files publicly.
