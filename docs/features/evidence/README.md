# F004 smoke evidence

## Opaque Float32 XOR — 2026-11-04

Normal/emulated xorbits captures retain four observations each, n=64,
count 2, 3x, fixed raw-word/mask inputs. Every output digest agrees across modes.
Windows/amd64, Go1.27.3-X:simd, GOMAXPROCS16; normal 256 bits/8 lanes,
emulated 128 bits/4 lanes. Base a322627 plus uncommitted XOR changes; missing
embedded VCS facts remain unknown. Original samples/run IDs/timestamps retained;
LF/trailing whitespace normalized. No performance claim.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group xorbits -bench '^BenchmarkXorBits$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/xor-bits-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group xorbits -bench '^BenchmarkXorBits$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/xor-bits-checkpoint-emulated
Remove-Item Env:GODEBUG
```

See [the XOR contract](../F004-xor-bits.md): raw word logic, no floating arithmetic.


## Exact bit reinterpretation — 2026-11-04

Normal/emulated bitroundtrip captures retain four observations each, n=64,
count 2, 3x, fixed raw-bit recipe. Every input/output digest agrees across modes.
Windows/amd64, Go1.27.3-X:simd, GOMAXPROCS16; normal 256 bits/8 float32 lanes,
emulated 128 bits/4 float32 lanes. Base 10e496a plus uncommitted roundtrip changes;
missing embedded VCS facts remain unknown. Original numeric data/run IDs/timestamps
are preserved, with LF/trailing whitespace normalized. No timing claims.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group bitroundtrip -bench '^BenchmarkBitRoundTrip$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/bit-roundtrip-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group bitroundtrip -bench '^BenchmarkBitRoundTrip$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/bit-roundtrip-checkpoint-emulated
Remove-Item Env:GODEBUG
```

See [the representation contract](../F004-bit-roundtrip.md). Bitcasts may be
optimized away; these are whole-kernel plumbing observations, not instruction latency.


## Uint8 saturating addition — 2026-11-04

Normal/emulated saturatingadd8 captures retain four observations each, n=64,
count2, 3x, fixed boundary-rich byte recipe. Every full output digest agrees.
Windows/amd64, Go1.27.3-X:simd, GOMAXPROCS16; normal 256 bits/32 byte lanes,
emulated 128 bits/16 byte lanes. Base 18c9f68 plus uncommitted saturation changes;
missing embedded VCS facts remain unknown. Original samples/run IDs/timestamps
are preserved, with LF/trailing whitespace normalized. No timing ranking claimed.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group saturatingadd8 -bench '^BenchmarkSaturatingAdd8$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/uint8-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group saturatingadd8 -bench '^BenchmarkSaturatingAdd8$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/uint8-checkpoint-emulated
Remove-Item Env:GODEBUG
```

See [the saturation contract](../F004-saturating-add8.md). Numeric byte output is
identified separately from predicate masks; actual Uint8 lanes are retained.


## Float64 addition extension — 2026-11-04

The addition64 normal/emulated captures each retain four observations, n=64,
count2, 3x, fixed generated-addition-float64-v1 inputs. Full output digests agree.
Windows/amd64, Go1.27.3-X:simd, GOMAXPROCS16; normal 256 bits with four Float64
lanes, emulated 128 bits with two. Base 8a59a63 plus uncommitted Float64 changes;
missing embedded VCS facts remain unknown. Original samples/run IDs/timestamps
are retained; LF/trailing whitespace normalized. These are plumbing checks.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group addition64 -bench '^BenchmarkAddition64$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/float64-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group addition64 -bench '^BenchmarkAddition64$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/float64-checkpoint-emulated
Remove-Item Env:GODEBUG
```

See [the binary64 contract](../F004-addition64.md) and its bounded SDK partial-load
workaround. Machine float32_lanes and per-case Float64 lanes are distinct facts.


## Three-array MulAdd extension — 2026-11-04

