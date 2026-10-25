# `computation`: scalar and SIMD compute experiments

This package provides scalar and SIMD implementations of the element-wise
expression `a*a + b*b + a*b` for `float32` slices.

Run its concise CLI comparison from the repository root with
`make run-computation`. Generate the detailed multi-size table under
`packages/benches/artifacts/computation/` with `make report-computation`.
