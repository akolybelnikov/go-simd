# Fixed-block horizontal sum

A block sum produces one total per group of eight values. It can reduce a signal into consecutive windows or summarize groups of samples.

## How the methods differ

Both methods add pairs (0,4), (1,5), (2,6), and (3,7), then fold the four results from left to right. Each add rounds to float32. The last block is padded with positive zero. The vector method uses lanes for the pair stage while maintaining the same order as the scalar method.

For input [−1e20, 1, 1e20, 0, 1e20, 0, −1e20, 0, 2], the two block results are [1, 2].

## What the benchmark measures

The timer includes loading and preparing each block, summing its pairs, folding those sums, and writing one output per block. For n inputs, the output length is ceil(n/8). Hashing and other report work happen outside the timed batch.

## Try it

Select **Fixed-eight-value horizontal sum** in **Explore**. Preview 7, 8, and 9 values in separate runs to see the transition from a partial block to an extra output.

See also [whole-array sum](14-Sum-and-reduction-order).
