# 28 — XOR of Float32 bit patterns

xorbits applies uint32 XOR to two opaque float32 representations. The
catalogue now has **twenty-one operations and fifty-three implementations**.
This is a representation-level operation; it performs no floating arithmetic
or numeric conversion and does not claim direct uint32 CLI input support.

## Run it

```powershell
simdreport -list -group xorbits
simdreport -group xorbits -bench '^BenchmarkXorBits$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -group xorbits -bench '^BenchmarkXorBits$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./examples/xorbits
simdlab catalog -group xorbits
```

```bash
make list ARGS="-group xorbits"
make report COUNT=3 BENCHTIME=100ms ARGS="-group xorbits -bench BenchmarkXorBits/n=64 -format md"
make run ARGS="-group xorbits"
```

Go/Make requires GOEXPERIMENT=simd and the checkout. Installed commands use the
same flags; default compiled reports need neither Go nor source. Summary and
.measurements.json go under artifacts/xorbits. Native developer smoke:

```powershell
go test ./internal/patterns -run '^$' -bench '^BenchmarkXorBits$/^n=(8|16)$' -benchtime 1x
```

## Bits define the result

| a bits | b bits | Output bits | Interpretation |
| --- | --- | --- | --- |
| 3f800000 (+1) | 80000000 (sign bit) | bf800000 | -1 |
| c0000000 (-2) | 80000000 (sign bit) | 40000000 | +2 |
| 7f812345 | 00000000 | 7f812345 | Original signaling NaN payload |
| 80000000 | 80000000 | 00000000 | -0 bits XOR -0 bits becomes +0 |

Scalar extracts uint32 bits, applies XOR and returns the resulting representation.
SIMDPartial uses ToBits/Xor/BitsToFloat32 around bounded loads/stores. Any 32-bit
result is allowed, even if it encodes a NaN. Do not interpret this as adding
floating values. The second array is a streamed operand, not a scalar broadcast.

Module-local patterns.XorBitsScalar/XorBitsSIMD accept (a,b,out []float32) error.
Length errors precede writes; empty/nil, disjoint output and exact aliases work.
Partial overlap is outside the contract; measured arrays are disjoint.

## Costs and evidence

Logical work is 12*n bytes: two float32 reads and one write. Timing includes
validation, width setup, bounded loads, bitwise path and stores. A fixed raw-bit
recipe includes sign/exponent/payload masks; other fixture settings are irrelevant.
Complete hashes retain both operands and every output bit outside timing, and
Case.Vector records actual lanes. Type/operation/arithmetic identities keep
raw-word logic separate from numeric arithmetic and predicate output.

Tests compare each result bit with an independent Boolean truth oracle, with
random/special words, tails, aliases, canaries, errors and allocations. Both
backends, SDK emulation and isolated workers preserve the same full output.
Tiny 3x captures verify plumbing without speedup claims. Source labels and timings
alone do not establish emitted instructions, and bitcasts may be optimized away.

See [the detailed contract/evidence](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-xor-bits.md)
and [the command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry).
