# Clamp and numerical contracts

Clamp keeps input values inside an interval. For instance, bounds of −1 and 1 turn [−2, −1, 0, 1, 2] into [−1, −1, 0, 1, 1].

## How the methods differ

Values strictly below or above a bound become that endpoint. Values equal to a bound, including signed-zero ties, keep their original bits. NaNs keep their original bits too. The scalar method branches, one vector method selects with comparisons, and the min/max method corrects cases where a raw min/max sequence would choose the wrong bits.

For bounds of positive zero and negative zero, negative inputs become
positive zero and positive inputs become negative zero. An input that is
already zero keeps its own sign. That difference is visible in the bits even
though both zeros compare equal.

## What the benchmark measures

The time includes comparisons, stores, and corrections for ties and NaNs in the min/max method. The share and arrangement of values that get clipped can affect scalar branches.

## Try it

In **Explore**, select **Ordered clamp** and compare the vector methods against the same scalar reference. Hold the bounds fixed while changing the mix of clipped and unclipped positions. The [fixture options](03-Command-registry#fixture-options) explain the available patterns.

See also [threshold and range experiments](13-Broadcast-and-predicate-experiments).
