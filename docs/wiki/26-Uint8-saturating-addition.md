# Saturating byte addition

Saturating addition clips a byte sum to 255. For example, combining pixel values 250 and 20 produces 255.

## How the methods differ

The scalar method widens each pair, adds, then caps the result. The vector method applies saturating addition to byte lanes. Both return the same integer value for every input pair. Ordinary byte wrapping would return 14 for 250 + 20.

## What the benchmark measures

Each call reads two byte arrays and writes one byte result per position. The timed work includes widening and clipping where required, loads, stores, and any incomplete-vector handling.

## Try it

Select **Uint8 saturating addition** in **Explore**. Compare one batch smaller than the byte lane count with one larger than it, and keep the input fixture identical between methods.

See also [clamp and numerical contracts](22-Clamp-and-numerical-contracts).
