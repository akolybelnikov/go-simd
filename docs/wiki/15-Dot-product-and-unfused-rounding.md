# 15 — Dot product and unfused rounding

Dot product multiplies matching x/y elements and reduces them to one float32.
F004's dot checkpoint brought the lab to nine operations/twenty implementations.
The later [accumulator checkpoint](https://github.com/akolybelnikov/go-simd/wiki/16-Accumulator-count-experiments)
adds SIMDStriped2/4 for a total of twenty-four implementations. This page specifies
the original Scalar/SIMDStriped controls; FMA remains a future experiment.

## Run a small dot experiment

Installed commands:

```powershell
simdreport -list -group dot
simdreport -group dot -dot-input cancellation -bench '^BenchmarkDot$/^n=64$' -count 3 -benchtime 100ms -format md
simdreport -group dot -dot-input mixed-magnitude -bench '^BenchmarkDot$/^n=1024$' -count 3 -benchtime 100ms -format csv
simdbench -group dot -dot-input balanced
```

For source runs, enable GOEXPERIMENT=simd and replace the executable with
`go run ./cmd/simdreport` or `go run ./cmd/simdbench`. Make forwards the same flags:

```bash
make list ARGS="-group dot"
make report COUNT=3 BENCHTIME=100ms ARGS="-group dot -dot-input cancellation -bench BenchmarkDot/n=64 -format md"
make run ARGS="-group dot -dot-input balanced"
```

There is no separate Make shortcut or installed-command dialect to remember.
The quick command uses 1,048,576 elements; report regex selects registered sizes.
The direct report examples anchor their sizes to avoid substring matches.

## Multiplication rounding comes first

Every product rounds to float32 before accumulation; every addition also rounds.
The scalar kernel starts at +0 and folds products left to right. SIMDStriped
accumulates full-vector products in one native-width vector accumulator, folds
its lanes low to high from +0, then adds rounded scalar-tail products.
Length validation, multiplication, extraction, final fold and tail are timed.
No padded input lanes are multiplied or reduced.

This is an **unfused** dot. For x=[-1,1+2^-23] and y=[1,1-2^-23], scalar returns
+0 because the second product rounds to 1 before adding -1. A fused second
multiply/add would retain -2^-46. An FMA variant would need a different contract
and identity; it is not enabled by this checkpoint.

Different fold orders can lose rounding information in either direction. With
w native lanes, length 2*w, y all ones and other x entries zero:

| Nonzero x entries | Scalar | SIMDStriped | Exact mathematical dot |
| --- | --- | --- | --- |
| x[0]=2^24, x[1]=1, x[w]=-2^24 | 0 | 1 | 1 |
| x[0]=2^24, x[1]=-2^24, x[w]=1 | 1 | 0 | 1 |

Scalar is the timing baseline. Neither fold is always more accurate; timing
percentages do not measure numerical accuracy. Tests use independently rounded
exact-rational products and additions, with a finite absolute forward-error
check where its assumptions apply. No universal relative tolerance near
cancellation is promised.

NaNs, infinities, product overflow/underflow and signed zero follow those orders.
NaN payloads are unspecified; finite inputs may yield different exceptional
results. Empty inputs and all-negative-zero products yield +0 under these folds.

## Choose an input

| `-dot-input` | Recipe |
| --- | --- |
| balanced (default) | x[i]=float32(i%33-16)/4; y[i]=float32(i%17-8)/8 |
| cancellation | Repeat 64 x values: 2^24 at 0, 1 at 1, -2^24 at 32, otherwise zero; all y=1 |
| mixed-magnitude | Repeat x=[1e-20,1,1e20,-1e20,-1,3,-3,1e-20] and y=[1e20,1e-20,1e-20,1e-20,1e20,1e-20,1e-20,1e20] as float32 |

The cancellation recipe produces Scalar=0 and SIMDStriped=1 at n=64 on the
currently tested 4/8/16-lane widths. Retain actual width and size when comparing.
`-sum-input`, predicate settings and scalar coefficients do not alter dot input.
Both x and y participate in the generated-input fingerprint. Logical work is
`8*n+4` bytes per call: two input arrays and one scalar result, not measured DRAM
traffic. No output array is allocated.

## Call it from module-local Go

```go
value, err := reduction.DotSIMD([]float32{1, 2, 3}, []float32{4, 5, 6})
if err != nil {
    // Handle unequal lengths, e.g. errors.Is(err, reduction.ErrLengthMismatch).
    return
}
fmt.Println(value) // 32
```

Use `github.com/akolybelnikov/go-simd/internal/reduction` inside this module;
these experimental internal APIs cannot be imported from an external module.
Both kernels return +0 and ErrLengthMismatch before reading unequal-length
inputs. Equal empty/nil slices return +0,nil. Inputs are unchanged; read-only
aliases/overlaps are allowed. There is no caller output buffer.

The runnable example supplies imports and width-adaptive cancellation cases:

```powershell
go run ./examples/dot
```

An eight-lane process prints:

```text
simple: Scalar=32 SIMDStriped=32
lanes=8; cancellation A: Scalar=0 SIMDStriped=1
cancellation B: Scalar=1 SIMDStriped=0
unfused result bits: 0x00000000
length mismatch: true
```

## Inspect evidence

Each JSON observation retains ScalarResultBits and Key.Case.Reduction:
OrderID, actual process NativeLanes, VectorAccumulators (0/1), TailID. Dot uses
`standalone-dot/v1` or the explicit developer backend's `legacy-report-dot/v1`.
The native importer requires exact result-bits/simd-lanes metrics. Mixed widths,
operations and harnesses cannot silently pool into one comparison.

```powershell
$path = Get-ChildItem ./artifacts/dot/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
foreach ($row in $capture.run.Observations) {
    $bits = [uint32]$row.ScalarResultBits
    $value = [BitConverter]::ToSingle([BitConverter]::GetBytes($bits), 0)
    "$($row.Label): result=$value, bits={0:x8}" -f $bits
    $row.Key.Case.Reduction
}
```

Arithmetic results remain separate from timing medians. Successful measurements
retain the returned value once after timing; errors propagate and earlier
observations survive a later failure. Inspect the
[kernel](https://github.com/akolybelnikov/go-simd/blob/main/internal/reduction/dot.go)
and [detailed contract](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-dot.md)
for numerical rules, test coverage and the checkpoint boundary.
