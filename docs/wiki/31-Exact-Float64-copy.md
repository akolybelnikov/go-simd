# 31 — Exact Float64 copy

Implemented 2026-11-04: `copy64` / `copy-f64` / `BenchmarkCopy64`, with Scalar,
GoCopy and SIMDPartial. Delivery checkpoint: **23 operations / 58 implementations**.

The experiment compares arithmetic-free binary64 movement through a source loop,
Go's built-in copy and explicit portable SIMD. Each output preserves every input
bit, including signed zeros, subnormals, infinities and all quiet/signaling NaN
payloads. No floating arithmetic, numeric conversion or precision widening occurs.

`movement.Copy64Scalar/Copy64Go/Copy64SIMD(x,out []float64) error` requires equal
lengths before writes. Empty/nil, disjoint storage and exact aliases are valid;
partial output overlap is unsupported by the shared contract. The GoCopy wrapper
obeys the same strict length policy even though the built-in has broader behavior.
SIMD bounds each partial load/store by actual Float64 lanes, including the
selected SDK's emulated partial-load length behavior. Measured buffers are disjoint.

Validation, width setup, loads/stores or source loop/Go copy are timed. Logical
traffic is `16*n` bytes: one binary64 read and write. Full input/output hashing,
fixture allocation and binding are outside timing. Source-loop and GoCopy labels
describe source strategies; the compiler/runtime may choose vector instructions.
SDK emulation does not disable GoCopy's independently selected implementation.
Logical bytes are effective work accounting, without physical cache/DRAM claims.

`generated-copy-float64-bits-v1` uses raw special-value/payload words followed by
deterministic mixed words. Complete little-endian input and output SHA256 hashes
must agree, including all bits of every NaN. Fixture/scalar flags for other groups
are irrelevant. A true unary Float64 binding reuses the typed retained output
adapter without fabricating a second input. Actual lanes and full digest survive
quick/native/compiled, catalogue/plan and isolated-worker paths under dedicated
harnesses. Native import requires positive integral lanes and complete digest
words; historical unknown input remains unknown. The bounded capture reader
rejects missing output/width facts. Comparisons use `copy-source-loop/v1`.

Tests compare raw uint64 representations over special/random values, actual-width
tails/nonmultiples, input preservation, canaries, exact aliases, pre-write errors
and zero current allocations. Typed fixtures/bindings, quick CLI, plans, isolated
workers, native/compiled/SDK-emulated parity, group isolation and all-group no-Go
standalone execution retain exact outputs. Existing Float64 SDK members are
reused; no new arithmetic or architecture-specific API is introduced.


## Run it

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./examples/copy64
simdreport -list -group copy64
simdreport -group copy64 -bench '^BenchmarkCopy64$/^n=64$' -count 3 -benchtime 100ms -format md
simdlab catalog -group copy64
```

Replace installed `simdreport` with `go run ./cmd/simdreport` from the checkout;
all flags stay the same. Make forwards those same options:

```bash
make list ARGS="-group copy64"
make report COUNT=3 BENCHTIME=100ms ARGS="-group copy64 -bench BenchmarkCopy64/n=64 -format md"
make run ARGS="-group copy64"
```

The example prints `8000000000000000` and `7ff0123456789abc`, preserving negative
zero and a signaling NaN payload. Read [the registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry)
for flags and [the contract/captures](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-copy64.md)
for evidence and limits.
