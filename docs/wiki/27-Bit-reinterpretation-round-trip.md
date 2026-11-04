# 27 — Bit reinterpretation round trip

bitroundtrip adds an exact Float32→Uint32→Float32 representation path. The
roundtrip checkpoint brought the catalogue to **nineteen operations and forty-nine
implementations**. [XOR bits](https://github.com/akolybelnikov/go-simd/wiki/28-XOR-of-float32-bit-patterns) now brings it to twenty/fifty-one.
This is bit reinterpretation, not numeric conversion: the result retains the
same 32-bit representation, including NaN payloads and signed zero.

## Run it

```powershell
simdreport -list -group bitroundtrip
simdreport -group bitroundtrip -bench '^BenchmarkBitRoundTrip$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group bitroundtrip -bench '^BenchmarkBitRoundTrip$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./examples/bitroundtrip
simdlab catalog -group bitroundtrip
```

```bash
make list ARGS="-group bitroundtrip"
make report COUNT=3 BENCHTIME=100ms ARGS="-group bitroundtrip -bench BenchmarkBitRoundTrip/n=64 -format md"
make run ARGS="-group bitroundtrip"
```

Go/Make requires the checkout and GOEXPERIMENT=simd. Installed commands accept
the same flags; default compiled reports need neither Go nor source. Summary
and .measurements.json go under artifacts/bitroundtrip. Native developer smoke:

```powershell
go test ./internal/patterns -run '^$' -bench '^BenchmarkBitRoundTrip$/^n=(8|16)$' -benchtime 1x
```

## Exact bits and meaningful interpretation

| Input bits | Meaning | Output bits |
| --- | --- | --- |
| 00000000 | +0 | 00000000 |
| 80000000 | -0 | 80000000 |
| 00000001 | Smallest positive subnormal | 00000001 |
| 7f800000 | +infinity | 7f800000 |
| 7fc12345 | Quiet NaN payload | 7fc12345 |
| 7f812345 | Signaling NaN payload | 7f812345 |

Scalar uses math.Float32bits/math.Float32frombits. SIMDPartial uses ToBits and
BitsToFloat32 around bounded vector loads/stores. Neither path rounds numeric
values or produces integer magnitudes. A numeric conversion would need separate
range, rounding and special-value rules and is still a future experiment.

The compiler may eliminate representational steps. A fast result does not prove
a bitcast instruction ran or establish its latency. This comparison measures
the complete kernel, including validation, width setup, loads and stores. The
existing copy group remains a separate movement experiment with GoCopy control.

## Contract and retained evidence

Module-local patterns.BitRoundTripScalar/BitRoundTripSIMD accept
(x,out []float32) error. Length mismatch fails before writes. Empty/nil,
disjoint output and exact aliases succeed; partial overlap is outside the contract.
Measured buffers are disjoint. Logical work is 8*n bytes, not physical bandwidth.

The fixed raw-bit recipe covers zeros, subnormals, infinities and quiet/signaling
NaNs. It uses a separate generator/operation identity from copy; other fixture
flags are irrelevant. Every input/output bit contributes to the corresponding
little-endian SHA-256, computed outside timing. Case.Vector retains actual
Float32 lanes, and copy-specific policy fields remain absent.

Tests check special/random bits, tails, aliases, canaries, input preservation,
errors and allocations. Native/compiled/emulated reports, agent plans and workers
retain matching complete digests. Tiny 3x captures verify plumbing without timing
claims. Source labels and an SDK type's availability do not prove acceleration.

See [the detailed contract/evidence](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-bit-roundtrip.md)
and [the command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry).
