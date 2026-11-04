# 24 — Three-array fused multiply-add

The muladd group extends the completed F004 baseline. The current catalogue has
**sixteen operations and forty-three implementations**. This experiment streams
three float32 input arrays and computes out[i]=round32(a[i]*b[i]+c[i]) with one
rounding. There are no broadcast coefficients and no reduction fold.

## Run it through any launcher

Go/Make commands need the checkout and GOEXPERIMENT=simd. Installed commands
use the same flags; the default report backend needs neither Go nor source.

```powershell
simdreport -list -group muladd
simdreport -group muladd -fma-input rounding -bench '^BenchmarkMulAdd$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group muladd -fma-input rounding -bench '^BenchmarkMulAdd$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./examples/muladd
```

```bash
make list ARGS="-group muladd"
make report COUNT=3 BENCHTIME=100ms ARGS="-group muladd -fma-input=rounding -bench BenchmarkMulAdd/n=64 -format md"
make run ARGS="-group muladd"
```

Markdown plus .measurements.json goes under artifacts/muladd. Quick comparison
uses the fixed payload size and saves no files. For a native developer smoke:

```powershell
go test ./internal/patterns -run '^$' -bench '^BenchmarkMulAdd$/^n=(8|16)$' -benchtime 1x
```

Use the report's -backend go-test to propagate configured recipes to the child.
The JSON catalogue also exposes all three operands: `simdlab catalog -group muladd`.

## Fused results and controls

| Label | Computation | Interpretation |
| --- | --- | --- |
| Scalar | Corrected software binary32 FMA32 for every element | A valid fused reference whose software cost is included |
| SIMDPartial | Process-verified portable vector MulAdd, or explicit software fallback | Actual gate/path is retained; label alone does not imply hardware acceleration |

For a=1+2^-23, b=1-2^-23, c=-1, fused output is -2^-46 (bits a8800000).
Separately rounding the product loses that residual. For a=MaxFloat32, b=2,
c=-MaxFloat32, fused output is MaxFloat32; an unfused product would overflow.
Run the example to see both, with the actual vector/software path.

Rounding is nearest ties-to-even once, including subnormals, signed zero,
infinities and overflow. NaN payload bits are unspecified. A float64 math.FMA
followed by narrowing is not a universal binary32 reference. Existing affine and
dot contracts remain distinct; compare their results as different experiments.

Module-local patterns.MulAddScalar/MulAddSIMD accept four []float32 slices and
return an error. All lengths must agree before writes. Empty equal/nil slices
succeed. Exact alias with any input is supported; partial overlap is outside the
contract. Benchmarks use disjoint arrays.

## Fixtures and honest measurement

- Balanced: bounded repeating triples.
- Rounding: cancellation, double-rounding-sensitive, signed-zero, subnormal and
  overflow-rescue triples.
- Mixed-magnitude: independently varying operand magnitudes.

Choose them with -fma-input. Alpha/beta and predicate/reduction settings are
irrelevant here. All three arrays contribute to the input fingerprint and the
recipe is retained. Shared agent plans keep only the relevant recipe settings.

Logical work is 16*n bytes: three float32 reads and one write. This is accounting,
rather than measured memory-bus traffic. Timing includes length checks, dispatch,
loads, FMA and stores. Allocation/fixtures and complete output hashing are outside
timing. Full output SHA-256, actual lane width and FMA policy/gate/path survive
in both report backends and isolated workers. Tiny retained 3x captures verify
this plumbing without establishing performance rankings.

See [the detailed contract and evidence](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-muladd.md)
and [the command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry).
