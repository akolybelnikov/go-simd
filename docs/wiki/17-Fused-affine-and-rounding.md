# 17 — Fused affine and float32 rounding

Checkpoint history: [three-array MulAdd](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add) is now implemented after the
original F004 closure. The current catalogue has seventeen operations/forty-five
implementations; future/stop statements below describe this earlier checkpoint.

The affinefma group computes alpha*x+beta with one final float32 rounding.
Existing affine still rounds its product before addition; dot remains unfused.
That checkpoint brought the catalogue to ten operations and twenty-six implementations.
[Fused dot](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders) brought it to eleven/thirty with separate update/fold semantics.

[Predicate byte export](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks) brought the catalogue to twelve/thirty-two; [packed predicates](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks) brought it to thirteen/thirty-four; [copy](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement) brought it to fourteen/thirty-eight; [clamp](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts) completes the current scope at fifteen/forty-one.

## Run a fused experiment

Installed commands:

```powershell
simdreport -list -group affinefma
simdreport -group affinefma -fma-input rounding '-alpha=1.0000001192092896' '-beta=-1' -bench '^BenchmarkAffineFMA$/^n=8$' -count 3 -benchtime 100ms -format md
simdbench -group affinefma -fma-input mixed-magnitude
```

Source runs replace the executable with go run ./cmd/simdreport or
go run ./cmd/simdbench, with GOEXPERIMENT=simd enabled. Make forwards the same flags:

```bash
make list ARGS="-group affinefma"
make report COUNT=3 BENCHTIME=100ms ARGS="-group affinefma -fma-input rounding -alpha=1.0000001192092896 -beta=-1 -bench BenchmarkAffineFMA/n=1024 -format md"
make run ARGS="-group affinefma"
```

The report compares Scalar with SIMDPartial under the fused contract. To study
unfused affine, use -group affine in a separate experiment. Their arithmetic/
harness identities prevent treating fused/unfused observations as one series.
Quote complete decimal flags as above in PowerShell to preserve the argument.

## One rounding changes values

For alpha=1+2^-23, x=1-2^-23 and beta=-1, the exact expression is -2^-46.
Unfused multiplication first rounds to 1, so affine returns +0. Fused affine
returns -2^-46, whose bits are a8800000. This is a different numerical contract,
not evidence that all fused/unfused outputs should agree or that fusion always
improves a complete algorithm's accuracy.

Correctly rounded float32 FMA also needs more care than widening math.FMA and
narrowing its result. With the same alpha/x and beta=16777218, the exact product
is 1-2^-46: correct float32 FMA returns 16777218. The widened/narrowed expression
returns 16777220 after landing on a midpoint at binary64 precision. Tests include
similar failures at subnormal and overflow boundaries.

The software reference uses an exact binary64 product, an error-free sum and a
round-to-odd correction before narrowing. Tests independently use exact rational
products/sums. Background on avoiding double rounding is credited to
[Sylvie Boldo and Guillaume Melquiond](https://guillaume.melquiond.fr/doc/08-tc.pdf),
with related [Flocq formal properties](https://flocq.gitlabpages.inria.fr/flocq/html/Flocq.Prop.Round_odd.html).
The Go implementation is tested; it does not inherit a formal verification claim.

## Read the actual execution path

The inspected SDK's emulated MulAdd uses ordinary multiply/add. The lab runs a
small startup conformance gate before choosing:

| Path | Meaning |
| --- | --- |
| vector-fma32/v1 | Native portable MulAdd passed the selected fused-rounding probes |
| software-fma32/v1 | Corrected software float32 FMA; used for emulation or a failed vector gate |

Scalar always uses the software reference. SIMDPartial may also use software;
its label alone does not establish vector acceleration. Detailed listing and
quick output state the path. Reports retain both the actual strategy path and
the process vector gate. A bounded startup probe is not exhaustive proof or a
speedup guarantee; the test suite separately checks numerical behavior.

To run the explicit software path in a fresh PowerShell process/environment:

```powershell
$env:GODEBUG = 'simd=0'
go run ./examples/fma
go run ./cmd/simdreport -group affinefma -fma-input rounding '-alpha=1.0000001192092896' '-beta=-1' -bench '^BenchmarkAffineFMA$/^n=8$' -count 2 -benchtime 3x -format md
Remove-Item Env:GODEBUG
```

Restore any prior GODEBUG value if you had one. Binaries also honor this setting
at startup. The probe is outside timing; validation, path dispatch, arithmetic,
vector broadcasts when used, and output writes are inside. Scalar includes the
software correction cost. This is a whole-kernel comparison.

## Fixtures and module-local calls

Both CLIs accept -fma-input, default balanced; -alpha/-beta keep their existing
defaults 1.5/-0.5. The chosen group defines fused versus unfused arithmetic.

| Fixture | Values |
| --- | --- |
| balanced | Signed quarters: float32(i%33-16)/4 |
| rounding | Repeat [1-2^-23,1+2^-23,1,-1,smallest subnormal,max finite,-0,1-2^-23] |
| mixed-magnitude | Repeat float32 [1e-20,1,1e20,-1e20,-1,3,-3,1e-20] |

Only x plus output are allocated; logical work is 8*n bytes, not three input
arrays or measured memory traffic. Alpha/beta bits and x's fixture/hash survive
in identity. Other operation-specific fixture settings do not alter this input.

Inside this module, patterns.AffineFMAScalar and patterns.AffineFMASIMD accept
(x,out []float32,alpha,beta float32) error. Equal empty/nil slices succeed;
length mismatch returns patterns.ErrLengthMismatch without changing output.
Disjoint output or exact in-place x is supported; partial overlap is unsupported.
Subnormals, signed zero, overflow and infinities follow one-rounding IEEE rules;
NaN payloads are not fixed. These internal packages are not a public import API.

Run the complete example:

```powershell
go run ./examples/fma
```

On a process passing the vector gate:

```text
FMA32 SIMD path: vector-fma32/v1
unfused bits: 00000000; fused scalar: a8800000; fused SIMD: a8800000
double-rounding example: widened=1.677722e+07 corrected=1.6777218e+07
```

## Inspect evidence

machine.fma32_path records the process gate. Key.Case.FMA records RoundingID,
PathID, VectorPathID and actual NativeLanes; scalar records its software path
alongside the same process vector capability. Mixed vector paths/widths cannot
pool. Observation.ArrayResultLastBits retains the last output element as uint32,
including NaN/Inf safely; it is not a full-array checksum or a timing average.

```powershell
$path = Get-ChildItem ./artifacts/affinefma/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
$capture.machine.fma32_path
foreach ($row in $capture.run.Observations) {
    $row.Label
    $row.Key.Case.FMA
    $bits = [uint32]$row.ArrayResultLastBits
    [BitConverter]::ToSingle([BitConverter]::GetBytes($bits), 0)
}
```

Harnesses are standalone-affine-fma/v1 and legacy-report-affine-fma/v1. The
explicit developer backend records strict child path/width/last-result metrics;
it does not infer child FMA support from the parent. No report backend silently
falls back to invoking Go. Older unrelated evidence remains readable.

Read the [kernel](https://github.com/akolybelnikov/go-simd/blob/main/internal/patterns/fma.go),
[software primitive](https://github.com/akolybelnikov/go-simd/blob/main/internal/fma32/fma.go)
and [detailed contract](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-fma.md).
[Fused dot](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders) is now implemented as a separate group with Scalar and
1/2/4-vector accumulators. Three-array MulAdd, other families and M4 remain future work.
