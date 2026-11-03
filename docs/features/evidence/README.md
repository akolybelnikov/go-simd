# F004 smoke evidence

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
