# F004 extension — Uint8 saturating addition

Implemented 2026-11-04 after Float64 addition. Group saturatingadd8, operation
add-sat-u8 and BenchmarkSaturatingAdd8 add Scalar/SIMDPartial controls over real
uint8 arrays. The catalogue now contains **18 operations / 47 implementations**.
This experiment distinguishes a bounded integer sum from modulo-256 wrapping;
it adds a real integer input type rather than reusing a float predicate export.

## Contract and timed work

integer.SaturatingAdd8Scalar/SaturatingAdd8SIMD accept `(a,b,out []uint8) error`.
For each element compute min(uint16(a)+uint16(b),255). Scalar widens before adding;
SIMD uses portable Uint8s.AddSaturated. Thus 128+128 and 255+1 both yield 255,
whereas wrapping would yield 0. Wrapping is a distinct unimplemented operation.

All lengths must agree before writes; otherwise return integer.ErrLengthMismatch.
Valid empty/nil succeeds. Disjoint output and exact aliases are supported;
partial output overlap is unsupported. Measured arrays are disjoint.
SIMD uses actual Uint8 lane count and bounded partial loads/stores. Validation,
width setup, bounding, loads, saturation and stores are timed. Allocation,
fixture generation and complete output hashing are outside timing.
Logical work is 3*n bytes: two byte reads and one write, not DRAM measurements.

The fixed generated-saturating-add-uint8-v1 recipe includes values around255,
both saturated and unsaturated sums, zeros and endpoints. Other scalar/fixture
flags do not alter it. Both byte inputs contribute to a streamed SHA-256;
no float32 arrays or narrowing conversions are fabricated. Typed BindByteArrays
closures propagate through quick/compiled/native execution and isolated workers.

## Evidence identity and acceptance

ByteResultSHA256 retains every numeric output byte after timing. Case.Vector
records actual Uint8 lanes, and TypeID=uint8 plus uint8-add-saturate-255-v1
identifies arithmetic. Case.Predicate remains absent: these bytes are numeric
results, not canonical 0/1 masks or packed predicate bits. Harnesses are
standalone-saturating-add8/v1 and legacy-report-saturating-add8/v1. Native import
requires exact integral lanes and all eight complete digest words. Historical
text without generated-fixture authority retains unknown input identity.

Tests exhaust all 65,536 input pairs under both SDK modes. Additional tests cover
actual-width tails, larger nonmultiples, offset/canary output, aliases, unchanged
inputs, errors before writes and current zero allocations. Binding/fixture tests
verify true byte arrays, all-input identity and irrelevant options. Report tests
match complete native/compiled outputs and contexts, including SDK emulation.
Agent catalogue/plans/worker tests verify true input/output types and 3*n work.
Standalone tests run all groups without Go/source. CI includes normal/emulated
kernels and parity, native smoke, full tests/vet and Linux races.

Retained [normal](evidence/saturatingadd8-vector.measurements.json) and
[emulated](evidence/saturatingadd8-sdk-emulated.measurements.json) captures have
four observations each, n=64, count 2, 3x; all output digests agree. See
[provenance](evidence/README.md). They verify plumbing without timing claims.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/saturatingadd8
go run ./cmd/simdreport -group saturatingadd8 -bench '^BenchmarkSaturatingAdd8$/^n=64$' -count 3 -benchtime 100ms -format md
```

See [the user guide](https://github.com/akolybelnikov/go-simd/wiki/26-Uint8-saturating-addition).
Other integer widths, wrapping operations, conversions and imported datasets
remain separate candidates; this extension does not imply their implementation.
