# 22 — Clamp and numerical contracts

Checkpoint history: [three-array MulAdd](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add) is now implemented after the
original F004 closure. The current catalogue has twenty-one operations/fifty-one
implementations; future/stop statements below describe this earlier checkpoint.

Clamp is F004's final slice. The current catalogue has **fifteen operations and
forty-one implementations**. It asks how branching, SIMD selection and Min/Max
compare when every strategy must return the same float32 bits.

## Try it

From the checkout, set GOEXPERIMENT=simd for Go/Make commands. Installed binaries
use the same flags and need no Go/source for the default report backend.

```powershell
simdreport -list -group clamp
simdreport -group clamp '-lower=-1.5' -upper 2.5 -pattern shuffled -seed 123 -bench '^BenchmarkClamp$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group clamp '-lower=-1.5' -upper 2.5 -pattern shuffled -seed 123 -bench '^BenchmarkClamp$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./examples/clamp
```

```bash
make list ARGS="-group clamp"
make report COUNT=3 BENCHTIME=100ms ARGS="-group clamp -lower=-1.5 -upper=2.5 -pattern=shuffled -seed=123 -bench BenchmarkClamp/n=64 -format md"
make run ARGS="-group clamp"
```

The report produces Markdown plus a .measurements.json evidence companion in
artifacts/clamp. The quick command uses one fixed-size comparison and saves no
files. Native developer benchmark:

```powershell
go test ./internal/patterns -run '^$' -bench '^BenchmarkClamp$/^n=(8|16)$' -benchtime 1x
```

Native scalar options travel through GO_SIMD_INPUT_OPTIONS when launched by
the report's -backend go-test. Prefer the report command for configured runs.

## What clipping means

For ordinary input [-2,-1,0,0.5,1,2], bounds [-1,1] give [-1,-1,0,0.5,1,1].
Values below lower become the exact lower bits; values above upper become the
exact upper bits. Equality, existing signed zeros and quiet/signaling NaN
payloads preserve the original input bits. Infinities follow ordered comparisons.

Bounds must satisfy lower<=upper and cannot be NaN. Equal bounds and infinite
bounds are valid. A useful bit-sensitive example is bounds (+0,-0):

| Input | Output bits | Why |
| --- | --- | --- |
| -1 | 00000000 (+0) | Below lower |
| +1 | 80000000 (-0) | Above upper |
| -0 | 80000000 (-0) | Equality keeps input |
| +0 | 00000000 (+0) | Equality keeps input |
| NaN payload 7f812345 | 7f812345 | Unordered input keeps its bits |

Run examples/clamp to see those rules for all three controls. Module-local
patterns.ClampScalar, ClampSIMD and ClampSIMDMinMax accept
(x,out []float32,lower,upper float32) error. Length mismatch takes precedence;
invalid bounds fail before writes, even on empty slices. Empty equal arrays
with valid bounds succeed. Disjoint/exact alias are supported; partial overlap
is outside the contract. Measured arrays are disjoint.

## Strategies and measured costs

| Label / strategy ID | Method | Included work |
| --- | --- | --- |
| Scalar / scalar | Ordered branches | Length/bounds validation, loop and writes |
| SIMDPartial / simd-partial | Comparisons plus IfElse | Validation, per-call broadcasts, partial loads/selections/stores |
| SIMDMinMax / simd-minmax | Max then Min with explicit repairs | All setup plus comparisons, mask Or, tie/NaN/outside-bit repairs and stores |

Raw Min/Max can select different operands for NaNs and ties. The last control
repairs that behavior inside timing. Its name does not promise a two-instruction
clamp or any speedup. Logical traffic is 8*n bytes; this is an algorithmic
read/write count, rather than measured memory-bus traffic.

## Fixtures and evidence

Lower/upper default to -1/1. The shared -pattern flag describes **clipped positions**
for clamp: none/all, even indices for alternating, first ceil(n/2) for clustered,
or approximately 50% deterministic shuffled positions selected by -seed (default 42).
Nonclipped values include boundary ties, valid signed zeros and payload NaNs.
Bounds [-Inf,+Inf] require -pattern none because no input can be outside them.
NaN/reversed bounds or an impossible clipped distribution fail before artifacts.

The run records exact lower/upper float32 bits, recipe/seed, input fingerprint,
actual SDK width, ordered-clamp-preserve-bits/v1 policy and full output SHA-256
over little-endian float32 bits after timing. Different widths/policies stay in
separate comparison contexts. Compiled and native reports have distinct harnesses.

Tests verify every output bit against independent scalar logic at actual lane
boundaries, special values and random raw bits. Native/compiled parity covers
all patterns, zero-sign bounds and the full infinite interval, including SDK
emulation. Tiny retained captures verify plumbing and make no timing claims.

See the [contract and evidence](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-clamp.md)
and [F004 completion boundary](https://github.com/akolybelnikov/go-simd/wiki/23-F004-completion-and-next-features).
