# 20 — Export predicates as packed bits

Current catalogue: **23 operations / 58 implementations**. Counts and stop
statements below describe this chapter’s historical delivery checkpoint. See
[the current contract index](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-completion.md).

The predicatepack group stores eight truth values per byte. It offers Scalar
and SIMDPartial, bringing that checkpoint to thirteen operations/thirty-four
implementations. [Copy](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement) brought it to fourteen/thirty-eight; [clamp](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts) completes the current scope at fifteen/forty-one. This extends the [canonical byte-mask experiment](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks)
with a compact representation and measured packing cost.

## Run a bounded experiment

```powershell
simdreport -list -group predicatepack
simdreport -group predicatepack '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicatePacked$/^n=64$' -count 3 -benchtime 100ms -format md
simdreport -group predicate '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicate$/^n=64$' -count 3 -benchtime 100ms -format md
simdbench -group predicatepack -pattern alternating
```

Use the same registered size (for example n=64) in both groups when comparing
their timings. The packed group additionally registers 7,9,63,65 to exercise
byte boundaries. Byte-mask and packed series have distinct arithmetic/output
identities and logical traffic; inspect ns/op with matched inputs and runtime.

For an exact partial-byte boundary, use this direct anchored selector:

```powershell
simdreport -group predicatepack '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicatePacked$/^n=65$' -count 2 -benchtime 3x -format md
```

From a checkout with GOEXPERIMENT=simd, replace executables with
go run ./cmd/simdreport or go run ./cmd/simdbench. Make forwards the same flags:

```bash
make list ARGS="-group predicatepack"
make report COUNT=3 BENCHTIME=100ms ARGS="-group predicatepack -threshold=-1.5 -pattern shuffled -seed 123 -bench BenchmarkPredicatePacked/n=64 -format md"
make run ARGS="-group predicatepack -pattern alternating"
```

In PowerShell, quote the complete signed decimal flag as shown. Defaults remain
threshold=0, pattern=alternating, seed=42. No new option is needed. The quick
comparison still uses 1,048,576 float32 inputs.

## Decode the exact representation

For input index i, inspect bit i%8 of byte i/8. Bit zero is the least-significant
bit. True ordered x[i] > threshold sets it; false clears it. Equality, NaNs and
either signed zero at a zero threshold are false. Infinity/subnormal comparisons
follow the same rules as the byte-mask group.

| Input truth values in index order | Packed bytes, hexadecimal | Explanation |
| --- | --- | --- |
| 1,0,0,1,0,0,1,0 | 49 | Bits 0,3,6 set |
| 1,0,0,1,0,0,1,0,1 | 49 01 | Ninth input starts byte two |
| One true input | 01 | Seven unused high bits zero |
| Nine true inputs | ff 01 | Final partial byte is normalized |

Output length is exactly ceil(n/8). Unused high bits are always zero, and every
byte is overwritten on each call. This makes output reproducible even when a
previous call left true bits. The output is a packed mask with unchanged logical
input count; it does not compact input values or guard memory accesses.

Scalar builds each byte from up to eight comparisons. SIMDPartial selects
explicit float32 0/1 from vector comparisons, converts/extracts to int32 scratch,
then packs only valid lanes. Partial output bytes carry across vector iterations,
including four-lane emulation. Raw SDK mask encodings never escape, and no
portable Mask32s.ToBits method is assumed. Invalid padded lanes cannot set bits
under negative thresholds.

## Understand the comparison cost

Validation, broadcasts, comparison, normalization, conversion, extraction,
packing and stores are timed. Buffer allocation, input fingerprints, output
retention and complete output hashing are outside timing. Current widths
allocate zero per call; a future wider scratch can appear in allocation counters.

| Representation | Output bytes | Logical bytes per call |
| --- | --- | --- |
| predicate | n | 5*n |
| predicatepack | ceil(n/8) | 4*n+ceil(n/8) |

At n=65 packed output is 9 bytes and logical work 269 bytes. Byte output would
be 65 bytes and 325 logical bytes. At a multiple of eight, packed output occupies
one eighth as much space. Packing itself has a cost, so smaller output need not
win in time. Logical bytes exclude scratch traffic and are not DRAM measurements.

Both groups share -threshold/-pattern/-seed recipes and identical input values.
All/none, alternating, clustered and seeded shuffled retain their meanings.
Requested true patterns above +Inf or NaN fail before artifacts; use -pattern none
for intentional all-false cases. Evidence retains exact threshold bits, relevant
seed/generator identity and x's hash.

## Read evidence and call the kernel

Case.Predicate records EncodingID=uint8-packed-lsb0/v1 and actual NativeLanes.
ByteResultSHA256 hashes every packed byte, including normalized final padding,
once after each reported measurement. It is separate from x's input hash and
does not save the mask data itself. Module-local calls expose the actual output.

Harnesses are standalone-packed-predicate/v1 and legacy-report-packed-predicate/v1.
Use -backend go-test for a developer comparison requiring Go/source. Native
width and eight exact digest-word metrics come from the child. Byte and packed
encodings, widths and harnesses stay separate when comparing reports.

```powershell
$path = Get-ChildItem ./artifacts/predicatepack/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
$capture.run.Observations | Select-Object Label, ByteResultSHA256
$capture.run.Observations[0].Key.Case.Predicate
```

Inside this module:

```go
x := []float32{1,0,0,1,0,0,1,0,1}
out := make([]uint8, (len(x)+7)/8)
err := patterns.PredicateGTPackedSIMD(x, out, 0)
// err == nil; out == []uint8{0x49, 0x01}
```

PredicateGTPackedScalar has the same signature. A wrong output length returns
patterns.ErrLengthMismatch before writes. Empty/nil succeeds; input bits are
preserved. Input/output storage must be disjoint; unsafe cross-type aliases
are unsupported. These remain internal experimental APIs.

Run the complete [example](https://github.com/akolybelnikov/go-simd/blob/main/examples/predicatepack/main.go):

```powershell
go run ./examples/predicatepack
```

```text
byte mask (9 bytes): [1 0 0 1 0 0 1 0 1]
packed Scalar (2 bytes): 4901
packed SIMDPartial (2 bytes): 4901
one true input; unused bits zero: 00000001
```

Read the [kernels](https://github.com/akolybelnikov/go-simd/blob/main/internal/patterns/predicate_packed.go)
and [contract/evidence record](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-packed-predicate.md).
Packed export is complete. [Exact copying](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement) is now available as a
separate movement experiment. Clamp, reordering, three-array MulAdd, standalone
conversion studies, other input types and M4 remain pending.

The implementation record links the signed code and retained hardware/emulated
captures with exact commands and source provenance. Tiny fixed-iteration captures
verify plumbing without performance claims.
