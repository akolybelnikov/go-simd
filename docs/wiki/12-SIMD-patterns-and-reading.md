# 12 — SIMD pattern gaps and further reading

**Status:** F004 broadcast/predicate and sum checkpoints implemented, reviewed 2026-11-03.
The lab has eight operations/eighteen implementations. The
[worked broadcast/predicate guide](https://github.com/akolybelnikov/go-simd/wiki/13-Broadcast-and-predicate-experiments)
and [sum guide](https://github.com/akolybelnikov/go-simd/wiki/14-Sum-and-reduction-order)
separate delivered patterns from the remaining proposals below.

## Reading credit

[**SIMD and Go**](https://internals-for-interns.com/posts/go-simd/), by
**Jesús Espino**, *Internals for Interns*, published **2026-10-19**, prompted this
review. Its examples target Go 1.27.0 and mainly amd64 `simd/archsimd`.
The gap assessment and proposed experiments here are specific to our repository,
with API availability checked independently against the installed Go 1.27.3 SDK.
The small numerical sketches below are our own planning examples.

## What our current catalogue misses

| Pattern | Current lab coverage | Useful proposed experiment |
| --- | --- | --- |
| Contiguous loads/stores and partial/scalar tails | Exercised in existing addition/expression kernels | Keep as controls; isolate movement only for a concrete question |
| Broadcast constants | Scale/affine implemented | Compare per-call broadcasts and matched scalar controls |
| Comparisons and mask logic | Threshold/selection/rangefilter implemented | Distribution/seed controls and range-mask And; exported masks deferred |
| Masking and conditional selection | Masked/IfElse outputs implemented | Threshold-to-zero and bit-preserving two-array selection |
| Min/max clamp | No clamp kernel | Compare branches, selection, and min/max only under matching numerical rules |
| Streaming accumulation | Sum with one vector accumulator and a final scalar fold implemented | Dot product and multiple independent accumulators next |
| Explicit FMA | No fused mode; existing expression/affine are unfused | A separately gated multiply-add contract |
| Permutes, horizontal folds, extraction | Sum includes portable store/fold cost; no isolated lane-rearrangement study | An explicitly scoped architecture-specific experiment |
| Conversion, bit representation, saturation | No kernels for these families | A later bounded numeric/representation experiment |

The useful next step is broader *behavioral* coverage, before nominal coverage
of every numeric type and method. See the [tracked F004 plan](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-operation-patterns.md)
for proposed identities, contracts, accounting, and acceptance criteria.

## Portable methods and architecture-specific methods differ

The local portable SDK provides `BroadcastFloat32s`, comparisons returning
`Mask32s`, mask `And`/`Or`, and vector `Masked`/`IfElse`. It does not expose
`Float32s.Permute`, portable horizontal-pair addition, lane `GetElem`, or
`Mask32s.ToBits` in the inspected inventory.

Verified amd64 examples of the separate `archsimd` surface include
`Float32x8.Permute`, `Mask32x8.ToBits`, `Float32x4.ConcatAddPairs`,
`Float32x8.GetLo/GetHi`, and `Float32x4.GetElem`. Method availability varies by
type/width; a method shown on a four-lane type is not automatically available
on an eight-lane type. Portable reduction can instead be composed using vector
storage and a final scalar fold, whose cost belongs in an end-to-end kernel.

Inspect your selected SDK without executing an architecture-specific kernel:

```powershell
$env:GOEXPERIMENT = 'simd'
go doc simd.BroadcastFloat32s
go doc simd.Float32s.IfElse
go doc simd.Mask32s
go doc simd/archsimd.Float32x8.Permute
```

Listing a method does not establish hardware support for executing it. Proposed
`archsimd` experiments need build/runtime capability gates and unavailable states.
The article's older partial-load caveat also needs a selected-toolchain check;
we do not infer that a particular fix is present merely from a newer version label.

## Original planning examples

These mathematical sketches informed the plan; broadcast/predicates and sum
now have runnable examples linked above. Reordering and dot remain proposals:

- **Broadcast affine:** for `x=[-2,0,3,6]`, scalar `alpha=2`, `beta=1`, the result
  of `alpha*x+beta` is `[-3,1,7,13]`. Scalar parameters differ from reading two
  additional coefficient arrays; setup and traffic accounting must reflect that.
- **Predicate and selection:** with threshold `1`, the predicate for that `x`
  is `[false,false,true,true]`. Keep-or-zero yields `[0,0,3,6]`. Choosing `x` when
  true and `y=[-4,2,5,7]` otherwise yields `[-4,2,3,6]`.
- **Reduction:** sum returns one scalar through its own typed path. Its specified
  scalar/vector orders can round differently. Dot and additional accumulator
  strategies still need their own correctness/error rules.
- **Reordering:** reversing lanes in each block is different from reversing an
  entire slice. State block size and partial-block behavior in case identity.

A mask selects values; it is not lazy evaluation, memory-access protection,
compaction, or a promise that fewer lanes will execute. Predicate output also
needs an explicit format: logical 0/1 elements, an SDK integer-vector export,
and packed bits are different contracts with different conversion costs.

For tails, padded zero lanes can satisfy a negative threshold and corrupt a
count, or bias a min/max reduction. Track valid lanes and the appropriate neutral
element rather than assuming addition's tail treatment works for every pattern.

## Proposed delivery priorities

1. M2 foundation is complete: shared descriptors/bindings for the existing six
   kernels preserve current identities. Scalar parameter descriptions and output
   shapes support the array-parameter slice; sum now implements scalar-result
   preparation and retention.
2. First slice complete: scale/affine, threshold/selection and range mask
   composition, with scalar controls and none/mixed/all-true arrangements.
3. Sum checkpoint complete: specified folds, typed scalar results, fixture and
   width identity, retained result bits. Dot and multiple accumulators follow;
   gate fused variants on a valid FMA reference, independently of fold order.
4. Add movement, conversion, or a gated `archsimd` experiment when it answers a
   documented learning question; grow numeric types purposefully.

This is the operations feature backlog, not a change to the architecture
migration sequence. Broadcast/masking/selection and sum are implemented.
Permutation, dot, explicit FMA and output-mask representations remain deferred.