Normal/emulated muladd captures each retain four observations, n=64, rounding
recipe, count2, 3x. All complete output digests agree. Windows/amd64, Go1.27.3
with GOEXPERIMENT=simd, GOMAXPROCS16; normal 256-bit/8-lane vector gate, emulated
128-bit/4-lane software gate. Base was 0534a7d plus uncommitted MulAdd changes;
missing embedded VCS facts remain unknown. Numeric data/run IDs/timestamps are
unchanged; only LF/trailing whitespace is normalized. These are plumbing checks.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group muladd -fma-input rounding -bench '^BenchmarkMulAdd$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/muladd-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group muladd -fma-input rounding -bench '^BenchmarkMulAdd$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/muladd-checkpoint-emulated
Remove-Item Env:GODEBUG
```

See [the MulAdd contract](../F004-muladd.md).


## Final clamp checkpoint — 2026-11-04

Normal-SDK clamp-vector.measurements.json and SDK-emulated
clamp-sdk-emulated.measurements.json each retain six observations: three controls,
two repetitions, exactly three iterations, n=64, shuffled seed 123, bounds -1.5/2.5.
Both output digests agree across controls and execution modes. Captured on
Windows/amd64, Go 1.27.3-X:simd, GOMAXPROCS 16; normal SDK has 256 bits/8 lanes,
emulated has 128 bits/4 lanes. Base was signed commit 72bb7ff plus uncommitted
clamp changes. Embedded source revision/modified facts remain explicitly unknown.
Files retain original data/run IDs/timestamps; only LF/trailing whitespace is normalized.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group clamp '-lower=-1.5' -upper 2.5 -pattern shuffled -seed 123 -bench '^BenchmarkClamp$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/clamp-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group clamp '-lower=-1.5' -upper 2.5 -pattern shuffled -seed 123 -bench '^BenchmarkClamp$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/clamp-checkpoint-emulated
Remove-Item Env:GODEBUG
```

These tiny fixed-iteration captures verify evidence plumbing, not performance.
See the [clamp contract](../F004-clamp.md) and [F004 closure](../F004-completion.md).
The [SDK declaration snapshot](f004-sdk-inventory.json) records the independent
API coverage ledger; it is inventory metadata, not benchmark results.


## Copy checkpoint — 2026-11-04

`copy-vector.measurements.json` and `copy-sdk-emulated.measurements.json` were
captured at 2026-11-03 22:04 UTC (2026-11-04 in Europe/Athens) on Windows/amd64,
Go 1.27.3, GOEXPERIMENT=simd and GOMAXPROCS 16. Normal SDK execution used 256-bit /
8-lane vectors; SDK emulation used 128-bit / 4 lanes. Source base was signed
5bdef6d plus uncommitted copy changes; missing embedded VCS facts remain unknown.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group copy -bench '^BenchmarkCopy$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/copy-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group copy -bench '^BenchmarkCopy$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/copy-emulated-checkpoint
Remove-Item Env:GODEBUG
```

Restore any prior GODEBUG value. SDK emulation does not disable independent
compiler/runtime Go-copy optimization. Each capture retains eight observations:
four controls, two repetitions and exactly three iterations. Disjoint arrays
contain 64 elements; logical work is 512 bytes per call. ArrayResultSHA256 is
`ba1b8dbfef6511e4d01e339bf15d3c05542c1cda7945de30b0386e287efb3273`
for every output and equals the complete input hash. Exact-bit policy, actual
SDK lanes, the fixed raw-bit generator and dedicated harness are retained.
Hashing is outside timing; validation/copying/tails are inside. Different SDK
contexts remain separate. Original numeric data, run IDs and UTC timestamps are
preserved; only filenames, LF endings and trailing whitespace differ. These
tiny captures verify plumbing without performance claims.
See [the copy contract](../F004-copy.md).

## Packed-predicate checkpoint — 2026-11-03

`packed-predicate-vector.measurements.json` and
`packed-predicate-software.measurements.json` were generated on Windows/amd64,
Go 1.27.3, GOEXPERIMENT=simd and GOMAXPROCS 16. Hardware used 256-bit / 8-lane
SIMD; forced emulation used 128-bit / 4-lane SIMD. Source base was signed
f91aa33 plus uncommitted packed-predicate changes. Missing embedded VCS facts
remain explicitly unknown.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group predicatepack '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicatePacked$/^n=65$' -count 2 -benchtime 3x -format md -output artifacts/packed-predicate-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group predicatepack '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicatePacked$/^n=65$' -count 2 -benchtime 3x -format md -output artifacts/packed-predicate-emulated-checkpoint
Remove-Item Env:GODEBUG
```

