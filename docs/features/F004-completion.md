# F004 completion — Agreed experimental scope

Current catalogue: **23 operations / 58 implementations**, after eight separately
delivered extensions. The original 15/41 acceptance audit below remains historical;
only unimplemented operations remain in the local future F004 tracker.

| Group | Implementations | Contract |
| --- | --- | --- |
| `muladd` | 2 | [Three-array fused multiply-add](F004-muladd.md) |
| `addition64` | 2 | [Float64 addition](F004-addition64.md) |
| `saturatingadd8` | 2 | [Uint8 saturating addition](F004-saturating-add8.md) |
| `bitroundtrip` | 2 | [Exact bit reinterpretation](F004-bit-roundtrip.md) |
| `xorbits` | 2 | [Opaque Float32 word XOR](F004-xor-bits.md) |
| `sum8` | 2 | [Fixed-eight-input horizontal sum](F004-sum8.md) |
| `clmul64` | 2 | [Low64 carryless polynomial multiplication](F004-clmul64.md) |
| `copy64` | 3 | [Exact Float64 copy](F004-copy64.md) |

## Original 15/41 acceptance checkpoint

**Complete for the agreed current scope, 2026-11-04.** Final slice 8 is ordered
clamp. The catalogue at this checkpoint contained **15 operations / 41 implementations**. No further
slice is required to close this checkpoint. Complete SDK/ISA coverage remains
a future vision, with every selected baseline declaration accounted for in the
[API coverage ledger](F004-api-coverage.md).

Final implementation: signed [9951b81](https://github.com/akolybelnikov/go-simd/commit/9951b81b167af193589b1b834a0b84297c18af3f). Local full tests/vet, pure packages,
SDK-emulated clamp/parity, native smoke, no-Go standalone and installed/Go/Make
checks passed before publication. CI repeats the platform/race/emulation checks.

## Acceptance audit

| Requirement | Delivered evidence / boundary | Result |
| --- | --- | --- |
| Describe useful SIMD behaviors beyond arithmetic | Broadcast constants, ordered masks/selection/composition, scalar reductions, accumulator counts, fused arithmetic, canonical/packed exports, movement controls and clamp; per-slice learning questions | Complete |
| Shared operation/strategy metadata | internal/catalog and typed kernel bindings; CLI/group/native registrations and inventory tests | Complete |
| Arity, types, parameters and output representation | Unary/binary float32 arrays, exact scalar parameter bits, float32 scalar results, uint8 canonical/packed arrays; physical output extents | Complete |
| Numerical contracts and scalar controls | Per-slice independent bit/order/FMA references; explicit NaN/zero/fusion/fold/alias rules and failure precedence | Complete |
| Tails and distributions | Empty correctness, actual native-width boundaries, nonmultiples, predicate/packing/clip boundaries, deterministic seeded fixtures and canaries | Complete |
| Matched timed work | Per-call broadcasts/validation, reduction folds, conversion/packing, full movement and clamp repairs; logical byte accounting | Complete |
| Retain individual measurements and context | Compiled exact-duration observations, native raw samples, generator/parameter fingerprints, full declared output digests/result bits, actual lanes/fusion paths | Complete |
| Honest comparisons | Gains and regressions, medians/ratios, no timing pass threshold; incompatible numerical/width/representation contexts rejected | Complete |
| FMA operand requirements | Affine constants and dot accumulators provide three vector operands; separately identified fused arithmetic and valid binary32 software/reference gate | Complete |
| User-facing consistency | Both installed CLIs, go run and Make use the shared catalogue/flags; README, full registry, examples and progressive Wiki match shipped operations | Complete |
| Execution verification | Full tests/vet; pure metadata/result packages; compiled/native parity; standalone all-group execution without Go/source; SDK-emulation checks; Linux/Windows CI | Complete |
| API/type completeness accounting | Selected Go 1.27.3 simd declarations, per-source hashes and explicit exercised/helper/deferred status; no invented portable methods | Complete ledger; exhaustive kernels deferred |
| Input files / automatic type mapping | CLI generated fixtures meet current contracts; external dataset import/type inference is conditional F002 work | Separate scope |
| Visualization and broader frontends | Exact-size comparisons and textual/Markdown/CSV output shipped; interactive/HTML charts belong to F008 | Separate scope |
| Architecture migration | M0-M3 completed before expansion; M4 isolation/resources remain independent | Separate scope |

## Delivered inventory

| Group | Implementations | Contract record |
| --- | --- | --- |
| addition | 4 | Existing partial/tail/parallel controls; internal/addition |
| computation | 2 | Existing unfused expression; internal/computation |
| scale, affine, threshold, selection, rangefilter | 10 | [Broadcast/predicates](F004-broadcast-predicates.md) |
| sum | 4 | [Sum](F004-sum.md), [accumulators](F004-accumulators.md) |
| dot | 4 | [Dot](F004-dot.md), [accumulators](F004-accumulators.md) |
| affinefma | 2 | [Fused affine](F004-fma.md) |
| dotfma | 4 | [Fused dot](F004-fused-dot.md) |
| predicate | 2 | [Canonical byte export](F004-predicate-export.md) |
| predicatepack | 2 | [Packed export](F004-packed-predicate.md) |
| copy | 4 | [Copy controls](F004-copy.md) |
| clamp | 3 | [Ordered clamp](F004-clamp.md) |

Earlier checkpoint records preserve the counts and stop decisions at delivery.
The reusable checklist in the [plan](F004-operation-patterns.md) remains a template
for new experiments, rather than unfinished acceptance work in this checkpoint.
Retained tiny fixed-iteration [captures](evidence/README.md) verify evidence
plumbing; they establish no cross-machine speedup or accuracy ranking.

## Optional work after this stopping point (historical)

Three-array element-wise MulAdd needs its own arity/16*n traffic contract.
Float64/integer inputs need typed fixture/binding/result policies. Standalone
conversion/bit-reinterpretation/saturation studies need separate numerical
questions. Fixed-width permutations and horizontal primitives need explicit
architecture/width/capability gates. None is advertised by -list today.
Imported datasets, M4 worker isolation/resource accounting and F008 HTML are
separate features. Choose the next one explicitly; this closure starts none.

The reading credit remains [Jesús Espino, SIMD and Go](https://internals-for-interns.com/posts/go-simd/),
which informed the behavioral gap review. SDK availability and implemented
contracts were checked against the selected local compiler, independently of
the article's fixed-width examples.
