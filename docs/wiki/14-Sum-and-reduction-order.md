# 14 — Sum and reduction order

Checkpoint history: [three-array MulAdd](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add) is now implemented after the
original F004 closure. The current catalogue has twenty operations/fifty-one
implementations; future/stop statements below describe this earlier checkpoint.

Sum's F004 checkpoint added one float32 input array and one returned float32,
bringing the catalogue to eight operations/eighteen implementations. The later
[dot checkpoint](https://github.com/akolybelnikov/go-simd/wiki/15-Dot-product-and-unfused-rounding)
raised that to nine/twenty; [2/4-accumulator strategies](https://github.com/akolybelnikov/go-simd/wiki/16-Accumulator-count-experiments)
brought it to nine/twenty-four. [Fused affine](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding)
brought it to ten/twenty-six. [Fused dot](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders) brought it to eleven/thirty. This page specifies the original sum controls.

[Predicate byte export](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks) brought the catalogue to twelve/thirty-two; [packed predicates](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks) brought it to thirteen/thirty-four; [copy](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement) brought it to fourteen/thirty-eight; [clamp](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts) completes the current scope at fifteen/forty-one.

## Run sum

Installed commands:

```powershell
simdreport -list -group sum
simdreport -group sum -sum-input cancellation -bench '^BenchmarkSum$/^n=64$' -count 3 -benchtime 100ms -format md
simdreport -group sum -sum-input mixed-magnitude -bench '^BenchmarkSum$/^n=1024$' -count 3 -benchtime 100ms -format csv
simdbench -group sum -sum-input balanced
```

From a checkout, replace the executable with `go run ./cmd/simdreport` or
`go run ./cmd/simdbench` and enable GOEXPERIMENT=simd. Make uses the same flags:

```bash
make list ARGS="-group sum"
make report COUNT=3 BENCHTIME=100ms ARGS="-group sum -sum-input cancellation -bench BenchmarkSum/n=64 -format md"
make run ARGS="-group sum -sum-input balanced"
```

`-sum-input` is shared by both CLIs. The quick comparison remains fixed at
1,048,576 elements. Regex selects registered sizes; `n=64` is anchored in the
direct example so it does not select other containing names.

## The two strategies

| Strategy | Order |
| --- | --- |
| Scalar | Start at +0; accumulate input left to right, rounding each addition to float32 |
| SIMDStriped | Accumulate full vectors in one vector register; extract and fold lanes low to high from +0; then accumulate the remaining scalar tail |

The final extraction/fold belongs to the timed kernel. SIMD does not add padded
input lanes. It returns a real float32; no one-element output slice is used.
Both kernels leave input unchanged. Nil/empty input returns positive zero.
Logical work is `4*n + 4` bytes, including the returned scalar. This is effective
work accounting, not hardware memory bandwidth.

Floating-point addition is order-sensitive. Using w native lanes and an array
of length 2*w with all other entries zero:

| Nonzero positions | Scalar | SIMDStriped | Exact mathematical sum |
| --- | --- | --- | --- |
| x[0]=2^24, x[1]=1, x[w]=-2^24 | 0 | 1 | 1 |
| x[0]=2^24, x[1]=-2^24, x[w]=1 | 1 | 0 | 1 |

Neither algorithm is always more accurate. The table's percentage change measures
time relative to scalar; it does not score numerical error. Do not require
cross-strategy bit equality or present scalar as an exact mathematical oracle.
Tests verify each specified order against exact rational float32 rounding and
check a finite absolute forward-error bound where its assumptions apply.
There is no fixed relative-error guarantee near cancellation.

NaNs, infinities and intermediate overflow follow the selected order. Finite
inputs can overflow differently: a scalar fold may remain finite while striped
accumulators reach opposite infinities and yield NaN. NaN payloads are not fixed.
All-negative-zero inputs return +0 under the initial-value rules. Results are
retained as bits, so non-finite arithmetic does not break JSON reports.

## Choose an input fixture

| `-sum-input` | Pattern |
| --- | --- |
| balanced (default) | Signed quarter values, x[i]=float32(i%33-16)/4 |
| cancellation | Repeat 64 entries: 2^24 at index 0, 1 at 1, -2^24 at 32, otherwise zero |
| mixed-magnitude | Repeat float32 [1e-20,1,1e20,-1e20,-1,3,-3,1e-20] |

The cancellation fixture exposes an order difference at n=64 on the currently
tested 4/8/16-lane targets. Exact results depend on the recorded width and size.
Pattern/seed/alpha/beta controls from other operations do not alter sum. Input
identity contains the sum fixture name and the fingerprint of x alone.

## Inspect returned values and fold identity

The Markdown/CSV/TXT table summarizes timings. Individual arithmetic results stay
separate in each JSON observation's `ScalarResultBits`; they are not averaged.
`Key.Case.Reduction` records OrderID, NativeLanes, VectorAccumulators and TailID.
NativeLanes records the process width for both strategies; scalar still uses no
vector accumulator. Mixed native widths cannot be pooled into one comparison.

```powershell
$path = Get-ChildItem ./artifacts/sum/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
foreach ($row in $capture.run.Observations) {
    $bits = [uint32]$row.ScalarResultBits
    $value = [BitConverter]::ToSingle([BitConverter]::GetBytes($bits), 0)
    "$($row.Label): result=$value, bits={0:x8}" -f $bits
    $row.Key.Case.Reduction
}
```

The sum harnesses are `standalone-reduction/v1` and
`legacy-report-reduction/v1`. Retention happens once after timing; every kernel
call still computes the complete result. Go controls calibration and GC.
The native backend reports exact integer result-bits and simd-lanes metrics;
the importer rejects missing/fractional/out-of-range sum metrics. Add
`-backend go-test` for recorded native comparisons with a custom sum fixture;
that development option needs the checkout and Go.

## Runnable example

From the checkout:

```powershell
go run ./examples/sum
```

On an eight-lane process it prints:

```text
simple: Scalar=6 SIMDStriped=6
lanes=8; cancellation A: Scalar=0 SIMDStriped=1
cancellation B: Scalar=1 SIMDStriped=0
empty result bits: 0x00000000
```

The example adapts its counterexamples to the actual width. Read the
[sum kernel](https://github.com/akolybelnikov/go-simd/blob/main/internal/reduction/sum.go)
and [detailed implementation contract](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-sum.md)
for exact rounding, retention, byte-accounting and verification boundaries.
