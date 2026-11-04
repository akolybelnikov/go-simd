# F004 extension — Three-array fused float32 multiply-add

Implemented 2026-11-04 after the completed 15-operation F004 baseline. This
extension adds muladd / muladd-f32 / BenchmarkMulAdd, bringing the catalogue
to **16 operations / 43 implementations**. It asks how scalar/SIMD controls
compare when all three operands stream from arrays, with no broadcast constants
or reduction fold. It makes no memory-bottleneck claim.

## Numerical and storage contract

MulAddScalar and MulAddSIMD in internal/patterns accept
`(a,b,c,out []float32) error`. Each element rounds a*b+c once to binary32,
nearest ties-to-even, including subnormals, overflow, infinities and signed zero.
NaN classification is preserved; payload bits are unspecified. The corrected
software FMA32 reference avoids widening/narrowing double rounding. SIMDPartial
uses the process-verified portable vector gate or the recorded software fallback.
Scalar always uses software FMA32; ratios therefore include that reference cost.
The existing affine/dot operations and their timing callbacks are unchanged.

All four lengths must agree before any write. Failure returns the existing
patterns.ErrLengthMismatch. Valid empty/nil succeeds. Disjoint output and exact
alias with any input are supported, including all inputs sharing storage; partial
output/input overlap is unsupported. Benchmarks use disjoint arrays.

## Fixtures, timing and identity

The existing -fma-input selects balanced, rounding or mixed-magnitude triples.
Alpha/beta and predicate/reduction flags do not affect this operation. Each input
has its own array; the third operand is not a scalar or scratch buffer.
MakeMulAdd prepares all four arrays outside timing. FingerprintTernaryFloat32 and
the streaming FingerprintForOperation agree over a's bits, b's bits, then c's
bits in little-endian order, with common n/arity retained in operation identity.
The two-input fixture helper explicitly rejects ternary descriptors.

The typed BindTernary seam is used by quick/native/compiled execution. Agent
catalogue DTOs expose all three inputs, plans normalize the relevant recipe,
and isolated workers retain identical inputs and outputs. Logical work is
16*n bytes, not measured DRAM traffic. Length checks, dispatch, loads, FMA and
stores are timed. Fixture construction and full output SHA-256 follow their
declared outside-timing boundaries. Every output is retained after timing.

Harnesses are standalone-muladd-fma/v1 and legacy-report-muladd-fma/v1.
Case.FMA retains one-rounding policy, actual implementation path, process vector
gate and child/native lanes. ArrayResultSHA256 covers all little-endian output
bits. Native import requires exact gate/path/width and eight complete digest
words; historical text without fixture facts keeps unknown input unknown.
Results cannot be pooled with affine, dot or an unfused three-array expression.

## Acceptance and evidence

An independent exact-rational oracle tests the kernels on special/rounding-sensitive
triples, random raw bits, actual lane boundaries, nonmultiples, exact aliases,
canaries, unchanged inputs, length errors and current zero allocations. The
reference package also retains its independent rational tests. Binding/fixture
tests prove all three operands contribute to identity; an irrelevant scalar
parameter does not. Compiled/native parity covers every recipe, width boundaries,
counts/iterations and full output, including SDK emulation. Worker and standalone
no-Go tests verify propagation beyond direct kernels. CI includes native smoke,
emulated parity, full Linux/Windows tests/vet and Linux races.

Retained [normal-SDK](evidence/muladd-vector.measurements.json) and
[emulated](evidence/muladd-sdk-emulated.measurements.json) observations preserve
four samples each (two strategies, two repetitions, n=64, 3x, rounding recipe).
See [provenance](evidence/README.md). These are plumbing observations, not speedup
claims. The selected SDK declaration ledger gains no new API members: Float32s.MulAdd
was already exercised by affine/dot, while this extension changes operand shape.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/muladd
go run ./cmd/simdreport -group muladd -fma-input rounding -bench '^BenchmarkMulAdd$/^n=64$' -count 3 -benchtime 100ms -format md
```

See [the user guide](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add).
Broader numeric/ISA studies remain separate operations with their own commits
and tracker transitions. The original [F004 closure](F004-completion.md) remains
a historical acceptance record for the baseline scope.
