# 03 — Command catalogue

Baseline: M3 / F004 slice 1, reviewed 2026-10-27. Installed binaries run from any directory.
Go/Make examples run from the checkout with `GOEXPERIMENT=simd`. The same CLI
flags work through every launcher.

## Pick a task

| Task | Installed / on PATH | Go from checkout | Make from checkout |
| --- | --- | --- | --- |
| List operations and runtime context | `simdreport -list` | `go run ./cmd/simdreport -list` | `make list` |
| Quick comparison | `simdbench` | `go run ./cmd/simdbench` | `make run` |
| Quick addition | `simdbench -group addition` | `go run ./cmd/simdbench -group addition` | `make run-addition` |
| Quick expression | `simdbench -group computation` | `go run ./cmd/simdbench -group computation` | `make run-computation` |
| Markdown sweep | `simdreport -format md` | `go run ./cmd/simdreport -format md` | `make report ARGS="-format md"` |
| Report help | `simdreport -h` | `go run ./cmd/simdreport -h` | `make report ARGS="-h"` |
| Quick help | `simdbench -h` | `go run ./cmd/simdbench -h` | `make run ARGS="-h"` |

Bare executable names assume their directory is on PATH. Use an absolute path,
`./simdreport` on POSIX, or `.\simdreport.exe` in PowerShell otherwise.
`make help` prints the short target catalogue; each CLI's `-h` prints options
and examples. Listing is identical through either executable.

## Shared options: both commands

| Flag | Default | Meaning / example |
| --- | --- | --- |
| `-group` | `all` | `all` or any group in the operation table below; `add` and `compute` remain aliases |
| `-addition` | `false` | Boolean alias for addition: `-addition` |
| `-computation` | `false` | Boolean alias for computation: `-computation=true` |
| `-function` | Unset | Deprecated alias: `add`, `compute`, `all`; canonical group names also accepted |
| `-list` | `false` | Show registered kernels and live runtime context without measuring |
| `-h`, `-help` | — | Help to stderr; success without measurements |

Explicit selectors must agree. `-group addition -function add` is accepted;
`-group all -addition` is rejected. False boolean aliases do not select a
group. No explicit selection means all seven groups. Positional arguments are
rejected. Repeating a flag uses Go's normal last-value rule.

`-list` prints runtime context and a compact seven-row operation table.
Add `-group <name>` for expression/semantics, operands, scalar values/bits,
fixture identity, strategy labels/IDs and worker policy.
Logical traffic is 8 bytes per element for unary patterns and 12 for addition, computation and two-array selection.
Listing allocates no benchmark fixtures, runs no kernels/subprocesses, and saves
no artifacts. It advertises only implemented kernels, not proposed SDK patterns.
Runtime emulation context does not guarantee acceleration or a speedup.

```powershell
simdreport -list
simdbench -list -group addition
simdreport -list -function compute
```

`-list=false` selects normal measurement. Report-specific arguments are still
validated by `simdreport` when listing. The listing is human-readable diagnostic
text, not a versioned machine-data export.

## Pattern options: both commands

| Flag | Default | Meaning |
| --- | --- | --- |
| `-alpha` | `1.5` | Scale/affine float32 multiplier |
| `-beta` | `-0.5` | Affine float32 offset, without FMA |
| `-threshold` | `0` | Threshold/selection strict comparison |
| `-lower`, `-upper` | `-1`, `1` | Exclusive rangefilter bounds |
| `-pattern` | `alternating` | none/all/alternating/clustered/shuffled predicate arrangement |
| `-seed` | `42` | Deterministic shuffled pattern seed |

