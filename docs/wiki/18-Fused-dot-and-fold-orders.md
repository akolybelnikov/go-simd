# 18 — Fused dot and fold orders

Checkpoint history: [three-array MulAdd](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add) is now implemented after the
original F004 closure. The current catalogue has nineteen operations/forty-nine
implementations; future/stop statements below describe this earlier checkpoint.

The dotfma group applies one float32 rounding to each x*y+accumulator update.
It offers Scalar and SIMDStriped/2/4, with explicit assignment/merge orders.
That checkpoint brought the catalogue to eleven operations and thirty implementations.
[Predicate byte export](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks) brought it to twelve/thirty-two; [packed predicates](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks) brought it to thirteen/thirty-four; [copy](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement) brought it to fourteen/thirty-eight; [clamp](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts) completes the current scope at fifteen/forty-one. The original
dot group remains unfused; this is not a single exact dot rounded at the end.

## Run a bounded experiment

```powershell
simdreport -list -group dotfma
simdreport -group dotfma -dot-input fusion-sensitive -bench '^BenchmarkDotFMA$/^n=64$' -count 3 -benchtime 100ms -format md
simdreport -group dot -dot-input fusion-sensitive -bench '^BenchmarkDot$/^n=64$' -count 3 -benchtime 100ms -format md
simdbench -group dotfma -dot-input mixed-magnitude
```

From a checkout with GOEXPERIMENT=simd, replace executables with
go run ./cmd/simdreport or go run ./cmd/simdbench. Make forwards the same flags:

```bash
make list ARGS="-group dotfma"
make report COUNT=3 BENCHTIME=100ms ARGS="-group dotfma -dot-input fusion-sensitive -bench BenchmarkDotFMA/n=64 -format md"
make run ARGS="-group dotfma"
```

The two report groups retain separate arithmetic/harness identities. Compare
them as separate experiments; do not merge fused/unfused observations into one
series. No new flag is required: -dot-input also selects dotfma's fixture.

## Fusion and fold order are separate decisions

Scalar starts at +0 and performs corrected software FMA32 updates left to right.
SIMDStriped/2/4 assign full vectors round-robin to 1/2/4 accumulators. Each lane
uses fused updates; remaining full vectors go to accumulators 0,1,2 as applicable.
After all full vectors, merge accumulators lane-wise left to right using ordinary
float32 additions, fold lanes low to high from +0, then apply software FMA32
updates to the scalar tail. No padded input lanes enter the reduction.

For n=2*w and other x/y entries zero:

| Nonzero entries | Fused Scalar / SIMDStriped | Fused SIMDStriped2/4 |
| --- | --- | --- |
| `x[0]=-1,y[0]=1,x[w]=1+2^-23,y[w]=1-2^-23` | -2^-46 | +0 |
| `x[0]=-max,y[0]=1,x[w]=max,y[w]=2` | max finite | +Inf |

The original unfused scalar dot gives +0 and +Inf respectively. Fusion can
preserve information lost by an intermediate rounded product, while splitting
updates across accumulators changes rounding again. These examples do not prove
a universal accuracy ranking. Tests use exact-rational fused updates plus
independently rounded merges/folds, with finite error bounds where applicable.

Signed zero also follows the order. For n=w, all x=smallest positive subnormal
and y=-0.25, fused Scalar returns -0. Vector lanes return -0 too, but folding
them from +0 yields +0. A later scalar tail update may produce -0 again.
NaNs/Inf and overflow follow the declared operations; NaN payloads are unspecified.

## Read actual per-case execution

The shared [FMA reference and gate](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding)
still apply. On software/emulated execution, each SIMD-labelled strategy retains
its own stripe order at the recorded process width; it does not become the
left-to-right scalar algorithm. A passing vector gate also does not mean a short
input executes vector FMA: inputs shorter than one vector use only software tails.

| PathID | Meaning |
| --- | --- |
| software-fma32/v1 | Scalar reference |
| vector-fma32/v1 | Full vector FMA updates with no scalar tail |
| software-striped-fma32/v1 | Software updates with the selected SIMD stripe/fold order |
| vector-fma32-with-software-tail/v1 | Native vector updates and corrected software tail |
| software-tail-fma32/v1 | Vector gate passes, but input has no full vector |

Case.FMA retains PathID and the process VectorPathID separately, plus RoundingID
and NativeLanes. Case.Reduction records OrderID, VectorAccumulators and TailID.
The count describes the algorithm, not physical registers. Process gates/widths
cannot pool; actual case paths can change across sizes within a sweep.

```powershell
$path = Get-ChildItem ./artifacts/dotfma/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
foreach ($row in $capture.run.Observations) {
    $row.Label
    $row.Key.Case.FMA
    $row.Key.Case.Reduction
    $bits = [uint32]$row.ScalarResultBits
    [BitConverter]::ToSingle([BitConverter]::GetBytes($bits), 0)
}
```

Harnesses are standalone-dot-fma/v1 and legacy-report-dot-fma/v1. Native child
metrics establish actual path/width/result facts; parent capabilities cannot
substitute. Numerical bits remain separate from timing medians. Validation,
dispatch, FMA updates, merge/fold and tail are timed; startup gate, preparation,
and post-measurement retention/metric encoding are outside. This compares whole
kernels, including software correction and compiler register/spill effects.

## Choose a fixture or call the kernels

- balanced, cancellation and mixed-magnitude use unchanged dot recipes.
- fusion-sensitive repeats 64 entries: x[0]=-1,y[0]=1 and
  x[32]=1+2^-23,y[32]=1-2^-23; otherwise x=y=0.

At n=64, fused Scalar gives -2^-46 versus unfused Scalar +0. Vector strategy
results depend on recorded width/assignment; four accumulators can give +0 at
width 16. Both arrays are fingerprinted. Fused/unfused groups share input values
with distinct generator/arithmetic identities. Alpha/beta, -fma-input and other
operation-specific controls do not change these fixtures.

Module-local calls are reduction.DotFMAScalar, DotFMASIMD, DotFMASIMD2 and
DotFMASIMD4, each accepting (x,y []float32) and returning (float32,error).
Empty equal/nil inputs return +0,nil; unequal lengths return +0,ErrLengthMismatch
before reading input. Inputs are unchanged and read-only overlap is supported.
No output array is used. Logical bytes are 8*n+4, not measured memory traffic.
Current widths allocate zero per call; future wider scratch requirements may
appear in counters.

Run the width-adaptive example:

```powershell
go run ./examples/dotfma
```

On the current eight-lane vector path it prints:

```text
vector gate: vector-fma32/v1; lanes: 8
unfused scalar bits: 00000000
fused Scalar bits: a8800000
fused SIMDStriped bits: a8800000
fused SIMDStriped2 bits: 00000000
fused SIMDStriped4 bits: 00000000
signed underflow: Scalar=80000000 SIMDStriped=00000000
```

Use GODEBUG=simd=0 in a fresh process to exercise software stripes. Read the
[kernels](https://github.com/akolybelnikov/go-simd/blob/main/internal/reduction/dotfma.go)
and [full contract](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-fused-dot.md).
This completes the fused-dot checkpoint; three-array MulAdd and M4 remain pending.

The [detailed contract](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-fused-dot.md)
links the signed implementation and retained hardware/software captures with
reproduction commands. These tiny captures verify plumbing, not performance.
