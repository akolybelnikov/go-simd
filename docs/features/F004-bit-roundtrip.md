# F004 extension — Exact Float32 bit reinterpretation round trip

Implemented 2026-11-04 after uint8 saturation. bitroundtrip / bit-roundtrip-f32 /
BenchmarkBitRoundTrip add Scalar/SIMDPartial, for **19 operations / 49 implementations**.
The question is how a representation path behaves as a complete load/store kernel,
without pretending that reinterpretation is numeric conversion or that source
API calls prove machine instructions were emitted.

patterns.BitRoundTripScalar/BitRoundTripSIMD accept `(x,out []float32) error`.
Every opaque 32-bit input representation is preserved exactly: signed zeros,
subnormals, infinities and quiet/signaling NaN payloads. Scalar uses
Float32bits/Float32frombits; SIMD uses Float32s.ToBits/Uint32s.BitsToFloat32.
There is no arithmetic rounding or numeric float-to-integer conversion.
Compilers may eliminate representational steps; the experiment measures the
whole kernel and makes no isolated bitcast-latency or instruction-selection claim.

Length mismatch returns patterns.ErrLengthMismatch before writes. Empty/nil,
disjoint output and exact alias are supported; partial overlap is unsupported.
SIMD queries actual Float32 width and bounds each partial load/store to valid
lanes. Length checks, width setup, loads, representation path and stores are
timed. Fixtures and complete output digest are outside timing; logical work
is 8*n bytes. Benchmarks use disjoint input/output storage.

The fixed generated-bit-roundtrip-float32-v1 recipe reuses the established
raw-bit fixture values but has a distinct generator/operation namespace from
copy. Other fixture/scalar settings are irrelevant. Input and complete output
SHA-256 cover little-endian float32 bits and must agree. Dedicated
standalone-bit-roundtrip/v1 and legacy-report-bit-roundtrip/v1 harnesses retain
Case.Vector actual lanes; copy-specific policy fields remain absent. Native
import requires exact positive integral lanes and complete digest words, while
historical unknown input stays unknown. Shared-package native selectors are
scoped to the requested operation's registered leaves.

Tests compare every output bit with saved input bits, including special values
and random representations, actual-width tails/nonmultiples, aliases/canaries,
input preservation, errors and current zero allocations. Typed unary/fixture
tests verify arity and irrelevant settings. Compiled/native/emulated reports,
agent catalogue/plans, isolated workers and no-Go standalone execution retain
the same full output evidence. CI includes native smoke, emulated parity,
full tests/vet and Linux races. Original frozen-kernel tests explicitly retain
their two original operation IDs, with new descriptors/bindings tested separately.

Retained [normal](evidence/bitroundtrip-vector.measurements.json) and
[emulated](evidence/bitroundtrip-sdk-emulated.measurements.json) captures have
four observations each, n=64, count 2, 3x. Input/output digests agree across both
strategies and modes. See [provenance](evidence/README.md); these tiny captures
verify plumbing, not performance. Uint32s is now an exercised intermediate
representation, rather than a new CLI uint32 input family.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/bitroundtrip
go run ./cmd/simdreport -group bitroundtrip -bench '^BenchmarkBitRoundTrip$/^n=64$' -count 3 -benchtime 100ms -format md
```

See [the user guide](https://github.com/akolybelnikov/go-simd/wiki/27-Bit-reinterpretation-round-trip).
Numeric conversions and further representation studies remain future operations.
