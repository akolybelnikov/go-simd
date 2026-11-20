# Three-array fused multiply-add

Each result is a product plus a corresponding value from a third array, rounded once. This pattern shows up in weighted sums and signal processing.

## How the methods differ

The scalar and vector methods use the same one-rounding rule. For each position, the first two values supply the product and the third supplies the addend. The vector method handles several independent positions together and may use software FMA where needed.

## What the benchmark measures

The benchmark includes all three input reads, the fused operation, and the output write. That is 16 logical bytes per float32 result. Read the actual FMA path before making a hardware comparison.

## Try it

In **Explore**, select **Three-array fused multiply-add**. Compare the scalar and vector paths, then use fused affine at the same size to see how using shared constants changes the amount of array data moved.

See also [fused affine transforms](17-Fused-affine-and-rounding).
