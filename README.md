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

## Run

Open a terminal in the repository root. The CLI needs Go 1.27.3 or later and
the `simd` experiment enabled. Run the concise comparison for both groups with:

```sh
make run
```

To run a specific benchmark group:

```sh
make run ARGS="-group addition"
make run ARGS="-group computation"
make run-addition
make run-computation
```

The default is all groups. The CLI now supports the split package layout with
`-group` and the dedicated `-addition` / `-computation` flags. Use
`make report` to save the detailed comparison tables under the corresponding
operation folders in `packages/benches/artifacts/`; `make clean` removes them.

## Packages

- [`packages/benches`](packages/benches/README.md) — shared benchmark utilities,
  full results, and performance interpretation.
- [`packages/benches/addition`](packages/benches/addition/README.md) — addition
  implementations and benchmarks.
- [`packages/benches/computation`](packages/benches/computation/README.md) —
  compute-heavy implementations and benchmarks.
