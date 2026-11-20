# XOR of float32 bit patterns

XOR marks positions where two bit patterns differ. Opaque bit patterns can be carried in float32 storage while being treated as binary words.

## How the methods differ

Compare the stored bits as integers. The scalar method processes one word at a time, while the vector method XORs groups of lanes. For example, XOR of the bit patterns for positive and negative zero produces the sign-bit mask.

## What the benchmark measures

The complete timed batch includes reading the two arrays, performing the XOR, and writing outputs. There are twelve logical bytes per pair of 32-bit values.

## Try it

Select **XOR of opaque Float32 bit patterns** in **Explore**. Trial sizes near a vector boundary and over much larger arrays reveal different overhead and movement costs.

See also [bit reinterpretation](27-Bit-reinterpretation-round-trip).
