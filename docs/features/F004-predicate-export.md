# F004, slice 5 — Canonical predicate export

Historical checkpoint record. The final [F004 acceptance record](F004-completion.md)
now closes the agreed scope after clamp; counts and stop decisions below describe
this earlier slice.

Implemented 2026-11-03 after fused dot. The predicate group adds Scalar and
SIMDPartial: twelve operations/thirty-two implementations. Architecture M0–M3
remains complete; this is a bounded representation checkpoint. Packed masks,
three-array MulAdd, clamp, other input types and M4 remain separate work.

Retained [hardware capture](evidence/predicate-vector.measurements.json),
[emulated capture](evidence/predicate-software.measurements.json) and
[commands/provenance](evidence/README.md) preserve observed results. These tiny
fixed-iteration runs verify plumbing, not performance rankings.

Implementation: signed commit [8294321](https://github.com/akolybelnikov/go-simd/commit/829432189ac4cfc31c31917052a7d015ab8a2b7a).

## Experiment and exact contract

Measure the complete cost of producing a usable predicate array:

```text
out[i] = x[i] > threshold ? uint8(1) : uint8(0)
```

Operation predicate-gt-f32; input type float32; output type uint8 array;
arithmetic float32-predicate-gt-uint8-01-v1; benchmark BenchmarkPredicate.
PredicateGTScalar and PredicateGTSIMD accept
`(x []float32, out []uint8, threshold float32) error` in internal/patterns.

Comparison is ordered and strict. Equal operands, either signed zero at a zero
threshold, and unordered NaNs produce 0. Infinities/subnormals follow ordinary
Go float32 comparison. Every output byte is exactly 0 or 1; no SDK mask encoding
or NaN payload escapes into the result. Input bits are preserved.

Lengths must agree. A mismatch returns patterns.ErrLengthMismatch before writes;
equal empty/nil slices succeed. Output is caller-provided; it must occupy
disjoint storage from input. Cross-type aliases created with unsafe are outside
the contract. No output allocation or resizing occurs in the kernel.

## SIMD pipeline and inactive lanes

Scalar compares and writes each byte. SIMDPartial:

1. Broadcast threshold, one and zero once per call.
2. Load a full/partial vector and retain its actual valid-element count.
3. Compare with Greater and use IfElse to select explicit float32 1/0.
4. Convert those exact values to an int32 vector and store into scratch.
5. Narrow/write only valid entries to the uint8 output.

This deliberately avoids interpreting raw Mask32s.ToInt32s true-lane bits.
There is no invented portable Mask32s.ToBits call. The selected SDK's portable
Greater/IfElse/ConvertToInt32/Store/LoadFloat32sPart API was inspected locally.
This is byte-per-predicate materialization, not packed bits, compaction, guarded
memory access or an isolated compare-instruction benchmark.

Padded zero input lanes may compare true under a negative threshold; their
results never leave scratch. Canaries and one-valid-element tests cover this.
Current widths use a sixteen-int32 stack scratch and allocate zero per call;
a future wider vector may allocate dynamic scratch, included in counters.

## Typed preparation, timing and work

Add a separate BindUint8 callback and Prepared.ByteOutput path. Inputs are
generated/fingerprinted and the real byte output allocated before timing.
No dummy float32 output or conversion wrapper is supplied to old bindings.
Original kernel bodies and old native benchmark callbacks are unchanged.

Each timed call includes validation, broadcasts, comparison, selection,
conversion, scratch stores, extraction/narrowing and final byte writes.
Compiled, native and quick paths retain the output after stopping timing.
Logical work is `5*n` bytes per call: one float32 read plus one uint8 write.
Scratch/intermediate traffic is not part of this logical model; it is not
measured DRAM bandwidth. Zero-length correctness is tested, not benchmarked.

Harnesses are standalone-predicate-export/v1 and
legacy-report-predicate-export/v1. The compiled adapter preserves exact T/N,
iteration/repetition counts, failure propagation and prior successful samples.
Array/reduction callbacks retain their existing measurement boundaries.

## Fixture, representation and evidence identity

Shared -threshold, -pattern and -seed controls need no new CLI flag. Reuse the
existing predicate-float32-v1 recipes: none, all, alternating, clustered and
shuffled. Alternating/clustered have ceil(n/2) true entries; shuffled is
deterministic and approximately half true. Only shuffled uses seed.
Threshold bits, relevant generator identity and x's fingerprint survive;
unrelated scalar/fixture controls do not alter this input.

Impossible requested true distributions (threshold +Inf or NaN) fail before
artifacts; -pattern none permits these intentional all-false cases.

Case.Predicate records EncodingID=uint8-zero-one/v1 and the actual process
NativeLanes for both strategies. Different encodings/widths cannot pool.
Case.TypeID continues to identify the float32 input, while the descriptor
declares the uint8 output and the arithmetic/encoding identity fixes 0/1.

Observation.ByteResultSHA256 retains SHA-256 over the entire completed byte
array in index order, computed once per reported measurement after timing.
It is output evidence, not a replacement for x's input hash, raw output data or
an averaged timing statistic. Older unrelated v1 evidence omits these fields.
The CLI saves measurements and summaries; calling the kernel gives access to
the actual byte array. It does not add an output-mask-file command.

Native reports include eight exact uint32 output-sha256-0..7 metrics in
big-endian digest-word order, plus exact positive integer simd-lanes. Import
rejects missing, duplicate, fractional, negative, non-finite or oversized digest
words and invalid widths. Digests and width come from the child; expected output
and parent facts never substitute. Native/compiled harnesses remain separate.

## Commands and acceptance

```powershell
go run ./cmd/simdreport -list -group predicate
go run ./cmd/simdreport -group predicate '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicate$/^n=64$' -count 3 -benchtime 100ms -format md
go run ./cmd/simdreport -backend go-test -group predicate '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicate$/^n=64$' -count 2 -benchtime 3x -format md
go run ./examples/predicate
```

Tests compare exact canonical bytes with independent scalar predicates over
random raw float32 bits, signed zeros, infinities, subnormals and quiet/signaling
NaNs. They cover length errors before writes, input preservation, output
canaries, native-width/tail boundaries and current zero allocations. Tests also
verify typed bindings, all five fixture distributions, relevant/irrelevant
identity, raw digest strictness, mixed-context rejection and failure retention.
Compiled/native parity covers all five recipes at sizes 1,w-1,w,w+1 with two
repetitions and exactly three iterations. All-group/no-Go standalone execution,
forced emulation, pure packages, full tests/vet, Go/installed/Make commands and
Windows/Linux CI accompany the checkpoint. Stop after predicate export.
