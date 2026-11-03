# F004 smoke evidence

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
