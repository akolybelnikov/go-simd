# Export predicates as packed bits

A packed mask stores eight decisions in one byte. Compact masks are useful when a later stage can read decisions directly as bits.

## How the methods differ

Input index i maps to bit i mod 8 in byte floor(i / 8), starting at the least significant bit. The high bits of the last byte are cleared when it contains fewer than eight decisions. For 9 input values you get 2 output bytes, and only the lowest bit of the second byte holds a decision.

## What the benchmark measures

The timer includes threshold comparisons, bit packing, and output stores, including the partial byte at the end. A packed mask reduces output size while adding packing work.

## Try it

Compare the packed method just below and above byte and vector boundaries. Then compare the same fixture as a byte mask to see whether the storage savings offset the extra computation for that batch.

See also [byte masks](19-Predicate-export-and-byte-masks).
