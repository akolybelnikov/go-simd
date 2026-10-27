# go-simd

An experimental Go SIMD laboratory: seven operations and sixteen scalar/SIMD
implementations over generated `float32` slices, including addition, a compute
expression, broadcast scale/affine and predicate-based masking/selection. Run a quick
comparison or save repeated measurements as TXT, Markdown, or CSV. Results depend
on hardware, toolchain, size and background load; SIMD can be slower.

**M3:** both commands run as standalone binaries on a compatible local machine,
cloud VM or remote server. The default report backend needs neither Go nor the
source checkout at runtime. Reports stay on the machine where you run the CLI.

## Start here

From a source checkout in PowerShell:

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -list
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1024$' -count 3 -benchtime 100ms -format md
```

This bounded experiment saves a comparison and a `.measurements.json` companion
under `artifacts/addition/`. The full default sweep reaches 16,777,216 elements
and can take several minutes.

## Choose how to run

Flags are the same for installed commands, `go run` and Make. Only the launcher
changes. Make and `go run` need Go and the checkout. Installed binaries run from
any working directory.

| Task | Installed / on PATH | Go from checkout | Make from checkout |
| --- | --- | --- | --- |
| List operations and runtime context | `simdreport -list` | `go run ./cmd/simdreport -list` | `make list` |
| Quick comparison, all groups | `simdbench` | `go run ./cmd/simdbench` | `make run` |
| Quick addition comparison | `simdbench -group addition` | `go run ./cmd/simdbench -group addition` | `make run-addition` |
| Save a Markdown sweep | `simdreport -format md` | `go run ./cmd/simdreport -format md` | `make report ARGS="-format md"` |
| Report help | `simdreport -h` | `go run ./cmd/simdreport -h` | `make report ARGS="-h"` |
| Quick help | `simdbench -h` | `go run ./cmd/simdbench -h` | `make run ARGS="-h"` |

Use `-group computation` for the expression experiment. Every Make run, list
and report target forwards `ARGS`, including group shortcuts:

```bash
make help
make list ARGS="-group addition"
make run-computation ARGS="-list"
make report-addition COUNT=3 BENCHTIME=100ms ARGS="-format md -bench BenchmarkAddition/n=1024/Scalar"
```

Prefer direct commands for complex regex or paths needing careful shell quoting.

## Command reference

Both commands share these options:

| Option | Default | Meaning |
| --- | --- | --- |
| `-group` | `all` | Choose a group from the catalogue below |
| `-list` | `false` | Implemented operations and runtime context; no measurements or artifacts |
| `-addition`, `-computation` | `false` | Group shortcuts |
| `-function add\|compute\|all` | Unset | Deprecated group alias retained for compatibility |
| `-h`, `-help` | — | Readable usage, options and examples |

Explicit group selections must agree. False boolean aliases do not select a
group. There is no public type/input-file selector yet.

### Operation catalogue

`-list` prints a compact table; `-list -group <name>` shows parameters, expressions
and strategy details. Existing aliases remain for addition/computation.

| Group | Operation | Implementations |
| --- | --- | --- |
| addition | `a+b` | Scalar, SIMDPartial, SIMDTail, SIMDParallel |
| computation | `a*a+b*b+a*b` | Scalar, SIMD |
| scale | `alpha*x` | Scalar, SIMDPartial |
| affine | `float32(alpha*x)+beta`, unfused | Scalar, SIMDPartial |
| threshold | `x > threshold ? x : +0` | Scalar, SIMDPartial |
| selection | `x > threshold ? x : y` | Scalar, SIMDPartial |
| rangefilter | `lower < x && x < upper ? x : +0` | Scalar, SIMDPartial |

Both commands accept these pattern options. Each operation uses only its declared
parameters; exact float32 bits and relevant fixture identity survive in evidence.

| Option | Default | Applies to |
| --- | --- | --- |
| `-alpha` | `1.5` | Scale and affine multiplier |
| `-beta` | `-0.5` | Affine offset |
| `-threshold` | `0` | Threshold and selection |
| `-lower`, `-upper` | `-1`, `1` | Exclusive range bounds |
| `-pattern` | `alternating` | Predicate fixture: none, all, alternating, clustered, shuffled |
| `-seed` | `42` | Deterministic shuffled predicate arrangement |

```powershell
go run ./cmd/simdreport -group selection -threshold 0 -pattern shuffled -seed 123 -bench '^BenchmarkSelection$/^n=1024$' -count 3 -benchtime 100ms -format md
go run ./examples/patterns
```

SIMD broadcasts once per measured call. Affine keeps separate float32 multiply
and add rounding. Predicates use strict ordered comparisons; equality/NaN is
false, false masking yields +0 and selection preserves chosen operand bits.
Impossible true distributions fail early; `-pattern none` permits deliberate
all-false bounds. See [broadcast and predicate experiments](https://github.com/akolybelnikov/go-simd/wiki/13-Broadcast-and-predicate-experiments)
for examples, numerical rules, byte accounting and the remaining F004 backlog.

`simdbench` measures a fixed 1,048,576-element input and prints a quick table.
It has no report/count/timing flags. `simdreport` adds:

| Option | Default | Meaning |
| --- | --- | --- |
| `-bench` | `.` | Go-style regex over benchmark / size / variant |
| `-count` | `5` | Independent measurements per selected case |
| `-benchtime` | `100ms` | Positive duration or fixed iteration count, e.g. `10x` |
| `-format` | `txt` | Saved summary: `txt`, `md`, `csv`; terminal stays text |
| `-output` | `artifacts` | Output root, relative to the current directory unless absolute |
| `-backend` | `compiled` | Standalone backend; `go-test` explicitly selects the developer backend |

Names are case-sensitive. For example,
`-bench '^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial)$'` selects two cases.
Regex selects existing sizes, not arbitrary new sizes. With `-group all`, it
must match cases in every selected group. Invalid timing, regex, format, backend or
selectors fail before artifacts are created.

## Build or install

Build hosts need Go **1.27.3** with `GOEXPERIMENT=simd`. GNU Make plus a POSIX
shell is optional. A compatible target binary needs neither Go nor the
`GOEXPERIMENT` environment setting at runtime.

```powershell
$env:GOEXPERIMENT = 'simd'
go build -o ./bin/ ./cmd/simdbench ./cmd/simdreport
go install ./cmd/simdbench ./cmd/simdreport
```

Make equivalents: `make build` and `make install`. Installation uses `GOBIN`,
or the Go workspace's `bin` directory when unset; add it to PATH. Windows builds
produce `bin/simdbench.exe` and `bin/simdreport.exe`; POSIX builds omit `.exe`.
Copy compatible binaries to a server or invoke them by absolute path. A prebuilt
release download/package-manager installer is not supplied yet.

See [installation and remote use](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry#build-install-and-run-remotely)
for cross-compilation, copying binaries and SSH examples. Windows/amd64 and
Linux/amd64 execution are covered by CI. Run `-list` on the target to inspect
its vector width, emulation state and GOMAXPROCS.

## Reports and evidence

Every completed group saves a summary and matching evidence:

- Default `compiled`: `.measurements.json`, schema `go-simd.measurements/v1`,
  retains exact duration/iterations, observations, case/input identities,
  allocation counters and runtime/build facts.
- Explicit `-backend go-test`: legacy `.raw` text. This developer option needs
  Go on PATH and the repository root. There is no automatic fallback.

TXT/Markdown include environment metadata below the table. CSV contains numeric
rows; metadata stays in the companion. Negative change means less time than
scalar. Missing baselines stay unavailable. Throughput uses logical work bytes,
not measured DRAM traffic. CPU model detection is currently `unknown`; source
revision/modified state are recorded when embedded by the Go build.

Keep backend/harness identities separate when comparing results. Failed or empty
compiled measurements retain a failed capture and no successful group summary.
Abrupt termination recovery and worker manifests remain later work.

## Develop and learn

```powershell
go test ./... -count=1
go vet ./...
gofmt -l cmd internal examples
go run ./examples/basic
```

Make equivalents include `make check`, `make test`, `make vet`, `make fmt` and
`make fmt-check`. `make clean` removes generated reports and Go build artifacts.
The [complete command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry)
lists every Make target/variable, CLI flag and native Go benchmark control.

The [progressive Wiki](https://github.com/akolybelnikov/go-simd/wiki) includes
kernel examples, recipes, numerical contracts, troubleshooting and credited SIMD
reading. Editable sources live in `docs/wiki/`. See the [changelog](CHANGELOG.md)
and [migration record](docs/migration/README.md). F004's
[first implementation slice](docs/features/F004-broadcast-predicates.md) is complete;
reductions, FMA, clamp, packed-mask export and new types remain planned.

Shared boundaries live in `internal/cli`, `internal/engine`, `internal/platform`,
`internal/catalog`, `internal/input`, `internal/evidence`, `internal/result` and `internal/report`.
Kernels remain in [addition](internal/addition/README.md) and
[computation](internal/computation/README.md), with new kernels in
[patterns](internal/patterns/README.md). They support disjoint output and
exact in-place aliases; partial overlap is unsupported and length errors leave
output unchanged. These internal packages are not a stable public API.
