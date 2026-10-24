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

Enable the experiment when running the project:

```sh
GOEXPERIMENT=simd go run .
```

In VS Code, the project configuration sets `GOEXPERIMENT=simd` for the
integrated terminal and the **Run Go** launch configuration. Open a new
integrated terminal after opening the project for the terminal setting to take
effect.

## Packages

- [`packages/ops`](packages/ops/README.md) — scalar versus SIMD operation
  benchmarks, representative results, and performance interpretation.
