# Accumulator count experiments

One, two, or four independent vector accumulators all reduce an array to a single result. Giving the processor separate chains of additions can change throughput.

## How the methods differ

Successive full vectors are assigned to accumulators in round-robin order. At the end, the methods merge accumulators lane by lane from left to right, fold the lanes from low to high, and finish any tail. Each addition rounds, so even a different accumulator count can change the result of a cancellation-heavy sum.

Let w be the vector width. In an array of 8w values, put 2²⁴ at position 0,
1 at position w, and −2²⁴ at position 2w, leaving all other positions zero.
The exact total is 1. A single vector accumulator returns 0, while two
accumulators return 1 because the two large terms land in the same
accumulator and cancel before the merge.

## What the benchmark measures

The lab times the full reduction: reading values, adding into accumulators, merging, folding lanes, and processing any remainder. Loop unrolling and register use can affect the observed difference. Short batches may spend more time on the merge than they save during accumulation.

## Try it

Compare one, two, and four accumulators against the same scalar reference at the same size. Try a short batch and a much longer batch. If two sizes give different rankings, include both in your report.

See also [sum and reduction order](14-Sum-and-reduction-order) and [unfused dot product](15-Dot-product-and-unfused-rounding).
