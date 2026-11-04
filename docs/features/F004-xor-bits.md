# F004 extension — XOR of opaque Float32 bit patterns

Implemented 2026-11-04 after exact bit roundtrip. xorbits / xor-bits-f32 /
BenchmarkXorBits add Scalar/SIMDPartial, for **20 operations / 51 implementations**.
The experiment applies uint32 logic to two opaque float32 representations. It
can flip sign/exponent/payload bits; it performs no floating-point arithmetic
or numeric conversion and introduces no uint32 CLI input family.

patterns.XorBitsScalar/XorBitsSIMD accept `(a,b,out []float32) error`.
Each output word is exactly bits(a) XOR bits(b). Scalar uses ordinary uint32
logic; SIMD uses ToBits/Xor/BitsToFloat32. Every resulting bit pattern is valid
as an opaque output, including zeros, infinities and quiet/signaling NaNs.
A sign-bit word 80000000 stored as -0 flips 1.0 to -1.0; an all-zero mask retains
every original payload bit. Numeric float addition is not an equivalent control.

Lengths must agree before writes; errors return patterns.ErrLengthMismatch.
Empty/nil, disjoint output and exact aliases are supported; partial output
overlap is unsupported. SIMD bounds partial loads/stores to actual lanes.
Validation, width setup, loads, bitwise path and stores are timed with 12*n
logical bytes. Fixtures and complete little-endian bit hashes are outside timing.
Measured arrays are disjoint. Source labels and timings do not prove instruction
selection; representational steps may be eliminated by the compiler.

The fixed generated-xor-float32-bits-v1 recipe combines raw special-value bits
with sign/exponent/payload/all-bit masks. Both operands contribute to identity;
other fixture/scalar flags are irrelevant. Full ArrayResultSHA256 and actual
Float32/Uint32 lanes survive standalone-xor-bits/v1 and legacy-report-xor-bits/v1
captures. Native import requires complete digest words and integral positive
lanes; historical unknown input remains unknown. Shared-package native groups
are scoped to their own registered leaves even with a broad selector.

Tests use an independent per-bit Boolean truth oracle on random/special words,
actual-width tails/nonmultiples, aliases/canaries, unchanged inputs, errors and
current zero allocations. Binding/fixture tests cover both run/bound paths and
every input bit. Native/compiled/emulated reports, agent plans/catalogue and
isolated workers retain identical complete output evidence. Standalone tests
execute all groups without Go/source. CI includes emulated parity, native smoke,
full Linux/Windows tests/vet and Linux races.

Retained [normal](evidence/xorbits-vector.measurements.json) and
[emulated](evidence/xorbits-sdk-emulated.measurements.json) captures have four
observations each, n=64, count 2, 3x. All output digests agree; see
[provenance](evidence/README.md). These tiny captures verify plumbing without
performance claims. Uint32s.Xor is now exercised in the SDK source-use ledger,
while direct uint32 inputs and further shifts/rotates remain separate candidates.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/xorbits
go run ./cmd/simdreport -group xorbits -bench '^BenchmarkXorBits$/^n=64$' -count 3 -benchtime 100ms -format md
```

See [the user guide](https://github.com/akolybelnikov/go-simd/wiki/28-XOR-of-float32-bit-patterns).
