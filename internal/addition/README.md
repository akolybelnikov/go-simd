# `addition`: scalar and SIMD addition experiments

This package provides scalar, partial-vector SIMD, scalar-tail SIMD, and
parallel SIMD addition for `float32` slices.

Run its concise CLI comparison from the repository root with `make run-addition`.
Generate the detailed multi-size table under
`artifacts/addition/` with `make report-addition`.

Inputs and output must have equal lengths. A length error leaves output unchanged.
Output may exactly alias either input; partial overlap is unsupported.
