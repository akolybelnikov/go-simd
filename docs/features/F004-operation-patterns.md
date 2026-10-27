# F004 — SIMD operation and experiment-pattern catalogue

Status: **extended requirements and implementation backlog**, 2026-10-27.
No kernels, catalogue registration, CLI flags, or measurement behavior are added
by this review. Architecture M0-M3 are now complete; new kernels remain separate.
M2 supplied the shared catalogue for existing kernels, descriptor traffic and
`simdbench -list`. The new patterns below remain an implementation backlog.
M3 supplies the standalone compiled report engine and shared CLI listing/aliases;
it does not implement these proposed operation patterns.

## Purpose and evidence

Grow the lab through experiments that demonstrate different SIMD behavior,
rather than implementing a list of arithmetic methods before addressing data
movement, constants, predicates, or accumulation. Keep an API/type coverage
ledger; exhaustive portable/ISA coverage is not an early release criterion.

The reading prompt was [Jesús Espino, *SIMD and Go*, Internals for Interns,
2026-10-19](https://internals-for-interns.com/posts/go-simd/). Credit also appears
in the user Wiki. Its Go 1.27.0, primarily amd64/`archsimd` examples are context;
the availability decisions below were independently checked against this
checkout and the installed Go 1.27.3 SDK with `GOEXPERIMENT=simd`.

Implementation evidence: `internal/addition/add.go`,
`internal/computation/compute.go`, the benchmark registrations, and the M1 result
model. API evidence: local `go doc simd.Float32s`, `simd.Mask32s`, individual
methods below, and SDK `src/simd/{simd_stubs.go,simd_emulated.go,midway_common.go}`.
The SDK inventory is not a list of implemented lab commands.

## Current coverage

Six callable kernels implement only two array-to-array operations: addition and
`a*a + b*b + a*b`. Existing addition variants compare partial vectors, full
vectors/scalar tails, and per-call parallel work. The expression composes
element-wise `Mul` and `Add`; it is neither a reduction nor an FMA benchmark.
Loads, stores, and tails are exercised inside those kernels, not isolated as
primitive latency experiments. There are no broadcast parameters, masks,
selection kernels, accumulator loops, scalar-output reductions, or permutations.

## Pattern gap ledger

All proposed operation IDs below are design candidates, not registered IDs or
new CLI selectors. M2 preserves existing M0 IDs in the shared runtime catalogue.

| Pattern | Present in the lab | Verified API / boundary | Proposed useful experiment |
| --- | --- | --- | --- |
| Contiguous load/operate/store | Yes, both operations | Portable full/partial loads and stores | Retain current arithmetic controls; add copy only to answer a movement question |
| Broadcast scalar parameters | No | `simd.BroadcastFloat32s` | `scale-f32`: `out[i]=alpha*x[i]`; `affine-f32`: `alpha*x[i]+beta` |
| Comparisons / predicate output | No | `Float32s.Greater`, related comparisons; `Mask32s` | `predicate-gt-f32`: export a defined logical mask, with representation conversion cost declared |
| Mask composition | No | `Mask32s.And`, `Or`; no portable mask `ToBits` in this SDK | Combine lower/upper predicates; test none/some/all true and valid tail lanes |
| Masking / selection | No | `Float32s.Masked`, `IfElse` | `threshold-zero-f32` and `select-gt-f32` with identical scalar result contracts |
| Min/max and clamp | No | Portable `Min`, `Max` and comparison/selection | `clamp-f32`; compare implementations only after NaN/signed-zero/tie policies agree |
| Explicit multiply-add | No; ordinary `Mul`/`Add` composition exists | Portable `MulAdd(y,z)`; semantics need a reference gate | `muladd-f32`, with separately identified fused/unfused modes |
| Streaming accumulation / reduction | No | No portable `Sum` or horizontal-pair method in the inspected inventory | `sum-f32`, using a declared fold; portable vector-store/final scalar fold is a possible composition |
| Dot product / accumulator count | No | Composed multiply/accumulate/fold, with FMA conditional on its semantics gate | `dot-f32`; compare one versus multiple independent accumulators as separate strategies |
| Permutation / shuffle | No | `archsimd.Float32x8.Permute`; absent on portable `Float32s` here | Explicitly gated fixed-block lane reorder; not silently a whole-slice reverse |
| Horizontal fold / lane extraction | No | `Float32x4.ConcatAddPairs` and `GetElem`; `Float32x8.GetLo/GetHi` in amd64 `archsimd` | Separate register-fold study and end-to-end reduction, including extraction cost |
| Numeric conversion / reinterpretation | No | Portable `ConvertToInt32`, `ToBits`, integer `BitsToFloat32` | Numeric conversion versus exact bit round-trip as different operations |
| Saturation / integer logic / shifts | No | Type-dependent integer API; not available uniformly | A bounded clipping/overflow or bit-manipulation experiment when justified |

Packed-mask export is another distinct representation: amd64
`archsimd.Mask32x8.ToBits` exists, while portable `Mask32s.ToInt32s` exports an
integer-vector representation. Do not invent a portable `Mask32s.ToBits`, assume
raw true-lane encoding, or compare packed bits with scalar bool arrays without
matching both output and conversion boundaries.

## Delivery order and architecture dependencies

1. **M2 foundation (complete):** register the current six kernels through shared typed
   descriptors. Descriptors must be able to describe array operands, scalar
   parameters, output kind, semantics, requirements, and work/byte accounting.
   Keep M2's migration commit focused on the existing kernels; new patterns are
   separately reviewable slices. M3's standalone measurement foundation is complete.
2. **First pattern slice:** float32 broadcast scale/affine plus predicate-driven
   threshold/selection. These exercise scalar parameters and masks while retaining
   an array output. Keep explicit `Mul`/`Add` separate from a gated FMA variant.
3. **Accumulation slice:** sum, then dot product and accumulator-count experiments,
   once scalar-output preparation/retention, fold semantics, and accounting are
   supported. Dot/FMA is not a prerequisite for a non-FMA dot experiment.
4. **Movement/representation slice:** a useful copy/reorder, predicate-export, or
   conversion experiment. Fixed-width permutes/horizontal primitives remain a
   separate `archsimd` extension with build/runtime requirements.
5. **Purposeful numeric breadth:** add float64 or a narrow integer family for a
   documented question about rounding, saturation, conversion, or bit operations.
   Carryless multiplication and broad ISA coverage remain demand-driven.

These are priorities within F004, not dates or a requirement to finish all
families. No external dataset importer, TUI, resource engine, or cloud workflow
is needed to learn from the first pattern slice.

## Concrete experiment contracts

### Broadcast and scalar parameters

Start with generated `x []float32` and scalar `alpha`, then add scalar `beta` for
affine transformation. Include a per-element coefficient-array alternative only
as a separate input contract: it reads different data and cannot be pooled with
the scalar-parameter series. Record parameter names, types, and exact float bits
in case identity; rounded display strings lose signed zero/NaN distinctions.

Decide whether broadcast construction is included once per kernel call or bound
once during preparation. Compare scalar/SIMD with matching boundaries and label
any prepared-constant variant separately. Neither array timing nor this setup
comparison is a measurement of one instruction's latency.

### Predicates, masking, selection, and clamp

Define threshold-zero as `x[i] > t ? x[i] : +0`, and selection as
`x[i] > t ? x[i] : y[i]`. Equality takes the false branch. Explicitly define NaN
comparison results, signed-zero output and payload-preservation requirements.
`IfElse` selects values; it does not promise lazy evaluation, guarded memory
access, output compaction, or a shorter result slice.

Use 0%, approximately 50%, and 100% true predicates, with alternating, clustered,
and deterministically shuffled arrangements. Retain pattern/seed/fingerprint in
input identity so scalar branch predictability is visible, not an uncontrolled
fixture change. Predicate arrays need a canonical representation, such as uint8
0/1, with normalization/storage included in both timed implementations; raw
SDK-mask and packed-bit studies are separate contracts.

For clamp, choose the comparison/tie/NaN policy before comparing `Min`/`Max` with
branching or `IfElse`. The SDK emulation source's operand-sensitive comparisons
are not proof of equivalence to Go built-in min/max across every backend.

### Reductions, dot products, and FMA

Add scalar-result output handling instead of forcing a one-element result into
the existing two-input/array-output abstraction. Retain the complete result
outside validation timing without adding a full checksum to every timed call.

Floating-point addition is order-sensitive. A vector tree or multi-accumulator
fold is not generally bitwise equal to a left-to-right scalar sum. Declare the
fold/order and an approved error/reference rule for finite values, cancellation,
overflow, NaNs, infinities, and signed zero. Record accumulator count and final
fold strategy. Include final reduction/extraction in a whole-kernel comparison;
a register-only primitive study gets a different suite ID.

`MulAdd` needs three vector operands, but those operands need not be three input
arrays: affine constants and a running accumulator are different operand
sources. A three-array element-wise kernel, an affine transform, and a dot
product must have different descriptors and traffic models. Establish a valid
float32 fused reference before claiming fused equality; widening operands into
float64 `math.FMA` and rounding back is not a universal exact float32 oracle.
Separate fusion changes from accumulation-order changes in identity and tests.

### Tails, lane movement, and capability gates

Every new pattern needs zero-length correctness and sizes around its actual lane
width, plus larger nonmultiples. A zero-filled inactive lane is not automatically
neutral for min/max, negative-threshold predicates, or conversions. Count only
valid predicate lanes; mask/exclude invalid reduction lanes or supply the proper
neutral element. Preserve partial-store bounds and declare alias/overlap rules.

The article describes an older partial-load issue; verify the selected compiler
and relevant fix rather than declaring Go 1.27.3 either affected or fixed from
the article alone. Value tests are not proof of absence of memory overreads.

For permutations, define block size, index behavior, and a separate partial-block
policy. Reversing lanes within each vector is not reversing the complete array;
a changing native width must not silently change the promised operation.
`archsimd` kernels require explicit GOARCH and feature checks before execution,
with an unavailable result on unsupported hosts. Never require a feature-check
bypass to run an experiment. Record actual child-process lanes/bits/emulation;
SDK controls such as `GODEBUG=simd=0` require separately initialized processes
and do not make an architecture-specific kernel portable.

## Work and traffic models

These are candidate *logical* byte counts for float32 kernels, not DRAM counters:

| Contract | Logical input/output bytes per call |
| --- | --- |
| Copy, scalar-parameter scale/affine/threshold/clamp | `8*n` (one array read, one array write) |
| Two-array selection | `12*n` |
| Three-array element-wise MulAdd | `16*n` |
| Float32 predicate to uint8 0/1 | `5*n` |
| Float32 sum | `4*n + 4` |
| Float32 dot product | `8*n + 4` |

Record constant setup, scratch traffic, mask packing, output kind and byte units
explicitly. The current `12*n` helper must not be reused for every descriptor.
Exclude zero-work sizes from throughput claims; retain them in correctness tests.
Do not pool contexts with different parameters, mask representation, fusion,
fold order, width, worker policy, or backend/harness boundaries.

## Acceptance checklist for each admitted pattern

- [ ] A documented experiment/learning question and selected initial types.
- [ ] Verified portable versus architecture-specific API/capability status.
- [ ] Scalar reference, arity, parameter sources, output shape and representation.
- [ ] Exact parameter/fixture identity and complete work/byte accounting.
- [ ] Explicit arithmetic, alias, inactive-lane and numerical acceptance rules.
- [ ] Boundary and distribution cases, including mixed predicates and tails.
- [ ] Matched measurement boundaries; setup/final fold/export costs declared.
- [ ] No fixed timing thresholds; gains, regressions, and unavailable cases kept.
- [ ] CLI/report fixtures and Wiki examples updated only when the kernel ships.

The remaining API/type inventory stays a coverage ledger. Unsupported methods,
infrastructure-only utilities and deferred patterns get reasons instead of fake
benchmarks or silent casts. See the [Wiki pattern review](https://github.com/akolybelnikov/go-simd/wiki/12-SIMD-patterns-and-reading)
for the user-facing current-versus-planned distinction.
