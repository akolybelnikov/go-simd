# SIMD patterns and further reading

SIMD is useful when the same operation can be applied to independent values. Think about the data layout, the work done in each lane, and how the result uses those lanes.

| Pattern | An experiment to try |
| --- | --- |
| Contiguous reads and writes | [Copy the same bits at small and large sizes](21-Copy-and-data-movement) |
| Reusing a constant in every lane | [Scale or transform an array](13-Broadcast-and-predicate-experiments) |
| Using comparison results | [Write a mask as bytes or bits](20-Packed-predicate-masks) |
| Combining lane results | [Compare accumulation and fold orders](16-Accumulator-count-experiments) |
| Combining arithmetic into one rounding | [Compare fused and unfused affine transforms](17-Fused-affine-and-rounding) |

Portable vector code describes the work in terms of lanes. A machine or compiler can realize that work in different ways. The timing of a whole batch measures the effect of that choice alongside setup, data access, and tail handling.

For background on vector operations in Go, read [Jesús Espino’s SIMD and Go](https://internals-for-interns.com/posts/go-simd/).
