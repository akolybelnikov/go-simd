# go-simd

Compare ordinary and vectorized Go operations on generated arrays. Timings
describe the operation on the machine running the lab. [Install the CLI](docs/cli.md#install-the-cli-from-source)
to get started.

## Open the local browser UI

```text
simdlab ui
```

Use **Explore** to see available operations, methods, and array sizes.
On **Compare**, choose **Run locally** for the addition starter or use a
preview from the catalog. **Results** shows the measured difference and
individual repeats, opens saved captures, and downloads Markdown. The
benchmark runs on the machine that launched `simdlab`. Comparisons work offline.
Choose **Ask an agent** in Explore or Results for questions about SIMD,
your workload, or a specific batch of results. Copy the question to an
assistant you already use, or enable in-app chat with a local Ollama model
or your own OpenAI API key for GPT-6.1 Sol. [Chat setup](docs/wiki/33-Chat-with-an-agent.md)

For local chat, start Ollama with an installed model and launch
`simdlab ui -ollama`. In **How to ask**, select that model, type a question,
and choose **Send**. You can ask follow-up questions in the same panel.
For OpenAI, set `OPENAI_API_KEY` before launching `simdlab ui` and select
**OpenAI · gpt-6.1-sol**. Review the shared context before sending.

Use the CLI for scripts, automation, and SSH. The public site source provides
downloads, documentation, and a sample report. See the [browser guide](docs/local-ui.md)
for the local app and saved reports.

## Run directly

```text
simdlab run addition
```

This compares ordinary and vectorized addition on two generated arrays of
1,024 float32 values, with three repeats per method. The terminal displays
the settings, runs the comparison, and prints the timing difference and the
paths to the Markdown report and measurement file. This path works without
a model. Press Ctrl+C to cancel.

## Guided CLI preview

The examples below work without Ollama. Other wording may use the optional local model.

```text
simdlab ask
```

Enter **Compare ordinary and vectorized addition for 1,024 numbers.** Review
the methods, batch size, and repeats, then enter **yes** to measure or **no**
to decline. After the run, ask **Which took less time, and by how much?** or
**Were the repeats consistent?** The answer uses measurements saved during
the session. Enter **quit** to leave.

You can also ask **Can I compare ways of summing an array?** to see the
methods and input sizes available on your machine. See the [CLI reference](docs/cli.md)
for Ollama setup, exact flags, and reopening saved evidence.

The lab covers 23 operations and 58 implementations. See the [operation catalogue](#operation-catalogue),
and [experiment recipes](docs/wiki/07-Experiment-recipes.md).
Inspired by [SIMD and Go](https://internals-for-interns.com/posts/go-simd/) by Jesús Espino.

## Command reference

Both commands share these options:

| Option | Default | Meaning |
| --- | --- | --- |
| `-group` | `all` | Choose a group from the catalogue below |
| `-list` | `false` | Implemented operations and runtime context; no measurements or artifacts |
| `-addition`, `-computation` | `false` | Group shortcuts |
| `-function add\|compute\|all` | Unset | Deprecated group alias retained for compatibility |
| `-h`, `-help` | — | Readable usage, options and examples |

Explicit group selections must agree. False boolean aliases do not select a
group. There is no public type/input-file selector yet.

### Operation catalogue

`-list` prints a compact table; `-list -group <name>` shows parameters, expressions
and strategy details. Existing aliases remain for addition/computation.

| Group | Operation | Implementations |
| --- | --- | --- |
| copy64 | Exact binary64 bit copy, including NaN payloads | Scalar source loop, GoCopy, SIMDPartial |
| clmul64 | Low64 polynomial product over GF(2) on true uint64 arrays | Scalar shift/XOR, SIMDPartial Even/Odd extraction |
| sum8 | One specified-order float32 result per eight inputs, padded with +0 | Scalar, SIMDPartial |
| xorbits | Exact XOR of two opaque float32 bit patterns | Scalar uint32 logic, SIMDPartial ToBits/Xor/BitsToFloat32 |
| bitroundtrip | Exact Float32/Uint32/Float32 representation round trip | Scalar bitcast source, SIMDPartial ToBits/BitsToFloat32; all bits preserved |
| saturatingadd8 | min(a+b,255) over true uint8 arrays | Scalar widened sum, SIMDPartial AddSaturated |
| addition64 | Binary64 a+b, true float64 arrays | Scalar, SIMDPartial; actual Float64 lanes |
| addition | `a+b` | Scalar, SIMDPartial, SIMDTail, SIMDParallel |
| computation | `a*a+b*b+a*b` | Scalar, SIMD |
| scale | `alpha*x` | Scalar, SIMDPartial |
| affine | `float32(alpha*x)+beta`, unfused | Scalar, SIMDPartial |
| affinefma | One-rounding `alpha*x+beta` | Scalar, SIMDPartial; vector/software path recorded |
| threshold | `x > threshold ? x : +0` | Scalar, SIMDPartial |
| selection | `x > threshold ? x : y` | Scalar, SIMDPartial |
| rangefilter | `lower < x && x < upper ? x : +0` | Scalar, SIMDPartial |
| predicate | `x > threshold ? uint8(1) : uint8(0)` | Scalar, SIMDPartial; conversion and byte stores included |
| predicatepack | Same predicate packed LSB first, eight bits per byte | Scalar, SIMDPartial; conversion, packing and stores included |
| copy | Exact float32 bit copy | Scalar source loop, GoCopy built-in control, SIMDPartial, SIMDTail |
| muladd | Three-array one-rounding a*b+c | Scalar software reference; SIMDPartial verified vector/software path |
| clamp | Clip outside inclusive bounds; preserve equality/NaN bits | Scalar, SIMDPartial selection, SIMDMinMax with timed semantic repairs |
| sum | One returned float32, specified fold orders | Scalar, SIMDStriped, SIMDStriped2, SIMDStriped4 |
| dot | Unfused float32 products and specified reduction folds | Scalar, SIMDStriped, SIMDStriped2, SIMDStriped4 |
| dotfma | FMA updates with specified merge/fold orders | Scalar, SIMDStriped, SIMDStriped2, SIMDStriped4; actual path recorded |

Both commands accept these fixture/parameter options. Each operation uses only its declared
parameters; exact float32 bits and relevant fixture identity survive in evidence.

| Option | Default | Applies to |
| --- | --- | --- |
| `-alpha` | `1.5` | Scale and affine/affinefma multiplier |
| `-beta` | `-0.5` | Affine/affinefma offset; group defines fusion |
| `-threshold` | `0` | Threshold, selection and byte/packed predicate export |
| `-lower`, `-upper` | `-1`, `1` | Exclusive rangefilter bounds; inclusive clamp bounds |
| `-pattern` | `alternating` | Predicate/clipped-position fixture: none, all, alternating, clustered, shuffled |
| `-seed` | `42` | Deterministic shuffled predicate/clipped-position arrangement |
| `-sum-input` | `balanced` | Sum/sum8 fixture: balanced, cancellation, mixed-magnitude |
| `-dot-input` | `balanced` | Dot/dotfma fixture: balanced, cancellation, mixed-magnitude, fusion-sensitive |
| `-fma-input` | `balanced` | Fused affine/muladd fixture: balanced, rounding, mixed-magnitude |

```text
go run ./cmd/simdreport -group selection -threshold 0 -pattern shuffled -seed 123 -bench "^BenchmarkSelection$/^n=1024$" -count 3 -benchtime 100ms -format md
go run ./examples/patterns
```

Original broadcast/predicate SIMD kernels broadcast once per measured call. Affine keeps separate float32 multiply
and add rounding. Predicates use strict ordered comparisons; equality/NaN is
false, false masking yields +0 and selection preserves chosen operand bits.
Impossible true distributions fail early; `-pattern none` permits deliberate
all-false bounds. See [broadcast and predicate experiments](https://github.com/akolybelnikov/go-simd/wiki/13-Broadcast-and-predicate-experiments)
for examples, numerical rules, byte accounting and optional follow-on experiments.

Predicate export creates a `[]uint8` mask with canonical 0/1 values. Strict
comparison makes equality and NaN false. SIMD includes conversion, extraction
and valid byte writes; logical work is `5*n`. Reports retain the full byte-output
digest after timing. See [predicate export](https://github.com/akolybelnikov/go-simd/wiki/19-Predicate-export-and-byte-masks).

```text
go run ./cmd/simdreport -group predicate "-threshold=-1.5" -pattern shuffled -seed 123 -bench "^BenchmarkPredicate$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/predicate
```

Packed export adds `predicatepack`: input i maps to bit i%8 of byte i/8.
Output occupies `ceil(n/8)` bytes with zeroed unused high bits. Timings include
conversion and packing; logical work is `4*n+ceil(n/8)`. Both representations
share input recipes but retain distinct output identities. Run
`go run ./examples/predicatepack`; see [packed masks](https://github.com/akolybelnikov/go-simd/wiki/20-Packed-predicate-masks).

```text
go run ./cmd/simdreport -group predicatepack "-threshold=-1.5" -pattern shuffled -seed 123 -bench "^BenchmarkPredicatePacked$/^n=65$" -count 3 -benchtime 100ms -format md
```

Copy compares a source loop, Go's built-in copy, partial vectors and full vectors
with a scalar tail. All preserve exact float32 bits, including NaN payloads and
signed zero. Measured buffers are disjoint; logical work is `8*n`. Percentages
use Scalar; inspect GoCopy as another optimized control. See [copy and movement](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement).

```text
go run ./cmd/simdreport -group copy -bench "^BenchmarkCopy$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/copy
```

Clamp compares branches, SIMD selection and Min/Max with explicit repairs.
Bounds are inclusive and ordered; NaN bounds fail before writes. Input NaNs and
signed-zero/equality ties keep their bits. `-pattern` selects clipped positions;
`-pattern none` also permits the full infinite interval. Work is `8*n`, and the
complete output digest is retained after timing. See [clamp](https://github.com/akolybelnikov/go-simd/wiki/22-Clamp-and-numerical-contracts).

```text
go run ./cmd/simdreport -group clamp "-lower=-1.5" -upper 2.5 -pattern shuffled -seed 123 -bench "^BenchmarkClamp$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/clamp
```

MulAdd streams three arrays and rounds a*b+c once. It shares -fma-input recipes,
with no scalar coefficients. Logical work is 16*n; Scalar uses corrected software
FMA32, while SIMDPartial records its verified vector/software path. See [MulAdd](https://github.com/akolybelnikov/go-simd/wiki/24-Three-array-fused-multiply-add).

```text
go run ./cmd/simdreport -group muladd -fma-input rounding -bench "^BenchmarkMulAdd$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/muladd
```

Float64 addition uses -group addition64 and true binary64 fixtures. It accounts
for 24*n logical bytes, with half as many lanes at a fixed vector width. Case.Vector
retains actual Float64 lanes. Existing addition remains float32. See [Float64 addition](https://github.com/akolybelnikov/go-simd/wiki/25-Float64-addition).

```text
go run ./cmd/simdreport -group addition64 -bench "^BenchmarkAddition64$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/addition64
```

Uint8 saturation computes min(a+b,255), with true byte inputs/output and 3*n
logical traffic. It preserves the distinction between clipping and wraparound.
See [uint8 saturation](https://github.com/akolybelnikov/go-simd/wiki/26-Uint8-saturating-addition).

```text
go run ./cmd/simdreport -group saturatingadd8 -bench "^BenchmarkSaturatingAdd8$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/saturatingadd8
```

Bitroundtrip preserves every float32 bit through uint32 reinterpretation,
including NaN payloads and signed zero. Logical work is 8*n. Compilers may remove
the bitcast steps; the experiment measures a complete kernel, not instruction
latency. See [bit representation](https://github.com/akolybelnikov/go-simd/wiki/27-Bit-reinterpretation-round-trip).

```text
go run ./cmd/simdreport -group bitroundtrip -bench "^BenchmarkBitRoundTrip$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/bitroundtrip
```

Xorbits applies word logic to two opaque float32 representations, such as a
sign-bit mask. It performs no floating arithmetic; logical traffic is 12*n and
all input/output bits survive in evidence. See [XOR bits](https://github.com/akolybelnikov/go-simd/wiki/28-XOR-of-float32-bit-patterns).

```text
go run ./cmd/simdreport -group xorbits -bench "^BenchmarkXorBits$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/xorbits
```

Sum returns a scalar through a typed reduction path. Scalar folds left to right;
SIMDStriped accumulates full vectors, folds lanes low to high, then handles the
scalar tail. Rounding can differ in either direction. Logical bytes are `4*n+4`;
individual results and fold identities survive in evidence. See the
[sum and reduction-order guide](https://github.com/akolybelnikov/go-simd/wiki/14-Sum-and-reduction-order).

```text
go run ./cmd/simdreport -group sum -sum-input cancellation -bench "^BenchmarkSum$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/sum
```

Dot multiplies two inputs, rounds each product before addition, and returns
`(float32,error)`. Scalar/SIMDStriped follow specified folds without FMA; unequal
lengths return `reduction.ErrLengthMismatch`. Logical work is `8*n+4` bytes.
The [dot guide](https://github.com/akolybelnikov/go-simd/wiki/15-Dot-product-and-unfused-rounding)
explains fixtures, unfused rounding and result evidence.

```text
go run ./cmd/simdreport -group dot -dot-input cancellation -bench "^BenchmarkDot$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/dot
```

Sum and dot also compare two/four independent vector accumulators. Full vectors
are assigned round-robin, accumulators merge lane-wise left to right, then lanes
and scalar tail are folded. This changes rounding as well as loop structure.
Existing report `-bench` selectors choose the variants; no new flag is needed.
Run `go run ./examples/accumulators` and see the
[accumulator guide](https://github.com/akolybelnikov/go-simd/wiki/16-Accumulator-count-experiments)
for exact order, result evidence and numerical examples.

Fused affine uses a correctly rounded software reference and a gated vector
MulAdd path, with an explicit software path for emulation or a failed gate.
Listing, quick output and evidence expose the actual path. The startup probe
runs outside measurement; existing affine/dot remain unfused. See the
[FMA guide](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding).

```text
go run ./cmd/simdreport -group affinefma -fma-input rounding "-alpha=1.0000001192092896" "-beta=-1" -bench "^BenchmarkAffineFMA$/^n=8$" -count 3 -benchtime 100ms -format md
go run ./examples/fma
```

Fused dot adds the separate `dotfma` group: Scalar and 1/2/4-vector accumulator
strategies. Each multiply-accumulate rounds once to float32; accumulator merges
and lane folds use ordinary rounded additions. Software execution preserves the
same stripes. Evidence records the actual path for full vectors, scalar tails
and emulation alongside fold order and scalar result bits. See the
[fused-dot guide](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders).

```text
go run ./cmd/simdreport -group dotfma -dot-input fusion-sensitive -bench "^BenchmarkDotFMA$/^n=64$" -count 3 -benchtime 100ms -format md
go run ./examples/dotfma
```

`simdbench` measures a fixed 1,048,576-element input and prints a quick table.
It has no report/count/timing flags. `simdreport` adds:

| Option | Default | Meaning |
| --- | --- | --- |
| `-bench` | `.` | Go-style regex over benchmark / size / variant |
| `-count` | `5` | Independent measurements per selected case |
| `-benchtime` | `100ms` | Positive duration or fixed iteration count, e.g. `10x` |
| `-format` | `txt` | Saved summary: `txt`, `md`, `csv`; terminal stays text |
| `-output` | `artifacts` | Output root, relative to the current directory unless absolute |
| `-backend` | `compiled` | Standalone backend; `go-test` explicitly selects the developer backend |

Names are case-sensitive. For example,
`-bench '^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial)$'` selects two cases.
Regex selects existing sizes, not arbitrary new sizes. With `-group all`, it
must match cases in every selected group. Invalid timing, regex, format, backend or
selectors fail before artifacts are created.

## Build or install

Set `GOEXPERIMENT=simd` as shown in the [installation guide](docs/cli.md#install-the-cli-from-source).
GNU Make plus a POSIX shell is optional. A compatible target binary runs without
Go or the `GOEXPERIMENT` environment setting at runtime.

```text
go build -o ./bin/ ./cmd/simdbench ./cmd/simdreport ./cmd/simdlab
```

Make equivalents: `make build` and `make install`. Installation uses `GOBIN`,
or the Go workspace's `bin` directory when unset; add it to PATH. Windows builds
produce `bin/simdbench.exe`, `bin/simdreport.exe` and `bin/simdlab.exe`; POSIX
builds omit `.exe`.
Copy compatible binaries to a server or invoke them by absolute path. A prebuilt
release download/package-manager installer is not supplied yet.

See [installation and remote use](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry#build-install-and-run-remotely)
for cross-compilation, copying binaries and SSH examples. Windows/amd64 and
Linux/amd64 execution are covered by CI. Run `-list` on the target to inspect
its vector width, emulation state and GOMAXPROCS.

## Reports and evidence

Every completed group saves a summary and matching evidence:

- Default `compiled`: `.measurements.json`, schema `go-simd.measurements/v1`,
  retains exact duration/iterations, observations, case/input identities,
  allocation counters and runtime/build facts.
- Explicit `-backend go-test`: legacy `.raw` text. This developer option needs
  Go on PATH and the repository root. There is no automatic fallback.

TXT/Markdown include environment metadata below the table. CSV contains numeric
rows; metadata stays in the companion. Negative change means less time than
scalar. Missing baselines stay unavailable. Throughput uses logical work bytes,
not measured DRAM traffic. CPU model detection is currently `unknown`; source
revision/modified state are recorded when embedded by the Go build.

Keep backend/harness identities separate when comparing results. Failed or empty
compiled measurements retain a failed capture and no successful group summary.
Abrupt termination recovery and worker manifests remain later work.

## Develop and learn

```text
go test ./... -count=1
go vet ./...
gofmt -l cmd internal examples
go run ./examples/basic
```

Make equivalents include `make check`, `make test`, `make vet`, `make fmt` and
`make fmt-check`. `make clean` removes generated reports and Go build artifacts.
The [CLI reference](docs/cli.md) covers common command options and comparison
recipes. Run `make help` for the Make targets.

The [Wiki](https://github.com/akolybelnikov/go-simd/wiki) covers benchmark
methods, experiment recipes, numerical behavior, and troubleshooting.
See the [changelog](CHANGELOG.md) for release changes.

Shared boundaries live in `internal/cli`, `internal/engine`, `internal/platform`,
`internal/catalog`, `internal/input`, `internal/evidence`, `internal/result` and `internal/report`.
Kernels remain in [addition](internal/addition/README.md) and
[computation](internal/computation/README.md), with new kernels in
[patterns](internal/patterns/README.md), [reductions](internal/reduction/README.md)
and [movement](internal/movement/README.md).
Elementwise Float32/Float64/Uint8/Uint64 array-output kernels support disjoint output and exact in-place aliases;
partial overlap is unsupported. Sum8 supports its same-start projected output
prefix, with exactly ceil(n/8) values. Predicate export requires disjoint float32 input
and uint8 output storage. Length errors leave output unchanged. These internal packages are not a stable public API.


| Focused example | Guide |
| --- | --- |
| `go run ./examples/sum8` | [Fixed-block horizontal sum](https://github.com/akolybelnikov/go-simd/wiki/29-Fixed-block-horizontal-sum) |
| `go run ./examples/clmul64` | [Carryless polynomial multiplication](https://github.com/akolybelnikov/go-simd/wiki/30-Carryless-polynomial-multiplication) |
| `go run ./examples/copy64` | [Exact Float64 copy](https://github.com/akolybelnikov/go-simd/wiki/31-Exact-Float64-copy) |
