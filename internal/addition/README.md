# `addition`: scalar and SIMD addition experiments

This package provides scalar, partial-vector SIMD, scalar-tail SIMD, and
parallel SIMD addition for `float32` slices.

Run its concise CLI comparison from the repository root with `make run-addition`.
Generate the detailed multi-size table under
`artifacts/addition/` with `make report-addition`.

Inputs and output must have equal lengths. A length error leaves output unchanged.
Output may exactly alias either input; partial overlap is unsupported.


## Float64 addition extension

Add64Scalar/Add64SIMD accept true (a,b,out []float64) error and share
ErrLengthMismatch. Length errors precede writes; empty/disjoint/exact aliases
are supported and partial output overlap is excluded. Arithmetic is ordinary
binary64 addition; NaN payloads unspecified. SIMD loads/stores are bounded to
actual Float64 lanes, including the selected SDK's emulated partial-load case.
Validation, width setup, bounding, loads/addition/stores are timed; logical work
is 24*n. Full little-endian binary64 output digest follows timing. Use
`go run ./examples/addition64`; see [the contract](../../docs/features/F004-addition64.md).
