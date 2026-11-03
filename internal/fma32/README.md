# Correctly rounded binary32 FMA

FMA(a,b,c float32) computes the exact product plus addend and rounds once to
float32, nearest ties-to-even. It supports gradual underflow, overflow and IEEE
special values; NaN payloads are unspecified.

Finite float32 products are exact in float64. TwoSum retains the exact residual
of product+c, then a round-to-odd adjustment prevents binary64-to-binary32
midpoint double rounding. Special values use math.FMA for classification/signs.
No big-number arithmetic or allocation is used in this primitive. A plain
float32(math.FMA(float64(a),float64(b),float64(c))) is not equivalent universally.

Tests use an independent exact-rational oracle, random bit patterns, special
values and explicit normal/subnormal/overflow double-rounding counterexamples.
This ordinary-source package works without GOEXPERIMENT=simd. The vector gate
is shared through internal/fmacap; array kernels live in internal/patterns.
Their actual path is recorded.

See [the contract](../../docs/features/F004-fma.md) and
[Wiki exercise](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding).