Restore any prior GODEBUG value. Each capture has four observations, two
strategies/repetitions and exactly three iterations. At n=65 physical output is
9 bytes, with unused high bits of the last byte zero, and logical work 269 bytes.
All ByteResultSHA256 values are
`03e14f2501e827d24080a2cab6add687e892fe4d6d53e778685b2ef12d08925f`.
The digest covers the complete packed output after timing; conversion, packing
and stores are timed. Encoding is uint8-packed-lsb0/v1; each capture retains its
actual process width and exact shared threshold/seed/input identity. Different
hardware/emulated contexts remain separate. Numeric values, run IDs and UTC
timestamps are preserved; only filenames, LF endings and trailing whitespace
differ. These tiny captures verify plumbing without performance claims.
See [the packed contract](../F004-packed-predicate.md).

## Predicate byte-export checkpoint — 2026-11-03

`predicate-vector.measurements.json` and `predicate-software.measurements.json`
were generated on Windows/amd64, Go 1.27.3, GOEXPERIMENT=simd, GOMAXPROCS 16.
Hardware execution used 256-bit / 8-lane SIMD; forced emulation used 128-bit /
4-lane SIMD. Source base was signed b8ecb2e plus uncommitted predicate changes;
missing embedded VCS facts remain explicitly unknown.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group predicate '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicate$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/predicate-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group predicate '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicate$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/predicate-emulated-checkpoint
Remove-Item Env:GODEBUG
```

Restore any prior GODEBUG value. Quote the full negative decimal argument in
PowerShell. Each capture retains four observations: two strategies/repetitions,
exactly three iterations and 320 logical bytes per call. Case.Predicate retains
canonical uint8-zero-one/v1 and actual lanes; input identity retains threshold
bits, shuffled seed and x's hash. Every ByteResultSHA256 is
`0140afddd8f37a2228872352701425d6ade797d97a44915645852e053cd07b9e`,
SHA-256 of the complete 64 canonical bytes, computed after timing. The SIMD
measurement includes normalization, conversion/extraction and byte stores.
Different hardware/emulated contexts remain separate. Original numeric values,
run IDs and UTC timestamps are retained; only filenames, LF endings and trailing
whitespace differ. These tiny captures verify plumbing without performance
claims. See [the contract](../F004-predicate-export.md).

## Fused-dot checkpoint — 2026-11-03

`dotfma-vector.measurements.json` and `dotfma-software.measurements.json` were
generated on Windows/amd64, Go 1.27.3, GOEXPERIMENT=simd and GOMAXPROCS 16.
Hardware execution used 256-bit / 8-lane SIMD and passed the vector FMA gate.
Forced emulation used 128-bit / 4-lane software stripes. Source base was signed
83e3d9f plus uncommitted fused-dot changes; absent embedded VCS facts remain unknown.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group dotfma -dot-input fusion-sensitive -bench '^BenchmarkDotFMA$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/dotfma-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group dotfma -dot-input fusion-sensitive -bench '^BenchmarkDotFMA$/^n=64$' -count 2 -benchtime 3x -format md -output artifacts/dotfma-emulated-checkpoint
Remove-Item Env:GODEBUG
```

