# F004 extension — Low64 carryless polynomial multiplication

Implemented 2026-11-04: `clmul64` / `clmul-low64-u64` / `BenchmarkClmul64`,
Scalar/SIMDPartial. Delivery checkpoint: **22 operations / 55 implementations**.

This experiment compares scalar polynomial multiplication with the portable SDK
including the costs of extracting useful lane results. Each true uint64 output
is the low64 product of its two inputs over GF(2), equivalently modulo x^64.
Coefficients add with XOR; no integer carry propagates. Thus 3 times 3 is 5.
High product bits are deliberately discarded. This is a representation and
algorithm study, without cryptographic or full128-product claims.

`integer.Clmul64Scalar/Clmul64SIMD(a,b,out []uint64) error` validates equal
lengths before writes. Empty/nil, disjoint output and exact aliases are valid;
partial output overlap is unsupported. SIMD supports current SDK Uint64 widths
2/4/8, rejects unsupported widths before writes, and bounds partial loads to
actual width. Scalar shifts/XORs the operand; SIMD calls CarrylessMultiplyEven
and CarrylessMultiplyOdd. Each SDK product fills an even/odd pair of low/high
words; the kernel extracts the low word for each corresponding input lane.
Both SDK calls, scratch initialization/stores, extraction and output writes are
timed, together with validation, width setup and loads. Logical traffic is
`24*n` bytes: two uint64 reads and one output write. Physical cache/DRAM traffic
and individual instruction costs are outside this measurement contract.

The fixed `generated-clmul-uint64-v1` recipe contains all combinations of zero,
one, 3, high-bit, all-bit and alternating/raw words, followed by deterministic
mixed words. Both complete little-endian inputs and every output word are
hashed outside timing. No floating conversion occurs. Unrelated scalar/fixture
flags have no effect.

Actual Uint64 lanes and HasHardwareCarrylessMultiply capability are retained
in Case.Vector for both strategies under dedicated standalone/native harnesses.
Capability is `sdk-hardware-clmul-available/v1` or `sdk-software-clmul-only/v1`;
it does not prove emitted instructions. SDK emulation is a separate machine fact.
The SDK controls defaults and emulation. Comparisons reject mixed capabilities.
Native import requires exact capability 0/1, supported integral lanes and all
digest words. Bounded capture reading rejects incomplete facts; unknown
historical input remains unknown.

Independent coefficient convolution checks all 65,536 small 8-bit operand pairs
plus random full-width words, truncation and special boundaries. Tests cover
actual-width tails, canaries, aliases, unchanged inputs, pre-write errors and
zero current allocations. Typed fixtures/bindings, quick CLI, plans, isolated
workers, native/compiled/SDK-emulated parity and no-Go standalone execution
retain full results. Uint64s and bounded load/store/Even/Odd are exercised in
the SDK ledger; actual width and hardware capability are infrastructure use.

Retained [normal](evidence/clmul64-vector.measurements.json) and
[SDK-emulated](evidence/clmul64-sdk-emulated.measurements.json) captures each
contain four observations at n=64, count 2, 3x. Every full digest agrees across
strategies/modes. Windows/amd64, Go1.27.3-X:simd, source base 9640ce7 plus
uncommitted carryless changes. Original timestamps/run IDs/samples retained;
see [capture notes](evidence/README.md). These make no performance claim.
