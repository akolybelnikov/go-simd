# Export predicates as byte masks

A byte mask stores one decision per input value. Other steps can reuse it to select readings or mark entries that pass a cutoff.

## How the methods differ

Each input strictly above the cutoff becomes byte 1. Equality, NaN, and all other values become byte 0. A scalar loop writes the bytes one at a time. A vector method first compares lanes, then turns each decision into a byte. The final partial vector writes only the corresponding output values.

## What the benchmark measures

The time includes the comparisons, any extraction or conversion from vector masks to bytes, and output stores. For float32 input, each value accounts for four input bytes and one output byte of logical work.

## Try it

Select **Ordered predicate export** for byte output and compare the available methods. Keep the cutoff fixed when testing the effect of alternating, clustered, and shuffled decisions.

See also [packed bit masks](20-Packed-predicate-masks).
