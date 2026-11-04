# Typed integer experiments

The initial operation is uint8 saturating addition, min(a+b,255), with true
byte arrays. SaturatingAdd8Scalar/SaturatingAdd8SIMD return ErrLengthMismatch
before writes; empty/disjoint/exact aliases are supported, partial overlap not.
The scalar sum widens before clipping; SIMD uses bounded Uint8 partial loads,
AddSaturated and stores at actual width. Validation/setup/loads/arithmetic/stores
are timed, with 3*n logical bytes. Complete output hashing follows timing.

Run `go run ./examples/saturatingadd8`; see
[the contract](../../docs/features/F004-saturating-add8.md). Wrapping and other
integer operations remain separate experiments. No float32 input or predicate
encoding is fabricated.
