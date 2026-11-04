# F004 extension — True Float64 addition

Implemented 2026-11-04 after three-array MulAdd. Group addition64, operation
add-f64 and BenchmarkAddition64 add Scalar/SIMDPartial controls. The current
catalogue contains **17 operations / 45 implementations**. Existing addition
continues to use float32 with its original four strategies.

The learning question is how binary64 precision, half the lane count at a fixed
vector width, and twice the logical bytes per element affect addition. Results
at equal n do not imply equal working-set bytes. Different element types remain
different comparison contexts; no implicit casts or cross-type speedup pooling.

## Contract, fixtures and timing

addition.Add64Scalar/Add64SIMD accept `(a,b,out []float64) error`. They compute
ordinary binary64 addition, including subnormals, signed zeros, infinities and
overflow. NaN payloads are unspecified. Unequal lengths return the existing
addition.ErrLengthMismatch before writes; valid empty/nil succeeds. Disjoint
output and exact aliases are supported; partial output overlap is unsupported.

The fixed generated-addition-float64-v1 recipe computes values directly as
float64 decimal fractions; it does not widen already-rounded float32 fixtures.
Other fixture/scalar flags do not alter it. Both inputs contribute complete
little-endian binary64 bits to input identity. Full output SHA-256 is retained
after timing in ArrayResultSHA256; Case.TypeID defines that representation.
The typed BindFloat64 seam, native callback, quick CLI and compiled adapter use
real []float64 arrays. No float32 output or fabricated cast crosses these seams.

Timing includes length checks, actual-width setup, bounded partial loads/stores,
addition and loop control. Fixtures and full-output hashing are outside timing.
Logical work is 24*n bytes (two reads and one write), not DRAM counters. The
standalone-addition64/v1 and legacy-report-addition64/v1 harnesses retain actual
float64 lane count in Case.Vector. The machine's float32_lanes field keeps its
existing meaning as a process fact; it is not the float64 operation width.
Plans/sweeps derive the corresponding Float64 boundaries, and native import
requires actual positive integral lanes plus all eight exact digest words.

## Selected SDK partial-load boundary

The inspected Go 1.27.3 simd_emulated.go Float64 partial loader can return the
supplied slice length while loading only its two emulated lanes. Passing an
entire remaining slice would skip values. Add64SIMD bounds every load/store to
min(actual lanes, remaining elements) and advances by the bounded load count.
The bounding/setup cost is timed. Native and separately initialized
GODEBUG=simd=0 tests verify the same semantics; no unsupported feature override
or silent claim of accelerated execution is required.

## Acceptance and evidence

An independent exact-rational oracle checks binary64 rounding, random raw bits,
signed zeros, subnormals, infinities, overflow and NaNs. Kernel tests cover actual
lane boundaries, nonmultiples, canaries, input preservation, aliases, errors and
current zero allocations. Binding/fingerprint tests reject implicit float32 use
and prove low binary64 input bits survive. Adapter tests retain the full output
and restore testing flags after success/error. Report tests compare both
backends' complete outputs and context, including emulation. Agent catalogue,
plan and isolated-worker tests verify true types, bytes and actual-width boundaries.
All-group standalone tests execute without Go/source. CI includes native smoke,
emulated kernels/parity and full Linux/Windows validation with Linux races.

Retained [normal-SDK](evidence/addition64-vector.measurements.json) and
[emulated](evidence/addition64-sdk-emulated.measurements.json) captures each
contain four observations: two strategies, two repetitions, n=64, 3x. Their
complete output digests agree. See [provenance](evidence/README.md); these tiny
captures verify plumbing and establish no performance ranking. The SDK coverage
ledger now marks Float64s, its Add/StorePart/Len and LoadFloat64sPart as exercised.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/addition64
go run ./cmd/simdreport -group addition64 -bench '^BenchmarkAddition64$/^n=64$' -count 3 -benchtime 100ms -format md
```

See [the user guide](https://github.com/akolybelnikov/go-simd/wiki/25-Float64-addition).
Other Float64 operations, integer families and imported datasets retain their
own future scope; this extension adds only generated binary64 addition.
