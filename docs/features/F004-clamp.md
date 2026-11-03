# F004, slice 8 — Ordered float32 clamp

Implemented 2026-11-04. This final slice closes [F004's agreed current scope](F004-completion.md)
at fifteen operations and forty-one implementations. The experiment asks how
scalar branching, portable SIMD selection and repaired Min/Max compare when
they promise exactly the same output bits.

Implementation: signed commit [9951b81](https://github.com/akolybelnikov/go-simd/commit/9951b81b167af193589b1b834a0b84297c18af3f).

## Contract and controls

Operation `clamp-f32`; group `clamp`; benchmark `BenchmarkClamp`; arithmetic
`float32-clamp-ordered-bits-v1`. One `[]float32` input, one caller-owned output,
and exact float32 `lower`/`upper` parameters. All functions return an error.

| Function in internal/patterns | Label / strategy | Timed implementation |
| --- | --- | --- |
| ClampScalar | Scalar / scalar | Ordered scalar branches |
| ClampSIMD | SIMDPartial / simd-partial | Partial load, two ordered comparisons, IfElse selections, partial store |
| ClampSIMDMinMax | SIMDMinMax / simd-minmax | Max/Min plus explicit comparisons/selections and Mask32s.Or repairs |

For each input v: below lower returns the exact lower bits; above upper returns
the exact upper bits; otherwise return the original input bits. Equality includes
both signs of zero. Input quiet/signaling NaNs keep their original payload bits.
For bounds (+0,-0), negative values become +0, positive values become -0,
existing zeros retain their signs, and NaNs retain their bits.

Bounds must be non-NaN and lower<=upper; equal bounds, either zero ordering and
infinities are valid. Length mismatch returns patterns.ErrLengthMismatch first.
Otherwise invalid bounds return patterns.ErrInvalidBounds, even for empty input.
Every failure precedes writes. Equal empty/nil arrays with valid bounds succeed.
Disjoint storage and exact x/output alias are supported; partial overlap is outside
the contract. Timings use disjoint buffers.

Raw portable Min/Max can choose different operands on NaNs and equality. The
Min/Max control explicitly repairs original ties/NaNs and exact outside-bound bits.
Repairs are part of its measured cost; this is not an isolated instruction study.
All three strategies preserve the same bit contract in SDK hardware/emulated modes.

## Fixtures, work and evidence

Reuse shared -lower, -upper, -pattern and -seed controls. Here the pattern marks
**clipped positions**: none/all, even-index alternating, first ceil(n/2) clustered,
or deterministic approximately-50% shuffled. Nonclipped positions include center,
bound ties, valid signed zeros and payload NaNs; NaNs remain unclipped.
Outside values use adjacent representable values on the available sides. Bounds
[-Inf,+Inf] cannot produce clipped values: require -pattern none. Parameter bits,
generator clamp-float32-v1/<pattern> and shuffled seed retain the exact context.
Input generation and streaming fingerprints agree; unrelated options do not alter it.

Logical traffic is 8*n bytes, not measured DRAM traffic. Validation, per-call
broadcasts, clipping, repairs and stores are timed. Fixture allocation and
complete output SHA-256 over little-endian float32 bits are outside timing.
Dedicated standalone-clamp/v1 and legacy-report-clamp/v1 harnesses retain
ArrayResultSHA256 and Case.Clamp with ordered-clamp-preserve-bits/v1 plus actual
native lanes. Native import requires positive integral lanes and all eight valid
digest words. Different policies/widths cannot silently share comparisons.

## Acceptance and retained observations

Independent bit-reference tests exercise random raw bits, NaNs, signed zeros,
subnormals, infinities, finite/equal/one-sided-infinite/infinite bounds, empty
input, actual-width boundaries, offset/canary storage, larger nonmultiples,
exact aliases, input preservation and failure precedence. Current native widths
allocate zero per call. Binding/fixture/identity tests cover all three controls.
Full report tests compare both backends for all five patterns, zero-sign bounds,
the unbounded interval, repetitions/iterations and every output bit via digest.
Standalone tests run every group with neither source nor Go on PATH. CI repeats
the contract/parity under GODEBUG=simd=0 and executes native clamp smoke tests.

Retained [normal-SDK](evidence/clamp-vector.measurements.json) and
[SDK-emulated](evidence/clamp-sdk-emulated.measurements.json) captures preserve
six observations each. [Commands/provenance](evidence/README.md) describe their
source state and exact configuration. Three iterations are plumbing evidence,
not a performance ranking. No timing threshold is asserted.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/clamp
go run ./cmd/simdreport -group clamp '-lower=-1.5' -upper 2.5 -pattern shuffled -seed 123 -bench '^BenchmarkClamp$/^n=64$' -count 3 -benchtime 100ms -format md
```

Continue with the [user guide](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts).
Stop after this slice and the closure audit. Broader types, three-array MulAdd,
reordering/conversion studies, M4 and HTML remain separately scoped future work.
