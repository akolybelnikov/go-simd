# 13 — Broadcast and predicate experiments

F004's first slice adds five operations with matching Scalar/SIMDPartial controls.
The lab now has seven operations and sixteen implementations. Inspect the short
catalogue with either CLI's `-list`; add `-group` for details and current values.

## Choose an experiment

| Group | Computation | Useful question |
| --- | --- | --- |
| scale | `alpha*x` | What does broadcasting one constant cost per whole kernel call? |
| affine | `float32(alpha*x)+beta` | Does extra unfused arithmetic change the comparison? |
| threshold | `x > threshold ? x : +0` | How do mask selection and scalar branches respond to input arrangement? |
| selection | `x > threshold ? x : y` | How does selecting between two arrays differ from masking to zero? |
| rangefilter | `lower < x && x < upper ? x : +0` | What happens when two comparison masks are combined with And? |

Every new operation has one serial scalar and one portable partial-vector
implementation. There is no FMA, reduction, packed-mask output or ISA-only kernel
in this slice. All return float32 arrays; most read one input array, selection
reads two logical inputs. Logical accounting is `8*n` or `12*n`, respectively.

## Run a bounded comparison

These examples use installed commands. For a checkout, replace `simdreport`
with `go run ./cmd/simdreport` and enable GOEXPERIMENT=simd; flags stay identical.

```powershell
simdreport -list -group affine
simdreport -group scale -alpha 2 -bench '^BenchmarkScale$/^n=1024$' -count 3 -benchtime 100ms -format md
simdreport -group affine -alpha 2 -beta 1 -bench '^BenchmarkAffine$/^n=1024$' -count 3 -benchtime 100ms -format md
simdreport -group threshold -threshold 0 -pattern alternating -bench '^BenchmarkThreshold$/^n=1024$' -count 3 -benchtime 100ms -format md
simdreport -group selection -threshold 0 -pattern shuffled -seed 123 -bench '^BenchmarkSelection$/^n=1024$' -count 3 -benchtime 100ms -format md
simdreport -group rangefilter -lower=-1 -upper=1 -pattern clustered -bench '^BenchmarkRange$/^n=1024$' -count 3 -benchtime 100ms -format md
```

Make equivalents preserve the same flags:

```bash
make list ARGS="-group affine -alpha 2 -beta 1"
make report COUNT=3 BENCHTIME=100ms ARGS="-group selection -pattern shuffled -seed 123 -bench BenchmarkSelection/n=1024 -format md"
make run ARGS="-group threshold -threshold 0 -pattern all"
```

The quick CLI remains fixed at 1,048,576 elements. Report JSON evidence preserves
individual measurements and parameter/fixture identity; Markdown is the summary.
New parameterized cases have their own harness IDs. Keep compiled/native and
legacy series separate when comparing measurements.

## Control predicate arrangements

| `-pattern` | Truth arrangement |
| --- | --- |
| none | 0% true |
| all | 100% true |
| alternating | Even indices true, ceil(n/2) total |
| clustered | First ceil(n/2) true, remaining false |
| shuffled | Deterministic seed/index mixing, approximately 50% true |

Defaults are alpha=1.5, beta=-0.5, threshold=0, lower=-1, upper=1,
pattern=alternating, seed=42. Only shuffled uses seed. Parameters apply only to
operations declaring them; pattern/seed do not alter scale/affine fixtures.

Predicate x uses a false value at the comparison bound and a true value at
bound+1 when representable/satisfying, otherwise the next float32 toward +Inf.
Selection y is a separate deterministic signed array. This controls truth
distribution instead of changing only a threshold and hoping for a percentage.
Ordering and seed/fingerprints are retained in evidence. Shuffled is approximate,
not a promise of exactly half true at each small size.

To compare arrangements, run separate commands with the same group/size/count
and change only pattern. Retain both captures. Branch predictability and input
values can affect results; one timing does not prove a universal SIMD advantage.

If a requested true pattern is impossible, preparation fails before artifacts.
For example, no float32 exceeds +Inf and an empty strict interval contains no
values. Use `-pattern none` to deliberately study those all-false cases:

```powershell
simdreport -group threshold -threshold=+Inf -pattern none -bench '^BenchmarkThreshold$/^n=1024$' -count 3 -benchtime 100ms
```

## Read the numerical contract

Equality and unordered NaN comparisons take the false branch. Masked false
outputs are positive zero. Selected operands preserve their bits, including
negative zero and NaN payloads in y. A negative zero selected by a negative
threshold stays negative zero. Selection is not lazy evaluation, compaction or
memory-access protection.

Affine rounds the float32 product before adding beta; it does not use FMA.
Arithmetic NaNs remain NaN, without a fixed payload promise. Overflow/Inf are
valid arithmetic results. CLI parameters parse to float32 and evidence stores
their exact bits, preserving signed-zero and programmatic NaN-payload distinctions.
The selected parameter identity appears under each observation's Case.Parameters;
config.input contains uint32 bits, not JSON NaN/Infinity numbers.

SIMD broadcasts happen once per measured kernel call. Scalar extraction/binding
happens outside timing. There are no pre-broadcast vector variants or primitive
instruction-latency claims. Both exact in-place aliases and disjoint output work;
partial overlap is unsupported. Zero-length calls succeed and invalid lengths
preserve output. The two-array selection traffic figure is logical accounting,
not a guarantee that a scalar branch physically loads both values.

## Small runnable example

From the checkout:

```powershell
go run ./examples/patterns
```

For x=`[-2,-1,0,1,2]` and y=`[10,20,30,40,50]`, the example prints:

```text
scale: [-4 -2 0 2 4]
affine: [-3 -1 1 3 5]
threshold: [0 0 0 1 2]
selection: [10 20 30 1 2]
rangefilter: [0 -1 0 1 0]
```

The [kernel source](https://github.com/akolybelnikov/go-simd/blob/main/internal/patterns/float32.go)
and [detailed F004 contract](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-broadcast-predicates.md)
describe test coverage, numerical rules, fixture identity and native registration.
For recorded custom parameters using native benchmarks, add `-backend go-test`
to the report command; it still requires Go/source. Direct `go test` native
benchmarks use default parameters and cannot establish historical custom values
without recorded context.

This expansion was prompted by [Jesús Espino's SIMD and Go](https://internals-for-interns.com/posts/go-simd/).
See the [credited pattern review](https://github.com/akolybelnikov/go-simd/wiki/12-SIMD-patterns-and-reading)
for the remaining reduction, FMA, movement and type backlog.
