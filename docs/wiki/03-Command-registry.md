# 03 — Complete command registry

Baseline: M1 / Feature 1. Run examples from the repository root with
`GOEXPERIMENT=simd` enabled. This page distinguishes project flags from Go-tool
flags; options supported by `go test` are not automatically project CLI options.

## Command surfaces

| Command | Purpose | Saves files? |
| --- | --- | --- |
| `go run ./cmd/simdbench` | Quick fixed-size comparison, both groups | No |
| `go run ./cmd/simdreport` | Repeated size sweep and summary export, both groups | Yes: selected report + raw per group |
| `go run ./examples/basic` | Deterministic six-kernel usage example | No |
| `go test ./...` | Tests, including short real CLI export tests | Test temporary files |
| `go test ... -bench ...` | Native Go benchmark workflow | Only if you redirect output |
| `go vet ./...` | Static checks | No report artifacts |
| `gofmt -l ...` / `gofmt -w ...` | Inspect / apply Go formatting | `-w` edits source |
| `go doc ...` | Inspect local package documentation | No |
| `go build ...` | Compile local commands | Writes binaries |
| `make <target>` | Convenience wrapper described below | Depends on target |

## `simdbench`: quick comparison

```powershell
go run ./cmd/simdbench
go run ./cmd/simdbench -group addition
go run ./cmd/simdbench -group computation
go run ./cmd/simdbench -h
```

Every variant uses **1,048,576 float32 elements**. There is no quick-CLI size,
count, duration, report-format, or output-file flag. It uses compiled
`testing.Benchmark` calibration and prints one result per strategy, with
allocation counters and change versus scalar. Its displayed ns/op comes from
Go's integer `NsPerOp()` result, unlike the report importer's parsed precision.

### All quick-CLI flags

| Flag | Default | Accepted values / behavior | Example |
| --- | --- | --- | --- |
| `-group` | Effective default `all` | Canonical values: `addition`, `computation`, `all` | `-group addition` |
| `-function` | Unset | Deprecated alias: `add`, `compute`, `all`; parser also accepts canonical group names | `-function add` |
| `-addition` | `false` | Boolean alias selecting addition when true | `-addition` |
| `-computation` | `false` | Boolean alias selecting computation when true | `-computation=true` |
| `-h`, `-help` | — | Print help to stderr and succeed without measurements | `-h` |

The quick parser also normalizes `add`/`compute` when supplied through `-group`;
use the canonical names in new commands. Explicit selections must agree:

```powershell
# Accepted: the selectors mean the same group.
go run ./cmd/simdbench -group addition -function add

# Rejected: all and addition are conflicting explicit selections.
go run ./cmd/simdbench -group all -addition
```

Setting a boolean alias to false does not select that group. With no true aliases
or explicit group/function selection, both groups run. Positional arguments are
rejected. Failures from the built command exit **2**; help exits **0**.

## `simdreport`: sweep and export

```powershell
go run ./cmd/simdreport -group addition -count 5 -benchtime 100ms -format md
```

### All report-CLI flags

| Flag | Default | Meaning | Example |
| --- | --- | --- | --- |
| `-group` | `all` | Exactly `addition`, `computation`, or `all` | `-group computation` |
| `-count` | `5` | Positive number of repetitions per selected case | `-count 10` |
| `-benchtime` | `100ms` | Go benchmark target duration or fixed iteration syntax | `-benchtime 500ms`, `-benchtime 10x` |
| `-bench` | `.` | Go slash-separated regular-expression selection | `-bench '^BenchmarkAddition$/^n=1024$/^Scalar$'` |
| `-output` | `artifacts` | Output root; group directories are added below it | `-output 'results/my run'` |
| `-format` | `txt` | Saved report format: exactly `txt`, `md`, or `csv` | `-format csv` |
| `-h`, `-help` | — | Print help to stderr and succeed without measurements | `-help` |

Flags may use `-name value` or `-name=value`. Keep quoted regex expressions as
single shell arguments. Group shortcuts from `simdbench` are not report flags.

`-count 5` repeats measurements; it does not mean five kernel calls. Duration
mode calibrates iteration counts. In `10x` mode the reported measurement uses
ten iterations; Go may perform additional calibration/setup invocations.
`1x` is useful for plumbing checks and poor for performance conclusions.

`-format` changes the saved summary only. Terminal output is always a readable
text table. Every successful group produces `.raw` plus one `.txt`, `.md`, or
`.csv` file. Metadata remains available in raw even for CSV.