Restore any prior GODEBUG value. Each capture retains eight observations:
four strategies, two repetitions and exactly three iterations. ScalarResultBits
is 0xa8800000 (-2^-46) for each of these observations. For other widths/orders,
such as four accumulators at width 16, this fixture can return +0. No universal
cross-strategy equality is claimed. Both FMA and Reduction identities retain
process/actual paths, width, count and fold; logical work is 516 bytes per call.
The software SIMD labels retain software-striped-fma32/v1 rather than scalar
order. Different process contexts remain separate. Original numeric values,
run IDs and timestamps are preserved; only filenames, LF endings and trailing
whitespace differ. These are plumbing captures without performance conclusions.
See [the fused-dot contract](../F004-fused-dot.md).

## Fused-affine checkpoint — 2026-11-03

`fma-vector.measurements.json` and `fma-software.measurements.json` were generated
on Windows/amd64, Go 1.27.3, GOEXPERIMENT=simd and GOMAXPROCS 16. The first used
256-bit / 8-lane SIMD and passed the native FMA gate. The second used
GODEBUG=simd=0: 128-bit / 4-lane emulation, explicitly selecting corrected software.
Source base was signed 2689f50 plus uncommitted FMA changes; missing embedded
VCS facts remain unknown.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group affinefma -fma-input rounding '-alpha=1.0000001192092896' '-beta=-1' -bench '^BenchmarkAffineFMA$/^n=8$' -count 2 -benchtime 3x -format md -output artifacts/fma-checkpoint
$env:GODEBUG = 'simd=0'
go run ./cmd/simdreport -group affinefma -fma-input rounding '-alpha=1.0000001192092896' '-beta=-1' -bench '^BenchmarkAffineFMA$/^n=8$' -count 2 -benchtime 3x -format md -output artifacts/fma-emulated-checkpoint
Remove-Item Env:GODEBUG
```

Restore a prior GODEBUG value if applicable. Each capture retains four observations,
two strategies/repetitions, exactly three iterations, exact coefficient/input
identity, process/strategy FMA paths and actual lanes. The final output bits are
0xa8800000 (-2^-46) for both strategies in both captures; unfused affine would
yield +0 in that cell. ArrayResultLastBits is only the final cell, not a checksum.
Different vector/software process contexts remain separate. Original numeric
data/run IDs/UTC timestamps are retained; only filenames, LF endings and trailing
whitespace differ. These tiny runs are plumbing evidence without performance
claims. See [the FMA contract](../F004-fma.md).

## Accumulator checkpoint — 2026-11-03

`accumulators-sum.measurements.json` and `accumulators-dot.measurements.json`
were generated on Windows/amd64, Go 1.27.3, GOEXPERIMENT=simd, GOMAXPROCS 16,
256-bit vectors / 8 float32 lanes in hardware mode. Source base was signed
ad5c928 plus uncommitted accumulator changes. Missing embedded VCS facts remain
explicitly unknown.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group sum -sum-input cancellation -bench='^BenchmarkSum$/^n=64$' -count=2 -benchtime=3x -format=md -output=artifacts/accumulator-checkpoint
go run ./cmd/simdreport -group dot -dot-input cancellation -bench='^BenchmarkDot$/^n=64$' -count=2 -benchtime=3x -format=md -output=artifacts/accumulator-checkpoint
```

Each capture retains eight observations: Scalar and SIMDStriped/2/4, two
repetitions, exactly three reported iterations. VectorAccumulators is 0/1/2/4
with distinct OrderIDs and unchanged harness/input/byte identities. For this
fixture/width/size scalar returns +0; all SIMD variants return 1. The runnable
example separately exposes count-sensitive numerical differences. Original
numerical data, timestamps and run IDs are preserved; only filename, LF endings
and trailing whitespace differ. These are plumbing observations without timing
or accuracy rankings. See the [accumulator contract](../F004-accumulators.md).

## Unfused dot checkpoint — 2026-11-03

