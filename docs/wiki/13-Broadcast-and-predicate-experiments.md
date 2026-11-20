# Broadcast and predicate experiments

Broadcasting applies the same setting to every value in an array. Predicates test each value and can either keep it, replace it, or choose between two arrays.

## How the methods differ

| Operation | Rule |
| --- | --- |
| Scale | Multiply by a shared factor |
| Affine transform | Multiply by a shared factor, round to float32, then add an offset |
| Threshold to zero | Keep values strictly above the cutoff and write positive zero elsewhere |
| Conditional selection | Use the first array above the cutoff and the second array otherwise |
| Range filter | Keep values strictly between the lower and upper bounds |

A scalar method processes one position at a time. A vector method copies constants into each lane and uses masks for the comparisons. A value equal to a strict cutoff or bound does not pass. A NaN also fails an ordered comparison.

## What the benchmark measures

The timing includes loading values, setting up vector constants, applying each comparison or calculation, and storing the resulting array. A scalar branch may behave differently when passing values alternate, cluster, or appear in shuffled order.

## Try it

Compare scaling with addition at the same batch size, then try thresholding. To study how the distribution affects timing, hold the cutoff and batch size fixed and vary one fixture pattern using the [fixture options](03-Command-registry#fixture-options).

See also [byte masks](19-Predicate-export-and-byte-masks) and [clamp](22-Clamp-and-numerical-contracts).
