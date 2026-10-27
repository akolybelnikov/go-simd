# go-simd

An experimental Go project comparing scalar and SIMD operations on float32 slices.
This is a learning and benchmarking workspace. Results depend on hardware,
toolchain, workload and background load; SIMD is not always faster.

Start with the [user Wiki](https://github.com/akolybelnikov/go-simd/wiki):
progressive experiments, the complete command registry, kernel examples, report
interpretation, and troubleshooting. Its editable sources live in
[`docs/wiki`](docs/wiki/Home.md). Run `go run ./examples/basic` for a small
module-local example of all six kernels.

The [architecture migration record](docs/migration/README.md) documents completed
milestones, compatibility contracts, and baseline evidence. See the
[changelog](CHANGELOG.md) for changes and validation notes.

## Requirements

- Go 1.27.3 or later with the `simd` experiment available.
- Set `GOEXPERIMENT=simd` for builds, tests and benchmarks.
- Optional: GNU Make and a POSIX shell for the convenience targets below.
  The Go commands also work directly in PowerShell without Make or Bash.

## Run and verify

From the repository root in PowerShell:

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdbench
go run ./cmd/simdbench -group addition
go run ./cmd/simdbench -list
go test ./...
go vet ./...
gofmt -l cmd internal
```

The CLI accepts `-group addition|computation|all` (default: all). Existing
`-addition`, `-computation`, and `-function add|compute|all` aliases remain
available; conflicting explicit selections are rejected.

Use `-list` (optionally with `-group`) to inspect implemented operations, stable
strategy identities, logical traffic, and this process's runtime SIMD context
without running benchmarks. The compiled quick binary can list from any working
directory without Go on PATH. The report runner's standalone migration is M3.

With Make, use `make run`, `make run-addition`, `make run-computation`,
`make check`, and `make fmt`. Make enables the experiment automatically.
`make clean` removes generated reports and Go build artifacts.

## Benchmark reports

```powershell
go run ./cmd/simdreport -count 5 -benchtime 100ms
go run ./cmd/simdreport -group addition -count 5 -benchtime 500ms
go run ./cmd/simdreport -group addition -format md
go run ./cmd/simdreport -group computation -format csv
# A quick filtered smoke measurement (not a performance conclusion):
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1$/^Scalar$' -count 1 -benchtime 1x
```

Reports save median comparison tables and raw output to `artifacts/<group>/`.
They retain environment metadata, individual measurements, allocations, and
throughput. Use repeated runs on an idle machine before drawing conclusions.
The full sweep reaches 16 million elements and can take several minutes.

Choose `-format txt|md|csv` for the saved report; `txt` is the default. Terminal
comparisons stay readable text for every format, and every report retains its
companion `.raw` measurements. Unknown formats fail before measuring or creating
artifacts. Markdown includes environment metadata below the table. CSV contains
only a header and data rows: timings and numeric percentage-point changes
(`-20.0` means 20% less time than scalar), with empty fields for unavailable
values. CSV metadata stays in the raw companion file.

Equivalent Make targets: `make report`, `make report-addition`, and
`make report-computation`; configure `COUNT=5 BENCHTIME=500ms` and `ARGS`.
For example: `make report-addition ARGS="-format md"`.

## Structure

- `cmd/simdbench`: concise fixed-size CLI comparisons.
- `cmd/simdreport`: benchmark execution, raw output capture and median reports.
- `internal/result`: typed run/case observations, legacy text import, medians,
  and scalar comparisons.
- `internal/report`: TXT, Markdown, and CSV renderers over shared comparisons.
- `internal/input`: deterministic generated inputs and input-bit fingerprints.
- `internal/catalog`: shared operation/strategy metadata and logical byte models.
- `internal/catalog/kernels`: typed bindings to the existing six kernels.
- [`internal/addition`](internal/addition/README.md): four addition variants.
- [`internal/computation`](internal/computation/README.md): scalar and SIMD
  implementations of `a*a + b*b + a*b`.
- [`internal/benchutil`](internal/benchutil/README.md): shared benchmark setup,
  interpretation and historical example results.
- `.github/workflows`: formatting, tests, vet and filtered benchmark checks.

Operations accept disjoint output or exact in-place output aliasing either input.
Partial overlap is unsupported. Length errors leave output unchanged.

The report command prints the comparison table in the terminal and saves it in
the selected format: one row per sample size, scalar timing,
each SIMD timing, and its percentage change from scalar. Negative percentages
mean less time than scalar; positive percentages mean more time. Environment
metadata follows TXT/Markdown tables, and `.raw` files retain individual
measurements. Filtered variants work without a scalar baseline; their changes
are unavailable rather than invented. The runner still requires Go and the
repository root during this migration stage.
