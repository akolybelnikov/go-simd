# 12 — SIMD pattern gaps and further reading

**Status:** catalogue review and proposed experiments, reviewed 2026-10-27.
The lab still implements the six kernels in the operation guide. Patterns below
are not new CLI flags or benchmark variants.

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
| Broadcast constants | No parameterized kernel | Scale or affine transform with scalar parameters |
| Comparisons and mask logic | No predicate kernels | Threshold distributions, range predicates, explicit mask representation |
| Masking and conditional selection | No masked/blended output | Keep-or-zero and choose-between-inputs transforms |
| Min/max clamp | No clamp kernel | Compare branches, selection, and min/max only under matching numerical rules |
| FMA and streaming accumulation | Existing expression uses `Mul`/`Add`, without reduction | A separate multiply-add mode; sum/dot and accumulator strategies |
| Permutes, horizontal folds, extraction | No lane-rearrangement or scalar-result study | An explicitly scoped architecture-specific experiment or portable composed reduction |
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

These are mathematical result sketches, not currently runnable lab operations:

- **Broadcast affine:** for `x=[-2,0,3,6]`, scalar `alpha=2`, `beta=1`, the result
  of `alpha*x+beta` is `[-3,1,7,13]`. Scalar parameters differ from reading two
  additional coefficient arrays; setup and traffic accounting must reflect that.
- **Predicate and selection:** with threshold `1`, the predicate for that `x`
  is `[false,false,true,true]`. Keep-or-zero yields `[0,0,3,6]`. Choosing `x` when
  true and `y=[-4,2,5,7]` otherwise yields `[-4,2,3,6]`.
- **Reduction:** a sum or dot product returns one scalar, unlike today's output
  slices. Vector folds and multiple accumulators change addition order, which
  can change floating-point rounding. A correctness/error rule is required.
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

1. Complete M2's shared descriptors for the existing six kernels, retaining
   current identities. Describe scalar parameters and output shapes cleanly.
2. Add a focused float32 broadcast and predicate/selection slice with equivalent
   scalar references. Cover none/mixed/all-true predicate arrangements.
3. Add sum/dot experiments after scalar-output and reduction-order contracts.
   Gate fused variants on a valid FMA reference, independently of fold order.
4. Add movement, conversion, or a gated `archsimd` experiment when it answers a
   documented learning question; grow numeric types purposefully.

This is the operations feature backlog, not a change to the architecture
migration sequence. No broadcast, mask, permutation, sum, or dot command has
been implemented by this documentation update.
