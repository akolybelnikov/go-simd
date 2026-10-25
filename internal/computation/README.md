# `computation`: scalar and SIMD compute experiments

This package provides scalar and SIMD implementations of the element-wise
expression `a*a + b*b + a*b` for `float32` slices.

Run its concise CLI comparison from the repository root with
`make run-computation`. Generate the detailed multi-size table under
`artifacts/computation/` with `make report-computation`.

Inputs and output must have equal lengths. A length error leaves output unchanged.
Output may exactly alias either input; partial overlap is unsupported.
