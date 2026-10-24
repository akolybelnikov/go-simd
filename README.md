# go-simd

An experimental Go project for exploring the standard library's SIMD support.
The goal is to compare scalar and SIMD approaches to the same problems, measure
potential performance improvements, and understand how vectorization can affect
algorithm design and implementation.

The project is a learning and benchmarking space, not a production library.
Performance results depend on the hardware, Go version, and workload, so
comparisons should use equivalent inputs and proper Go benchmarks rather than
assuming SIMD is always faster.

## Requirements

- Go 1.27.3 or later
- A Go toolchain with the `simd` experiment

The SIMD package is experimental and may change between Go releases.

## Run

Open a terminal in the repository root (the directory containing `go.mod`).
The CLI needs Go 1.27.3 or later and the `simd` experiment enabled. Run the
concise comparison for both benchmark groups with:

```sh
make run
```

The Makefile enables the SIMD experiment automatically. To run one group, pass
the CLI flags through `ARGS`:

```sh
make run ARGS="-function add"
make run ARGS="-function compute"
make run ARGS="-h"
```

The default is `-function all`. Each group runs on the same fixed input size
(1,048,576 `float32` elements) and prints a comparison table. `TIME (ns/op)`
is the time for one operation call; `CHANGE VS SCALAR` compares that time with
the scalar implementation. A negative percentage means that implementation
was faster in that run, while a positive percentage means it was slower.
Benchmark results vary with the machine and system load, so treat a single run
as an experiment, not a guarantee.

For the detailed multi-size tables (including sizes around SIMD vector
boundaries), run:

```sh
make report
```

This saves only the clean comparison tables (without per-benchmark terminal
output) in [`packages/ops/artifacts/`](packages/ops/artifacts/). The terminal
shows the concise fixed-size comparison first, followed by the detailed report
path. Reports are generated artifacts and are ignored by Git. Remove the report
folder with:

```sh
make clean
```

`make report` uses a 100 ms measurement per benchmark by default; override it
with `make report BENCHTIME=500ms` for longer measurements. The CLI input size
and report sweep settings are not yet exposed as CLI flags. See
[`packages/ops`](packages/ops/README.md) for details and interpretation.

In VS Code, the project configuration sets `GOEXPERIMENT=simd` for the
integrated terminal and the **Run Go** launch configuration. Open a new
integrated terminal after opening the project for the terminal setting to take
effect. If you use PowerShell instead of a Linux or macOS shell, enable the
experiment for the command like this:

```powershell
$env:GOEXPERIMENT = "simd"
go run ./cmd/simdbench
```

## Packages

- [`packages/ops`](packages/ops/README.md) — scalar versus SIMD operation
  benchmarks, representative results, and performance interpretation.