`dot-cancellation.measurements.json` was generated on Windows/amd64, Go 1.27.3,
GOEXPERIMENT=simd, GOMAXPROCS 16, 256-bit vectors / 8 float32 lanes in hardware
mode. Source base was signed commit 6c1e94b plus uncommitted dot changes.
The capture's absent embedded VCS facts remain explicitly unknown.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group dot -dot-input cancellation -bench='^BenchmarkDot$/^n=64$' -count=2 -benchtime=3x -format=md -output=artifacts/dot-checkpoint
```

Four observations retain exactly three iterations each, the dot harness,
both-input fingerprints, unfused fold identity and result bits. Scalar returns
+0 (bits 0); SIMDStriped returns 1 (bits 1065353216) for this width/fixture/size.
Logical work is 516 bytes per call. Original numerical data, run IDs and UTC
timestamps are retained; only filename, LF endings and trailing whitespace differ.
This is a plumbing observation without performance or accuracy ranking.
See the [dot contract](../F004-dot.md) for numerical acceptance and boundaries.

## Ordered sum checkpoint — 2026-11-03

`sum-cancellation.measurements.json` was generated on Windows/amd64, Go 1.27.3,
GOEXPERIMENT=simd, GOMAXPROCS 16, 256-bit vectors / 8 float32 lanes in hardware
mode. Source base was signed commit 71ceb59 plus the uncommitted sum checkpoint.
The go-run capture's missing embedded VCS facts remain explicitly unknown.

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group sum -sum-input cancellation -bench='^BenchmarkSum$/^n=64$' -count=2 -benchtime=3x -format=md -output=artifacts/sum-checkpoint
```

Four observations retain exactly three iterations each, the reduction harness,
input identity, ordered fold metadata and result bits. Scalar returns +0 (bits 0),
SIMDStriped returns 1 (bits 1065353216), under the documented eight-lane fold.
Logical work is 260 bytes per call. The original numeric data, run IDs and UTC
timestamps are retained; only filename, LF endings and trailing whitespace differ.
This is a plumbing observation, with no timing or accuracy ranking claimed.
See the [sum contract](../F004-sum.md) for numerical tests and limitations.

## Broadcast/predicate first slice — 2026-10-27

The selection capture was generated on 2026-10-27 during first-slice development
on Windows/amd64, Go 1.27.3, GOEXPERIMENT=simd, GOMAXPROCS 16, 256-bit vectors,
8 float32 lanes and hardware mode. Source base was signed commit d9f01f0 plus
uncommitted F004 changes. Missing VCS facts inside the go-run capture remain
explicitly unknown; no source identity was fabricated.

Command from the checkout in PowerShell:

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group selection -pattern shuffled -seed 123 -threshold=-1 -bench='^BenchmarkSelection$/^n=(8|16)$' -count=2 -benchtime=2x -format=md -output=artifacts/f004-smoke
```

The retained file contains eight observations: scalar/SIMDPartial, sizes 8/16,
two repetitions, exactly two reported iterations. It preserves the pattern
harness, threshold bits, seeded generator/fingerprints and runtime context.
Its filename is shortened; embedded run IDs/timestamps and numeric data are
unchanged, with LF line endings and trailing whitespace removed.

This is a plumbing check. At 2x and tiny sizes, timer resolution and scheduling
dominate; the values are not performance or regression claims. Complete numerical,
distribution, backend parity and no-Go standalone verification live in tests;
hardware and emulated execution are checked separately.


## Fixed-block sum8

[Normal](sum8-vector.measurements.json) and [emulated](sum8-sdk-emulated.measurements.json) retain n=64, cancellation, count 2, 3x at source base 78f6f13 plus uncommitted sum8 changes, Windows/amd64 Go1.27.3-X:simd. Original samples and timestamps retained. Every full output digest agrees; physical extent is 8. These verify plumbing, not performance. Reproduce with `go run ./cmd/simdreport -group sum8 -sum-input cancellation -bench '^BenchmarkSum8$/^n=64$' -count 2 -benchtime 3x -format md`, then repeat with `GODEBUG=simd=0`.
