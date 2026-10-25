// Package computation provides scalar and SIMD experiments on float32 slices.
//
// All operations require equal input and output lengths and leave output unchanged
// on a length error. Output may be disjoint from the inputs or exactly alias an
// entire input slice. Partial overlap between output and an input is unsupported;
// callers must avoid it, especially for parallel operations.
package computation