The built report command exits **1** on failure and **0** on success/help. The
outer `go run` process may additionally print `exit status 1`; distinguish Go
wrapper status from the built program's own exit code in scripts.

### Benchmark names and variants

Names are case-sensitive:

| Group | Top-level name | Variant component |
| --- | --- | --- |
| `addition` | `BenchmarkAddition` | `Scalar`, `SIMDPartial`, `SIMDTail`, `SIMDParallel` |
| `computation` | `BenchmarkComputation` | `Scalar`, `SIMD` |

The shape is `BenchmarkAddition/n=1024/SIMDPartial-16`. The numeric suffix is
Go's concurrency annotation, not a variant name or SIMD lane count. Depending
on concurrency, the suffix can be absent.

The size sweep contains powers of two from 1 to 16,777,216 plus sizes immediately
around one and two native vector widths. You can select existing sizes; a regex
does not create a new arbitrary size.

### Selection recipes

| Intent | `-group` | `-bench` expression |
| --- | --- | --- |
| All addition cases | `addition` | `.` |
| One exact scalar case | `addition` | `^BenchmarkAddition$/^n=1024$/^Scalar$` |
| Three sizes, serial variants | `addition` | `^BenchmarkAddition$/^n=(128|1024|8192)$/^(Scalar|SIMDPartial|SIMDTail)$` |
| Partial-vector only | `addition` | `^BenchmarkAddition$/^n=1024$/^SIMDPartial$` |
| Scalar and parallel | `addition` | `^BenchmarkAddition$/^n=1048576$/^(Scalar|SIMDParallel)$` |
| Compute scalar and SIMD | `computation` | `^BenchmarkComputation$/^n=1024$/^(Scalar|SIMD)$` |
| Both groups, two sizes | `all` | `^Benchmark(Addition|Computation)$/^n=(8|9)$/^(Scalar|SIMD.*)$` |

Go matches slash-separated regex components. Anchors matter: unanchored `n=1`
also matches `n=16`, `n=128`, and other names containing that substring.
`-group all` runs each group separately; if your expression matches only addition,
the subsequent computation group has no observations and the command fails.
Use the matching group or an expression covering both.

An empty selection leaves raw evidence and reports `no benchmark measurements
found`. With filtered SIMD-only results, changes are unavailable because scalar
was not measured; TXT/Markdown show an em dash and CSV uses an empty field.

### Ready-to-copy exports

```powershell
# TXT: default behavior, selected cases, repeated measurements.
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial|SIMDTail)$' -count 5 -benchtime 100ms

# Markdown: same experiment with a shareable table and metadata.
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial|SIMDTail)$' -count 5 -benchtime 100ms -format md

# CSV: exact numeric cells for downstream tables or charts.
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial|SIMDTail)$' -count 5 -benchtime 100ms -format csv -output 'results/my run'

# Full default sweep: both groups; can take several minutes.
go run ./cmd/simdreport
```

Different invocations measure again; changing `-format` does not re-render an
existing raw file. An offline analysis/import CLI is not implemented yet.

## Native Go benchmark commands

```powershell
# List top-level benchmark names, not every nested size/variant.
go test ./internal/addition ./internal/computation -list '^Benchmark'

# Native benchmark output with memory counters.
go test ./internal/addition -run='^$' -bench='^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial|SIMDTail)$' -benchmem -count=5 -benchtime=100ms

# Native concurrency sweep: use this interface for -cpu.
go test ./internal/addition -run='^$' -bench='^BenchmarkAddition$/^n=1048576$/^(Scalar|SIMDParallel)$' -benchmem -count=5 -benchtime=100ms -cpu=1,4
```

| Native flag | Meaning |
| --- | --- |
| `-run '^$'` | Skip ordinary tests for a focused benchmark invocation |
| `-bench <regex>` | Select benchmark names/components |
| `-benchmem` | Print allocation counters |
| `-count N` | Repeat the selected benchmarks |
| `-benchtime T` | Duration target or `Nx` iteration mode |
| `-cpu 1,4` | Run with each listed GOMAXPROCS value |
| `-list <regex>` | List matching top-level test/benchmark names |
| `-race` | Enable race instrumentation for correctness checking; do not treat its timings as normal performance data |

These are Go tool flags, not additional `simdreport` options. Native output can
be saved with shell redirection or PowerShell `Tee-Object`; add your own source
and environment notes when retaining it.

## Correctness, formatting, and source inspection

