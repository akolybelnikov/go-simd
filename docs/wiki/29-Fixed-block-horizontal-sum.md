# 29 — Fixed-block horizontal sum

Implemented 2026-11-04: `sum8` / `sum8-f32` / `BenchmarkSum8`, Scalar and
SIMDPartial. Delivery checkpoint: **21 operations / 53 implementations**.

The experiment asks how a portable SIMD pair stage compares to scalar code for
an explicitly matched eight-input horizontal fold. It is a composed algorithm;
architecture-specific permutation and horizontal instructions remain future work.

`reduction.Sum8Scalar/Sum8SIMD(x,out []float32) error` emits one result per
consecutive block of eight inputs. Pad the final block with +0. Round float32
pairs `(x0+x4,x1+x5,x2+x6,x3+x7)`, then fold those four pairs left to right,
rounding after each addition. Both implementations and all native widths use
exactly this order. NaN payloads are unspecified; signed zeros follow the padded
order. This is different from summing an entire array or folding each block
strictly left to right. For `[-1e20,1,1e20,0,1e20,0,-1e20,0,2]`, outputs are `[1,2]`.

Output shape is `blocks-of-8`: exactly `ceil(n/8)` values. Validate output length
before writes; empty/nil, disjoint output and same-start output-prefix aliases
are supported, other overlap is unsupported. Fixtures are disjoint when measured.
Validation, native-width setup, block initialization/padding/copying, loads,
pair addition, intermediate stores and the scalar final fold are timed. Logical
traffic is `4*n+4*ceil(n/8)` bytes, not physical cache/DRAM traffic.

The existing `-sum-input balanced|cancellation|mixed-magnitude` recipes use a
separate `generated-sum8-float32-v1/<recipe>` namespace. Input hash covers every
input bit. The full projected output digest, actual Float32 lanes and physical
output count survive plans, worker execution and both report backends. Native
import requires integral positive lanes, exact output count and all digest words;
unknown historical recipe identity remains unknown. Comparisons allow extents
to vary across sizes and reject contradictions at the same size.

Tests use an independent exact-rational per-add rounding oracle, random/raw
special values, cancellation, fixed/native boundaries, partial blocks, canaries,
aliases, input preservation, pre-write errors and zero current allocations.
Backend/SDK-emulated parity, projected plans, isolated workers, bounded evidence
reading and no-Go standalone execution exercise the full typed path. The MulAdd
comparison baseline is also corrected to its fused software reference policy.


## Run the experiment

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/sum8
simdreport -list -group sum8
simdreport -group sum8 -sum-input cancellation -bench '^BenchmarkSum8$/^n=64$' -count 3 -benchtime 100ms -format md
simdlab catalog -group sum8
```

Use `go run ./cmd/simdreport` in place of installed `simdreport` with identical
flags. From a checkout:

```bash
make list ARGS="-group sum8"
make report COUNT=3 BENCHTIME=100ms ARGS="-group sum8 -sum-input cancellation -bench BenchmarkSum8/n=64 -format md"
make run ARGS="-group sum8 -sum-input cancellation"
```

Reports always retain raw measurements alongside the selected presentation.
Read the [command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry)
for shared flags and the [source contract](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-sum8.md)
for retained captures.
