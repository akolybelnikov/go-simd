# 25 — Float64 addition

addition64 is the first operation over true []float64 inputs/output. The
catalogue now has **seventeen operations and forty-five implementations**.
The original addition group remains float32 with its original strategies.

## Run through any launcher

```powershell
simdreport -list -group addition64
simdreport -group addition64 -bench '^BenchmarkAddition64$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group addition64 -bench '^BenchmarkAddition64$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./examples/addition64
simdlab catalog -group addition64
```

```bash
make list ARGS="-group addition64"
make report COUNT=3 BENCHTIME=100ms ARGS="-group addition64 -bench BenchmarkAddition64/n=64 -format md"
make run ARGS="-group addition64"
```

Go/Make commands need the checkout and GOEXPERIMENT=simd. Installed commands
use the same flags and default compiled reports need neither Go nor source.
The report saves a summary plus .measurements.json under artifacts/addition64.
Native developer smoke:

```powershell
go test ./internal/addition -run '^$' -bench '^BenchmarkAddition64$/^n=(8|16)$' -benchtime 1x
```

## What changes with Float64?

At 256 bits, float32 holds eight lanes and float64 four. Actual width comes
from the selected SDK; emulation uses its own recorded width. The values
16777216+1 illustrate precision: float32 rounds back to 16777216, while float64
returns 16777217. This is an element-type change, rather than a SIMD/scalar
disagreement. Run the example to see the values and actual lane counts.

| Group | Input/output | Controls | Logical bytes/call |
| --- | --- | --- | --- |
| addition | float32 | Scalar, SIMDPartial, SIMDTail, SIMDParallel | 12*n |
| addition64 | float64 | Scalar, SIMDPartial | 24*n |

For a focused type study, select Scalar/SIMDPartial in each group. Equal element
counts mean twice the logical bytes for float64; equal-byte studies need different
n. Retain both contexts and inspect separate results. Neither these byte counts
nor timing ratios prove physical memory bandwidth or a bottleneck.

## Numerical and storage rules

Both controls perform ordinary binary64 addition, including signed-zero,
subnormal, infinity and overflow behavior. NaN payload bits are unspecified.
Fixed fixtures generate decimal fractions directly as float64; there is no
float32 widening. Scalar-parameter/predicate/reduction flags are irrelevant.

Module-local addition.Add64Scalar/Add64SIMD accept (a,b,out []float64) error.
Length mismatch returns addition.ErrLengthMismatch before any write. Valid
empty/nil succeeds. Disjoint output/exact aliases are supported; partial output
overlap is outside the contract. Benchmarks use disjoint arrays.

SIMDPartial uses bounded partial loads/stores with the actual Float64 lane count.
This also avoids a selected-SDK emulation issue where an unbounded partial load
can report more elements than it loads. Boundary tests cover that mode explicitly.
Length checks, width setup, bounding, loads, addition and stores are timed.
Fixture allocation and complete output hashing are outside timing.

## Evidence and limits

Input fingerprints retain every binary64 bit from both operands. Full output
SHA-256 uses little-endian float64 bits. Case.Vector.NativeLanes retains actual
float64 lanes, while the existing machine float32_lanes field remains a separate
runtime fact. Reports, JSON catalogue/plans and isolated workers agree on true
types, widths and 24*n work. Native import requires complete output digest words
and actual integral lanes; incompatible type/width contexts do not pool silently.

Tests use independent exact-rational rounding, special/random values, tails,
canaries, aliases, errors, allocation checks and compiled/native/emulated/worker
parity. Tiny retained captures verify evidence plumbing without speedup claims.
Other Float64 operations and imported datasets remain future work.

See [the detailed contract/evidence](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-addition64.md)
and [the complete command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry).
