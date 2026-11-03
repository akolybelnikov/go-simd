# F004, slice 6 — Packed predicate export

Historical checkpoint record. The final [F004 acceptance record](F004-completion.md)
now closes the agreed scope after clamp; counts and stop decisions below describe
this earlier slice.

Implemented 2026-11-03 after canonical byte masks. Group predicatepack adds Scalar
and SIMDPartial: thirteen operations/thirty-four implementations. This checkpoint
answers how much conversion and packing cost when output occupies one bit per
predicate. Stop after packed export; clamp, copy/reorder, three-array MulAdd,
standalone conversions, other input types and M4 remain separate work.

Retained [hardware capture](evidence/packed-predicate-vector.measurements.json),
[emulated capture](evidence/packed-predicate-software.measurements.json) and
[commands/provenance](evidence/README.md) preserve observed results. They verify
plumbing at fixed iteration counts without performance claims.

Implementation: signed commit [d61d905](https://github.com/akolybelnikov/go-simd/commit/d61d905a94dca2d8e8e1745ca4ba7d7a03ae62aa).

## Representation contract

Operation predicate-gt-packed-f32, benchmark BenchmarkPredicatePacked, arithmetic
float32-predicate-gt-packed-lsb0-v1. Input is x []float32 and scalar threshold;
output is []uint8 with exactly ceil(len(x)/8) elements.

PredicateGTPackedScalar and PredicateGTPackedSIMD in internal/patterns accept
`(x []float32,out []uint8,threshold float32) error`.

- Input i maps to bit i%8 in byte i/8, least-significant bit first.
- Ordered x[i] > threshold sets that bit; equality or NaN clears it.
- Signed zeros compare equal at zero; infinities/subnormals follow Go comparison.
- Unused high bits of the last byte are always zero.
- Every output byte is overwritten each call, including all-false results.
- Wrong output length returns patterns.ErrLengthMismatch before any write.
- Equal empty/nil input/output succeeds. Input bits are preserved; storage must
  be disjoint. Cross-type aliases made with unsafe are outside the contract.

For x=[1,0,0,1,0,0,1,0,1] at threshold zero, the canonical byte mask is
[1,0,0,1,0,0,1,0,1]; packed output is [0x49,0x01]. A single true input produces
0x01, never 0xff. The result does not depend on native vector width.

## Algorithms and timing

Scalar builds each output byte from up to eight ordered comparisons and stores
it once. SIMDPartial broadcasts threshold/one/zero, loads partial vectors,
compares, selects explicit float32 0/1, converts to int32 scratch and packs valid
lanes into bytes. A partial output byte carries across vector boundaries.
This matters on four-lane emulation and remains defined for other widths.

There is no raw-mask interpretation or portable Mask32s.ToBits assumption.
Conversion, extraction, packing and all byte writes are inside the measured
kernel. Padded input zeros may compare true for negative thresholds, but invalid
lanes never set output bits. Current widths use sixteen stack int32 slots and
allocate zero per call; future wider vectors may allocate scratch in timing.

Typed BindUint8 and Prepared.ByteOutput are reused. A new PackedBits descriptor
shape and OutputElements helper distinguish logical input length from physical
storage. Rounded byte counts divide before adding, avoiding n+7 overflow.
Logical work is `4*n+ceil(n/8)` bytes, including all input/output elements; it
does not count scratch traffic or measure DRAM transactions. Negative sizes and
overflowed work counts fail. Ordinary array/scalar counts and sweeps are preserved.

Separate standalone-packed-predicate/v1 and legacy-report-packed-predicate/v1
harness identities keep the series explicit. The native callback is separate
from the original byte-mask callback. Full output digest and sink retention run
after timing; no full-output hash is added to each kernel iteration.

## Fixtures, sizes and evidence

Reuse -threshold/-pattern/-seed and predicate-float32-v1, with identical input
values, parameter bits, relevant seed and x hash as the byte-mask group.
None/all/alternating/clustered/shuffled retain their original meanings.
Impossible true distributions at threshold +Inf/NaN fail before artifacts;
-pattern none permits intentional all-false comparisons.

The standard power-of-two and native-width boundaries remain. Packed output
also registers 7,9,63,65 (8 and 64 are already powers of two); both native and
compiled selectors consume the same sorted/deduplicated size list. Zero length
is covered by correctness tests, not throughput benchmarks.

Case.Predicate has EncodingID=uint8-packed-lsb0/v1 and actual process NativeLanes.
Input TypeID stays float32; the descriptor declares uint8 PackedBits output.
ByteResultSHA256 hashes the entire packed byte sequence after each reported
measurement, including its zeroed padding. Child width and eight exact uint32
output-sha256-0..7 metrics use the existing strict digest importer.
No expected digest or parent width substitutes for child facts.

Byte and packed groups have different operation/arithmetic/encoding/harness
identities and traffic models. Compare their ns/op at matching input/size/context
as separate experiments; do not pool them as interchangeable series. Smaller
output does not guarantee a faster kernel. At n=65, output is 9 bytes and logical
work 269 bytes; the byte-mask counterpart writes 65 bytes and counts 325.

## Commands and acceptance

```powershell
go run ./cmd/simdreport -list -group predicatepack
go run ./cmd/simdreport -group predicatepack '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicatePacked$/^n=65$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -backend go-test -group predicatepack '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicatePacked$/^n=(7|8|9|63|64|65)$' -count 2 -benchtime 3x -format md
go run ./examples/predicatepack
```

Quote signed decimal flags in PowerShell. Installed binaries and Make share
the same options; no additional flag, architecture gate or external SDK is added.

Independent bit-by-bit tests cover random float32 bit patterns, signed zeros,
subnormals, infinities/NaNs, exact bytes, zero padding and stale-output clearing.
Tests cover nil/empty/mismatched lengths, input preservation, output canaries,
byte/vector boundaries, current zero allocations, safe extent/overflow handling,
typed bindings and shared input identity. All five fixtures pass compiled/native
case/digest parity at byte/tail/native boundaries and exact repetitions/iterations,
including forced emulation. Partial captures, all-group/no-Go standalone runs,
pure packages, full tests/vet, Go/installed/Make commands and Windows/Linux CI
complete acceptance. Original predicate and numerical kernel bodies are retained.
