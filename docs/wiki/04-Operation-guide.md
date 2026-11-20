# Operation guide

A comparison asks how two methods perform the same kind of work. The [measurement guide](32-Operation-measurement-guide) explains the methods, uses, and timing boundaries for each operation.

| Family | Examples | What to watch |
| --- | --- | --- |
| Element by element | Addition, scaling, affine transforms | Lane width, loading and storing, and the final partial vector |
| Comparisons and masks | Threshold, selection, packed masks, clamp | Which values pass, how the output is stored, and how ties are handled |
| Reductions | Sum, dot product, block sum | Addition order, rounding, merges, and tails |
| Exact representation | Copy, bit reinterpretation, XOR | Preserving bits and the amount of data moved |
| Integer arithmetic | Saturating addition, carryless multiplication | Clipping rules or polynomial arithmetic |

## Addition methods

### Scalar addition

For arrays [1, 2, 3] and [4, 5, 6], addition produces [5, 7, 9]. The scalar method works through one pair at a time. The time includes reading the inputs, adding, and writing each result.

### Partial vectors

A vector method processes several pairs together. Partial loads and stores work with the valid values when the last chunk is shorter than the vector width. Try a size just below the measured width and one just above it.

### Full vectors with a scalar tail

A full-vector method handles complete chunks first, then finishes remaining values one by one. With eight lanes and nine values, it handles eight together and one in the tail. The lane count depends on the machine and element type.

### Parallel addition

A parallel method divides the arrays across workers. Its timing includes coordination and waiting for workers. Comparing it with a serial scalar loop measures the combined effects of vectorization and concurrency.

## Compute expression

The expression a² + b² + ab performs more arithmetic per value than addition while using the same two input arrays and one output array. Compare both operations at several sizes to see how extra arithmetic affects the result. See [Understanding results](06-Understanding-results) for how to read those differences.
