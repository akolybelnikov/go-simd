# 26 — Uint8 saturating addition

saturatingadd8 is the first experiment with true uint8 input arrays. The
saturation checkpoint brought the catalogue to **eighteen operations and forty-seven
implementations**. The current catalogue is 23 operations /58 implementations; earlier counts describe delivery checkpoints.
The operation computes min(a+b,255), using a widened scalar sum and portable
SIMD AddSaturated. Numeric byte output is distinct from predicate masks.

## Run it

```powershell
simdreport -list -group saturatingadd8
simdreport -group saturatingadd8 -bench '^BenchmarkSaturatingAdd8$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group saturatingadd8 -bench '^BenchmarkSaturatingAdd8$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./examples/saturatingadd8
simdlab catalog -group saturatingadd8
```

```bash
make list ARGS="-group saturatingadd8"
make report COUNT=3 BENCHTIME=100ms ARGS="-group saturatingadd8 -bench BenchmarkSaturatingAdd8/n=64 -format md"
make run ARGS="-group saturatingadd8"
```

Go/Make requires the checkout and GOEXPERIMENT=simd. Installed commands use the
same flags; compiled reports need neither Go nor source at runtime. Summaries and
.measurements.json evidence go under artifacts/saturatingadd8. Native smoke:

```powershell
go test ./internal/integer -run '^$' -bench '^BenchmarkSaturatingAdd8$/^n=(16|32|64)$' -benchtime 1x
```

## Saturation versus wrapping

| a | b | Saturating result | Wrapping result, for explanation only |
| --- | --- | --- | --- |
| 1 | 2 | 3 | 3 |
| 128 | 127 | 255 | 255 |
| 128 | 128 | 255 | 0 |
| 255 | 1 | 255 | 0 |
| 255 | 255 | 255 | 254 |

Both registered controls promise the saturating column. A wrapping operation
would be a different arithmetic contract, not an equivalent benchmark baseline.
Scalar widens to uint16 before adding, so it cannot wrap before clipping.
SIMDPartial uses bounded Uint8 loads, AddSaturated and stores at actual width.
At 256 bits there are 32 byte lanes; listing and evidence record the actual count,
including separately initialized SDK emulation.

Module-local integer.SaturatingAdd8Scalar/SaturatingAdd8SIMD accept
(a,b,out []uint8) error. Unequal lengths return integer.ErrLengthMismatch before
writes. Empty/nil, disjoint output and exact aliases are supported; partial
overlap is outside the contract. Measurements use disjoint arrays.

## Fixtures, costs and evidence

A fixed boundary-rich recipe includes zeros, endpoints, saturated and ordinary
sums. There is no operation-specific flag and other fixture settings are
irrelevant. Both inputs are real bytes; no float32 conversion occurs.

Logical traffic is 3*n bytes. Timing includes length checks, width setup,
bounding, loads, saturation and stores. Allocation and full output hashing are
outside timing. ByteResultSHA256 covers every output byte; Case.Vector retains
actual Uint8 lanes and Case.Predicate is absent. Type/operation/arithmetic IDs
keep these results separate from canonical 0/1 masks and packed bits.

All 65,536 input pairs are tested in normal and emulated modes, with tails,
aliases, canaries, errors and allocations checked separately. Native/compiled
reports and isolated workers retain identical complete output evidence.
Tiny retained 3x captures verify plumbing, without asserting speedups or hardware
instruction selection. Other integer widths and wrapping operators remain future.

See [the detailed contract/evidence](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-saturating-add8.md)
and [the command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry).
