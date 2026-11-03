# 16 — Accumulator-count experiments

Sum and unfused dot now offer Scalar plus SIMDStriped, SIMDStriped2 and
SIMDStriped4. The suffix is the number of independent vector accumulators in
the algorithm. That checkpoint brought the catalogue to nine operations/twenty-four implementations.
[Fused affine](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding)
raised it to ten/twenty-six. [Fused dot](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders) raised it to eleven/thirty;
this page's sum/dot strategies preserve their existing arithmetic.

[Predicate byte export](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks) brought the catalogue to twelve/thirty-two; [packed predicates](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks) brought it to thirteen/thirty-four; [copy](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement) now brings it to fourteen/thirty-eight.

## Run the comparison

Installed commands use the same flags as Go/Make launchers:

```powershell
simdreport -list -group dot
simdreport -group dot -dot-input mixed-magnitude -bench '^BenchmarkDot$/^n=1024$' -count 3 -benchtime 100ms -format md
simdreport -group sum -sum-input cancellation -bench '^BenchmarkSum$/^n=64$' -count 3 -benchtime 100ms -format csv
simdreport -group dot -bench '^BenchmarkDot$/^n=1024$/^(Scalar|SIMDStriped2|SIMDStriped4)$' -count 3 -benchtime 100ms
simdbench -group dot
```

From source, enable GOEXPERIMENT=simd and replace the executable with
`go run ./cmd/simdreport` or `go run ./cmd/simdbench`. Make forwards the flags:

```bash
make list ARGS="-group sum"
make report COUNT=3 BENCHTIME=100ms ARGS="-group dot -bench BenchmarkDot/n=1024 -format md"
make run ARGS="-group sum -sum-input balanced"
```

There is no new accumulator flag. Use existing -bench variant selection for
reports; quick comparison runs all four strategies at its fixed size.
The three sum/dot fixture choices and their fingerprints are unchanged.

## What changes in the algorithm

SIMDStriped has one accumulation chain. SIMDStriped2/4 unroll two/four full
vectors and assign successive full vectors round-robin to independent vector
accumulators. If fewer than k full vectors remain after the main blocks, they
go to accumulators 0, 1, 2 in order. No input lanes are padded.

After all full vectors, merge accumulators **lane-wise, left to right**:

```text
2 accumulators: a0+a1
4 accumulators: ((a0+a1)+a2)+a3
```

Store the merged vector and fold its lanes low to high from +0, then process
the remaining scalar tail. Every addition rounds to float32. Dot also rounds
each product before adding it; Fused affine is available separately; these dot strategies remain unfused.
Length validation, loop control, merging, extraction/fold and tail are timed.

Independent chains may change throughput, but so do loop unrolling and Go's
register allocation/spills. The complete kernel comparison includes all those
effects. More accumulators need not win, especially on short inputs. There are
no fixed speedup thresholds and these timings do not isolate instruction latency.

## Rounding changes too

The mathematical sum/dot stays the same while the rounded operation order
changes. For n=8*w, other x entries zero and dot y all ones:

| Nonzero x positions | Scalar | SIMDStriped | SIMDStriped2 | SIMDStriped4 |
| --- | --- | --- | --- | --- |
| `x[0]=2^24, x[w]=1, x[2*w]=-2^24` | 0 | 0 | 1 | 0 |
| `x[0]=2^24, x[2*w]=1, x[4*w]=-2^24` | 0 | 0 | 0 | 1 |

Both exact mathematical results are 1. These examples do not establish an
accuracy ranking. NaNs/Inf and intermediate overflow also follow the declared
order. Tests verify independently rounded exact-rational operations and check
finite absolute error bounds where applicable; no universal relative tolerance
near cancellation is promised.

Run the width-adaptive example inside the module:

```powershell
go run ./examples/accumulators
```

An eight-lane process prints:

```text
float32 lanes: 8
layout 1 sum: Scalar=0 SIMDStriped=0 SIMDStriped2=1 SIMDStriped4=0
layout 1 dot: SIMDStriped=0 SIMDStriped2=1 SIMDStriped4=0
layout 2 sum: Scalar=0 SIMDStriped=0 SIMDStriped2=0 SIMDStriped4=1
layout 2 dot: SIMDStriped=0 SIMDStriped2=0 SIMDStriped4=1
```

## Evidence and kernel calls

Each observation's Key.Case.Reduction stores OrderID, NativeLanes,
VectorAccumulators (0,1,2,4) and TailID, with actual ScalarResultBits retained
separately from timing medians. VectorAccumulators is an algorithm description,
not a measurement of physical registers. Native width belongs to comparison
context; count/order are strategy dimensions. Keep different widths/harnesses
in separate comparisons. Existing sum/dot harnesses, input identities and old
strategy/fold IDs stay unchanged; the v1 evidence schema already holds these facts.

```powershell
$path = Get-ChildItem ./artifacts/dot/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
foreach ($row in $capture.run.Observations) {
    $row.Label
    $row.Key.Case.Reduction
    $bits = [uint32]$row.ScalarResultBits
    [BitConverter]::ToSingle([BitConverter]::GetBytes($bits), 0)
}
```

Module-local Go calls are SumSIMD2(x), SumSIMD4(x), DotSIMD2(x,y), DotSIMD4(x,y)
from internal/reduction. Dot returns (float32,error); handle ErrLengthMismatch
for unequal lengths. Equal empty inputs return +0; dot's error is nil. Inputs
are unchanged and dot read-only overlaps are supported. No output slice is used.
Logical bytes remain 4*n+4 for sum and 8*n+4 for dot; extra merge/scratch work
is included in timing, not counted as measured memory traffic.

Read the [kernels](https://github.com/akolybelnikov/go-simd/blob/main/internal/reduction/multi.go)
and [detailed contract](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-accumulators.md)
for exact order, numerical tests and checkpoint boundaries. Fused affine is available separately. Fused reductions, additional operation
families and M4 remain pending.
