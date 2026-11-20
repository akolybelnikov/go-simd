# Fused dot and fold orders

A fused dot product uses fused multiply-add to update a running total. This pattern occurs in matrix multiplication and digital filters.

## How the methods differ

Each fused update rounds the product plus the accumulator once. Merging accumulators and folding lanes are separate rounded additions. Fusion and fold order therefore affect results independently. A single fused accumulator can disagree with multiple fused accumulators on the same inputs.

## What the benchmark measures

The measurement covers reading both input arrays, fused updates, merging, lane folding, and tail work. The report's execution path helps separate a software run from a vector FMA run.

## Try it

In **Explore**, select **Float32 fused dot**. Compare one and four vector accumulators against the scalar reference, then compare the same fixtures against the unfused dot product. Keep fusion and accumulator count as separate changes when interpreting the outcomes.

See also [accumulator counts](16-Accumulator-count-experiments) and [unfused rounding](15-Dot-product-and-unfused-rounding).
