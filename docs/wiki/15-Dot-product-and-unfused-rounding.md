# Dot product and unfused rounding

A dot product multiplies corresponding values from two arrays and adds the products. It is used for projections, similarity scoring, and matrix multiplication.

## How the methods differ

The scalar method adds the products from left to right. Vector methods accumulate products in lanes and then combine them. In both cases each product is rounded to float32 before the addition rounds again. That first rounding can lose a small residual that a fused update would retain.

## What the benchmark measures

The timing includes reading both arrays, computing products, accumulating them, folding vector lanes, and processing the tail. One reported operation is one dot product of the two whole arrays.

## Try it

Select **Float32 dot product** in **Explore**, and keep both the batch size and the input fixture fixed while changing the method. Use the cancellation or fusion-sensitive fixture to make rounding differences easier to see.

See also [fused dot and fold orders](18-Fused-dot-and-fold-orders).