Scalars parse to float32 and are recorded as exact bits. Settings apply only to
operations declaring them. Predicate fixtures guarantee none/all truth or mixed
arrangements; shuffled is approximately 50%, with the seed in input identity.
Impossible true distributions fail before fixtures/artifacts; `-pattern none`
permits all-false bounds. See [the worked pattern guide](https://github.com/akolybelnikov/go-simd/wiki/13-Broadcast-and-predicate-experiments)
for NaN/Inf/signed-zero, alias and rounding rules.

## Quick comparison: simdbench

Each strategy uses **1,048,576 generated float32 elements**. Go calibrates one
`testing.Benchmark` result per strategy, with allocation counters and change
versus scalar. Displayed ns/op retains Go's integer `NsPerOp()` presentation.

The shared flags above are its entire CLI. It has no size/count/benchtime/report
format/output/backend flag. Errors exit **2**; success/help exit **0**.

## Repeated reports: simdreport

| Additional flag | Default | Meaning / example |
| --- | --- | --- |
| `-bench` | `.` | Go-style slash-separated regex; `^BenchmarkAddition$/^n=1024$/^Scalar$` |
| `-count` | `5` | Positive independent measurement count per selected case |
| `-benchtime` | `100ms` | Positive duration or iteration count: `500ms`, `10x` |
| `-format` | `txt` | Exactly `txt`, `md`, `csv`; terminal stays readable text |
| `-output` | `artifacts` | Output root, e.g. `-output 'results/my run'`; group directories are added |
| `-backend` | `compiled` | `compiled` standalone, or explicit `go-test` developer backend |

Flags accept `-name value` or `-name=value`. `-count` counts measurements,
not kernel calls. Duration mode calibrates iteration counts; `10x` reports ten
iterations, with possible extra calibration/setup calls. `1x` is a plumbing
check, not a sound performance conclusion. Nonpositive/invalid timing, regex,
count, format, backend and selectors fail before artifacts.

Each completed group saves a summary and `.measurements.json` under the chosen
root. The compiled backend captures exact duration/iterations and build/runtime
facts without invoking Go. Explicit `-backend go-test` requires Go, source and
the experiment setting, retains legacy `.raw`, and never silently falls back.
See [reports and evidence](https://github.com/akolybelnikov/go-simd/wiki/05-Reports-and-evidence)
for schema, precision and failure behavior. Errors exit **1**; success/help **0**.
The outer `go run` process can also print `exit status 1`.

### Names and selection

| Group | Top-level benchmark | Variant component |
| --- | --- | --- |
| `addition` | `BenchmarkAddition` | `Scalar`, `SIMDPartial`, `SIMDTail`, `SIMDParallel` |
| `computation` | `BenchmarkComputation` | `Scalar`, `SIMD` |
| `scale` | `BenchmarkScale` | `Scalar`, `SIMDPartial` |
| `affine` | `BenchmarkAffine` | `Scalar`, `SIMDPartial` |
| `threshold` | `BenchmarkThreshold` | `Scalar`, `SIMDPartial` |
| `selection` | `BenchmarkSelection` | `Scalar`, `SIMDPartial` |
| `rangefilter` | `BenchmarkRange` | `Scalar`, `SIMDPartial` |

New native benchmark wrappers share `./internal/patterns`. The developer report
backend restricts it to the requested group's validated cases and records exact
input/parameter context. Direct native commands use defaults; choose the report
CLI with `-backend go-test` for recorded custom parameters.

The selector name is `BenchmarkAddition/n=1024/SIMDPartial`. A numeric suffix in
native output, e.g. `-16`, is GOMAXPROCS context, not a variant or lane count.
The sweep contains powers of two from 1 through 16,777,216 plus positive
sizes around one and two vector widths. Regex selects existing cases only.

| Intent | Group | Expression |
| --- | --- | --- |
| All addition | `addition` | `.` |
| Exact scalar case | `addition` | `^BenchmarkAddition$/^n=1024$/^Scalar$` |
| Serial strategies, three sizes | `addition` | `^BenchmarkAddition$/^n=(128\|1024\|8192)$/^(Scalar\|SIMDPartial\|SIMDTail)$` |
| Scalar and parallel | `addition` | `^BenchmarkAddition$/^n=1048576$/^(Scalar\|SIMDParallel)$` |
| Expression strategies | `computation` | `^BenchmarkComputation$/^n=1024$/^(Scalar\|SIMD)$` |
| All groups, portable small sizes | `all` | `^Benchmark.*$/^n=(8\|16)$/^(Scalar\|SIMD.*)$` |

Matching preserves Go's component regex, top-level alternatives, parentheses,
character classes, escapes and unanchored semantics. Unanchored `n=1` also
matches `n=16` and `n=128`. An empty expression selects nothing. Paths deeper
than the registered benchmark/size/variant leaves have no matching cases.
With `-group all`, match every selected group or the unmatched group will fail with
`no benchmark measurements found`, retaining failed evidence.

### Copy a bounded first experiment

```powershell
simdreport -group addition -bench '^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial|SIMDTail)$' -count 3 -benchtime 100ms -format md
```

For a source run, replace `simdreport` with `go run ./cmd/simdreport` and keep
the same flags. Filtered SIMD-only tables leave baseline changes unavailable.
The full default sweep can take minutes and allocates about 192 MiB for three
slices at its largest point.

## Build, install and run remotely

On a build host, enable the experiment and build or install:

```powershell
$env:GOEXPERIMENT = 'simd'
go build -o ./bin/ ./cmd/simdbench ./cmd/simdreport
go install ./cmd/simdbench ./cmd/simdreport
.\bin\simdreport.exe -list
```

On POSIX shells use `export GOEXPERIMENT=simd` and `./bin/simdreport`.
`make build` and `make install` do the same jobs. `go install` uses `GOBIN`,
otherwise the Go workspace's `bin` directory; put that directory on PATH.
The checkout's current revision is installed by these commands. No prebuilt
release download or package-manager installer is supplied yet.

To build Linux/amd64 from a POSIX build shell and copy to a server:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOEXPERIMENT=simd go build -o dist/linux-amd64/ ./cmd/simdbench ./cmd/simdreport
ssh user@server 'mkdir -p ~/bin'
scp dist/linux-amd64/simdbench dist/linux-amd64/simdreport user@server:~/bin/
ssh user@server '~/bin/simdreport -list'
ssh user@server '~/bin/simdreport -group addition -bench BenchmarkAddition/n=1024/Scalar -count 3 -benchtime 100ms -format md'
```

Results are created on the server under its current directory's `artifacts/`;
use `-output` to choose a different directory and copy evidence back with scp.
The target needs a matching OS/architecture binary, not Go, Make, the checkout
or the GOEXPERIMENT environment setting. Inspect `-list` on that target.
Windows/amd64 and Linux/amd64 execute in CI; cross-compilation does not prove
execution or performance on other targets. Remote provisioning/job management
is outside these commands. There is no unified `simd` executable or
`--version` flag at this milestone.

## Complete Make catalogue

Make requires GNU Make and a POSIX shell. All run/list/report targets forward
`ARGS`. Report targets also supply `COUNT`, `BENCHTIME`, `ARTIFACT_ROOT`.

| Target | Purpose |
| --- | --- |
| `help` | Readable target catalogue |
| `list` | Quick CLI's shared `-list` surface |
| `run` | Quick comparison, all groups |
| `run-addition` | Quick addition; forwards ARGS |
| `run-computation` | Quick expression; forwards ARGS |
| `report` | Reports, all groups |
| `report-addition` | Addition report |
| `report-computation` | Expression report |
| `build` | Build both commands into BIN_DIR |
| `install` | Install both commands via go install |
| `test` | Full Go tests |
| `vet` | Go vet |
| `fmt` | Format cmd and internal source |
| `fmt-check` | Check formatting without edits |
| `check` | Formatting check, tests, vet |
| `clean` | Go clean and recursive removal of ARTIFACT_ROOT |

| Make variable | Default | Meaning |
| --- | --- | --- |
| `GOEXPERIMENT` | `simd` | Exported build experiment |
| `GO` | `go` | Go executable used by Make; developer backend finds go on PATH |
| `ARGS` | Empty | Extra flags for every run/list/report target |
| `BENCHTIME` | `100ms` | Report target measurement timing |
| `COUNT` | `5` | Report target repetitions |
| `ARTIFACT_ROOT` | `artifacts` | Report output and directory deleted by clean |
| `BIN_DIR` | `bin` | Build output directory |

```bash
make list ARGS="-group addition"
make run-addition ARGS="-list"
make report-addition COUNT=3 BENCHTIME=100ms ARGS="-format csv -bench BenchmarkAddition/n=1024/Scalar"
make build BIN_DIR=bin
make install
```

Prefer direct Go/installed commands for complex quoting, anchored regex with
`$` or paths with spaces. Make introduces shell quoting and dollar expansion.
`make clean` deletes saved reports recursively; preserve needed evidence first.

## Native Go and developer tools

Direct new-operation benchmarks:
`go test ./internal/patterns -run='^$' -bench='^BenchmarkScale$/^n=1024$' -benchmem -benchtime=100ms`.
They use default scalar/fixture options. For custom recorded settings, prefer
`simdreport -backend go-test -group scale -alpha 2 ...` from the checkout.

These need Go/source and GOEXPERIMENT. They are not installed CLI options.

```powershell
go test ./internal/addition ./internal/computation -list '^Benchmark'
go test ./internal/addition -run='^$' -bench='^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial)$' -benchmem -count=5 -benchtime=100ms
go test ./internal/addition -run='^$' -bench='^BenchmarkAddition$/^n=1048576$/^(Scalar|SIMDParallel)$' -benchmem -count=3 -benchtime=100ms -cpu=1,4
go test ./... -count=1
go test -race ./...
go vet ./...
gofmt -l cmd internal examples
gofmt -w examples/basic/main.go
go run ./examples/basic
go run ./examples/patterns
go doc ./internal/addition
go doc ./internal/computation
go doc ./internal/result
go doc ./internal/report
go version
go env -json GOEXPERIMENT GOAMD64 GOARM64 GOTOOLCHAIN
git rev-parse HEAD
git status --short
```

| Native Go flag | Purpose |
| --- | --- |
| `-run '^$'` | Skip ordinary tests |
| `-bench` | Benchmark name/component selection |
| `-benchmem` | Allocation counters |
| `-count` | Repeat benchmarks; -count=1 also avoids cached test results |
| `-benchtime` | Duration or Nx |
| `-cpu 1,4` | Native GOMAXPROCS sweep |
| `-list` | Matching top-level test/benchmark names |
| `-race` | Correctness instrumentation; timings are not normal performance data |

`gofmt -w` edits source; `-l` only lists formatting issues. Race builds need a
suitable C toolchain. Native output can be saved using redirection or
PowerShell `Tee-Object`; add source/environment notes.

## Environment settings

| Setting | Scope |
| --- | --- |
| `GOEXPERIMENT=simd` | Required for builds, go run/tests and explicit go-test backend; not for running standalone binaries |
| `GOMAXPROCS` | Runtime concurrency and current parallel addition's worker bound |
| `GOBIN` | Go installation destination; must be absolute when set |
| `GOTOOLCHAIN` | Build-host Go toolchain selection |
| `GOOS`, `GOARCH`, `CGO_ENABLED` | Cross-compilation/build settings |
| `GOAMD64`, `GOARM64` | Target assumptions embedded in build information when supplied |

There is no project `-workers` flag. Use GOMAXPROCS in separate processes or
native `-cpu` sweeps. Do not change a string environment variable inside an
already running process and assume its initialized runtime/SIMD state changed.

Next: [operation guide](https://github.com/akolybelnikov/go-simd/wiki/04-Operation-guide)
and [experiment recipes](https://github.com/akolybelnikov/go-simd/wiki/07-Experiment-recipes).