```powershell
go test ./...
go test ./... -count=1
go test ./internal/addition -run '^TestImplementations$'
go vet ./...
go test -race ./...
gofmt -l cmd internal examples
gofmt -w examples/basic/main.go
go doc ./internal/addition
go doc ./internal/computation
go doc ./internal/result
go doc ./internal/report
go version
go env -json GOEXPERIMENT GOAMD64 GOARM64 GOTOOLCHAIN
git rev-parse HEAD
git status --short
```

`-count=1` avoids reuse of cached test results. `gofmt -w` edits source; use `-l`
to inspect first. Race checks need a suitable C toolchain. Use source and raw
evidence to diagnose behavior before interpreting a noisy timing change.

## Build local binaries

PowerShell:

```powershell
New-Item -ItemType Directory -Force bin | Out-Null
go build -o ./bin/simdbench.exe ./cmd/simdbench
go build -o ./bin/simdreport.exe ./cmd/simdreport
.\bin\simdbench.exe -group addition
.\bin\simdreport.exe -group addition -bench '^BenchmarkAddition$/^n=1$/^Scalar$' -count 1 -benchtime 1x -format csv
```

POSIX shell:

```bash
mkdir -p bin
go build -o ./bin/simdbench ./cmd/simdbench
go build -o ./bin/simdreport ./cmd/simdreport
./bin/simdbench -group addition
```

The quick binary uses compiled measurement facilities. The report binary still
launches `go env` and repository-relative `go test`; building it does not remove
its runtime Go/source dependency. Keep the experiment enabled for its child Go
commands. No `--version`, capabilities subcommand, or unified `simd` executable
is implemented at this baseline.

## Every Make target

GNU Make targets are optional; they require a POSIX shell even on Windows.

| Target | Action | Argument forwarding |
| --- | --- | --- |
| `make help` | Show common invocations | None |
| `make run` | Run `simdbench`, both groups by default | `ARGS` |
| `make run-addition` | Run quick addition comparison | Fixed group; no `ARGS` forwarding |
| `make run-computation` | Run quick computation comparison | Fixed group; no `ARGS` forwarding |
| `make report` | Run report CLI, both groups by default | Count, benchtime, output root, `ARGS` |
| `make report-addition` | Run addition report | Count, benchtime, output root, `ARGS` |
| `make report-computation` | Run computation report | Count, benchtime, output root, `ARGS` |
| `make test` | Run `go test ./...` | None |
| `make vet` | Run `go vet ./...` | None |
| `make fmt` | Format `cmd` and `internal` in place | None |
| `make fmt-check` | Fail when those directories need formatting | None |
| `make check` | Formatting check, tests, and vet | None |
| `make clean` | `go clean ./...`, then remove configured report root | Destructive for reports: see below |

### Every Make variable

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOEXPERIMENT` | `simd` | Exported experiment setting |
| `GO` | `go` | Go executable used by Make's own commands; the report process still finds `go` on PATH |
| `ARGS` | Empty | Additional CLI arguments where forwarding is supported |
| `BENCHTIME` | `100ms` | Report measurement duration/iteration mode |
| `COUNT` | `5` | Report repetitions |
| `ARTIFACT_ROOT` | `artifacts` | Report root and the directory removed by `make clean` |

```bash
make run ARGS="-group computation"
make report-addition COUNT=5 BENCHTIME=100ms ARGS="-format md -bench BenchmarkAddition/n=1024/Scalar"
make report-computation COUNT=5 BENCHTIME=100ms ARGS="-format csv"
```

Prefer direct Go commands for complex regex alternatives or paths with spaces;
Make adds another layer of shell quoting and dollar-sign expansion. In particular,
prefer direct Go commands for anchored expressions containing `$`.
`make clean` recursively deletes
`ARTIFACT_ROOT`, including saved raw observations. Use a dedicated generated
directory and preserve evidence you want to retain before cleaning.

## Environment controls

| Setting | Supported use |
| --- | --- |
| `GOEXPERIMENT=simd` | Required when building/running code that imports `simd` |
| `GOMAXPROCS` | Bounds the runtime concurrency used by current parallel addition |
| `GOTOOLCHAIN` | Go's toolchain-selection setting; record it with results |
| `GOAMD64`, `GOARM64` | Target-specific Go settings recorded by reports; changing them changes the experimental environment |

There is no project `-workers` flag. Use `GOMAXPROCS` for separate report runs,
or native `go test -cpu` for concurrency sweeps. Do not change an environment
string inside a running process and assume already-initialized SIMD state changed.

Next: [operation guide](https://github.com/akolybelnikov/go-simd/wiki/04-Operation-guide)
and [experiment recipes](https://github.com/akolybelnikov/go-simd/wiki/07-Experiment-recipes).
