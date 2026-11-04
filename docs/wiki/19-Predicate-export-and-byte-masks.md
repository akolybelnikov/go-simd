# 19 — Export predicates as byte masks

Checkpoint history: [three-array MulAdd](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add) is now implemented after the
original F004 closure. The current catalogue has twenty-one operations/fifty-one
implementations; future/stop statements below describe this earlier checkpoint.

The predicate group creates a usable uint8 array: one byte per input element,
exactly 1 when x > threshold and 0 otherwise. It adds Scalar and SIMDPartial,
bringing that checkpoint to twelve operations/thirty-two implementations.
[Packed predicates](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks) brought it to thirteen/thirty-four; [copy](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement) brought it to fourteen/thirty-eight; [clamp](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts) completes the current scope at fifteen/forty-one.

This answers: how much does comparison plus producing a usable mask cost?
The existing threshold group keeps numerical values or writes float32 zero;
predicate exports the truth values themselves.

## Run the same experiment through any launcher

Installed commands:

```powershell
simdreport -list -group predicate
simdreport -group predicate '-threshold=-1.5' -pattern shuffled -seed 123 -bench '^BenchmarkPredicate$/^n=64$' -count 3 -benchtime 100ms -format md
simdbench -group predicate -pattern alternating
```

From a checkout with GOEXPERIMENT=simd, use go run ./cmd/simdreport or
go run ./cmd/simdbench. Make forwards the same options:

```bash
make list ARGS="-group predicate"
make report COUNT=3 BENCHTIME=100ms ARGS="-group predicate -threshold=-1.5 -pattern shuffled -seed 123 -bench BenchmarkPredicate/n=64 -format md"
make run ARGS="-group predicate -pattern alternating"
```

No new flag is needed. -threshold defaults to 0; -pattern to alternating;
-seed to 42. Quick comparison uses 1,048,576 elements. Report regex chooses
registered sizes and -count/-benchtime control repeated measurement.
Quote the complete negative-decimal flag in PowerShell as shown, so its parser
preserves the argument. The Make example passes it inside quoted ARGS.

## Read the output contract

| Input compared with threshold 0 | Output byte |
| --- | --- |
| Positive finite value or +Inf | 1 |
| Negative finite value or -Inf | 0 |
| +0 or -0 | 0 |
| NaN | 0 |

Equality and unordered comparisons are false. NaN threshold makes every
comparison false. The output contains canonical bytes, not packed bits or
raw vector-mask storage. It has the same element count as x; there is no
compaction, guarded access or selected-value output.

Scalar compares/writes bytes directly. SIMDPartial broadcasts threshold/one/zero
once per call, compares full/partial vectors, selects explicit float32 0/1,
converts to int32 scratch and writes only valid bytes. This avoids depending on
the SDK's raw true-lane encoding. A negative threshold can make padded zero lanes
true, but those inactive results never reach the output.

Timing includes validation, broadcasts, comparison, normalization, conversion,
scratch stores, extraction/narrowing and byte writes. Buffers/fixtures/binding,
fingerprinting, output retention and hashing are outside timing. This measures
complete kernels; it does not isolate instruction latency or promise SIMD wins.

Logical work is 5*n bytes: four bytes read and one written per element.
This excludes intermediate scratch traffic and is not measured DRAM bandwidth.
Current widths allocate zero per call; future wider scratch may allocate.

## Select a predicate distribution

| -pattern | Generated truth arrangement |
| --- | --- |
| none | All false |
| all | All true |
| alternating | Even indices true, ceil(n/2) total |
| clustered | First ceil(n/2) true |
| shuffled | Deterministic seed/index mixing, approximately half true |

The same recipes serve threshold/selection and predicate export. Only shuffled
uses seed. Each case retains threshold's exact float32 bits, relevant generator
identity and x's input hash. Alpha/beta, bounds and sum/dot/FMA fixtures do not
alter predicate input. Requesting true values above +Inf or NaN fails before
artifacts; use -pattern none to study those all-false comparisons.

```powershell
simdreport -group predicate '-threshold=+Inf' -pattern none -bench '^BenchmarkPredicate$/^n=64$' -count 2 -benchtime 3x -format md
```

## Inspect evidence without confusing input and output

Case.Predicate records EncodingID=uint8-zero-one/v1 and NativeLanes for the
measurement process. The input TypeID stays float32; the catalogue declares
uint8 output. Different widths/encodings cannot pool into one comparison.

ByteResultSHA256 is SHA-256 over the complete byte output in index order,
computed after timing for each observation. It is separate from Case.Input's
hash of x. It is neither the full mask data nor a timing average. The report CLI
exports summaries/evidence; module-local kernel calls expose the actual array.

```powershell
$path = Get-ChildItem ./artifacts/predicate/*.measurements.json | Sort-Object LastWriteTime -Descending | Select-Object -First 1
$capture = Get-Content $path.FullName -Raw | ConvertFrom-Json
foreach ($row in $capture.run.Observations) {
    $row.Label
    $row.Key.Case.Predicate
    $row.ByteResultSHA256
}
```

Harnesses are standalone-predicate-export/v1 and
legacy-report-predicate-export/v1. Add -backend go-test for a developer comparison
requiring Go/source. It retains strict child simd-lanes and eight digest-word
metrics, encoded as exact uint32 values; parent capabilities cannot replace
child facts. Native/compiled series remain distinct despite equal numerical
identity and output hashes.

## Call the kernels inside this module

```go
x := []float32{-2, 0, 1, 2}
out := make([]uint8, len(x))
err := patterns.PredicateGTSIMD(x, out, 0)
// err == nil; out == []uint8{0, 0, 1, 1}
```

PredicateGTScalar has the same signature. Unequal lengths return
patterns.ErrLengthMismatch before output writes. Empty/nil equal slices succeed;
input is unchanged. Input/output storage must be disjoint; cross-type aliases
created with unsafe are unsupported. These internal packages are module-local
experimental APIs, not a stable public import library.

Run the [complete example](https://github.com/akolybelnikov/go-simd/blob/main/examples/predicate/main.go):

```powershell
go run ./examples/predicate
```

```text
x > 0, Scalar:     [0 0 0 1 1 0 1]
x > 0, SIMDPartial: [0 0 0 1 1 0 1]
one valid false lane at threshold -1: [0]
```

Read the [kernels](https://github.com/akolybelnikov/go-simd/blob/main/internal/patterns/predicate.go)
and [implementation/evidence record](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-predicate-export.md).
Tests check exact bytes, NaN/equality/signed-zero rules, valid tail writes,
all distributions and backend evidence parity, including emulation.
The byte-mask contract remains as described. [Packed masks](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks) are now
available as a separate group with explicit bit order and packing costs. Clamp,
movement, three-array MulAdd, other input types and M4 remain separate work.

The implementation record links the signed code and retained hardware/emulated
captures with reproduction commands and source provenance. Tiny fixed-iteration
captures verify plumbing without performance claims.
