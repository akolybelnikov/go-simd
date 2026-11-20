# How the operation benchmarks work

Each comparison uses generated inputs of the same size and fixture for both
methods. The worker prepares inputs and output buffers before timing. Go’s
benchmark harness then calls one kernel repeatedly. One operation in the
report is one call over the whole batch. The timer includes work inside that
call, such as length checks, loads, arithmetic, stores, and tails. Results
are retained or checked after timing. The browser shows the median and
individual repeats. A small batch can be dominated by setup, a large one by
data movement, and background activity can affect either.

Widths and remainders depend on the measuring host and selected element
type. Start with a sizable batch, then try a length just below or above a
vector boundary to see how the implementation handles the remainder. For
more on variability and interpreting differences, read
[Understanding results](06-Understanding-results).

## Addition

<!-- operation: addition -->

Adds corresponding values from two float32 arrays. This is the same basic operation as mixing aligned audio streams or adding two numeric columns.

The scalar loop reads and writes one pair at a time. The vector methods use partial loads and stores or full vectors followed by a scalar tail. The parallel variant divides the arrays among workers.

**What is timed.** Two input reads and one output write per value, plus the kernel’s checks and tail handling. The parallel variant also measures worker coordination. To study vectorization alone, compare the scalar and serial methods. [Kernel and tail details](04-Operation-guide#scalar-addition)

## Float64 addition

<!-- operation: addition64 -->

Adds corresponding double-precision values. Use this pattern when combining scientific data arrays that need float64 precision.

The scalar and vector loops both apply ordinary binary64 addition and round at each element’s addition. The vector path uses the host’s actual float64 lane count and bounded loads and stores for the final batch of lanes.

**What is timed.** The kernel reads two eight-byte values and writes one per position, plus checks and tail handling. The float32 and float64 benchmarks use different data widths, so compare them as separate experiments. [Float64 details](25-Float64-addition)

## Compute heavy expression

<!-- operation: computation -->

Evaluates a quadratic expression for each pair of values. It is a small example of the arithmetic used in per-point scoring and numerical transforms.

Both methods compute the two squares and the cross product, then add the terms. The vector version applies these operations to several pairs at once using partial-vector loads and stores.

**What is timed.** The full expression plus data movement and stores. Comparing it with addition on the same array size shows how adding arithmetic can change the timing balance for a similar amount of input and output. [Kernel details](04-Operation-guide)

## Scale

<!-- operation: scale -->

Multiplies every value by a shared factor. Common examples include applying audio gain and converting units in a numeric array.

The scalar loop reuses one multiplier. The vector version broadcasts that scalar to all lanes and multiplies a chunk at a time.

**What is timed.** The float32 multiply, the broadcast setup inside the vector call, loads, and output stores. Inputs and the multiplier are prepared before timing. [Broadcast details](13-Broadcast-and-predicate-experiments)

## Affine transform

<!-- operation: affine -->

Multiplies every value by a shared factor and adds an offset. This is useful for sensor calibration and linear rescaling.

The methods explicitly round the float32 product first, then round again after adding the offset. The vector version broadcasts both constants and uses separate multiply and add operations.

**What is timed.** Broadcast setup, the two arithmetic stages, loads, and stores. Use the fused affine benchmark to study the alternative rounding rule. [Unfused affine details](13-Broadcast-and-predicate-experiments)

## Fused affine transform

<!-- operation: affinefma -->

Applies a scale and offset with one final rounding. This is useful in calibration and numerical transforms that specify fused multiply-add.

The kernel computes the product and offset as one fused operation. Depending on runtime support, the benchmark uses a vector path or a software path that preserves the same one-rounding contract.

**What is timed.** The complete affine kernel, including whichever implementation ran and any edge handling. Inspect the recorded path before attributing a difference to hardware FMA. [Fused affine details](17-Fused-affine-and-rounding)

## Threshold to zero

<!-- operation: threshold -->

Keeps values above a cutoff and replaces the rest with zero. This pattern can suppress low readings or build a dense thresholded signal.

The scalar method tests each value. The vector method compares all lanes to the threshold and uses the resulting mask to retain a value or write positive zero. Values equal to the cutoff and NaNs take the zero path.

**What is timed.** Testing the predicate and writing one output for every input. The chosen distribution of passing values can affect how the scalar branch behaves. [Threshold details](13-Broadcast-and-predicate-experiments)

## Conditional selection

<!-- operation: selection -->

Chooses between two aligned arrays based on a cutoff. For example, a pipeline can keep a reading above a threshold and otherwise use a fallback value.

The scalar branch and vector mask both select the first array’s value when it exceeds the cutoff and the second array’s value otherwise. The chosen value’s original bits are copied through.

**What is timed.** The predicate, selection, and dense output stores. The vector path loads both candidate arrays and performs a lane-wise selection. [Selection details](13-Broadcast-and-predicate-experiments)

## Range filtering

<!-- operation: rangefilter -->

Keeps values strictly between two bounds and writes zero elsewhere. This can gate readings to a range while preserving the array’s shape.

The scalar method checks both bounds. The vector method combines two comparison masks, then retains the value or writes positive zero. Exact boundary values and NaNs take the zero path.

**What is timed.** Both comparisons, mask combination or scalar branching, and one output store per input. The benchmark retains every input position in the output. [Range mask details](13-Broadcast-and-predicate-experiments)

## Byte mask

<!-- operation: predicate -->

Writes one byte per value to indicate whether it passed a cutoff. The mask can carry selection decisions into a later processing stage.

A strict ordered comparison produces a canonical byte 0 or 1. The vector version turns its internal lane mask into those byte values.

**What is timed.** Comparison, conversion or extraction, and byte stores. A faster comparison alone may contribute little if writing the mask dominates. [Byte mask details](19-Predicate-export-and-byte-masks)

## Packed bit mask

<!-- operation: predicatepack -->

Stores cutoff decisions one bit at a time, with eight decisions per byte. This representation is useful for compact validity or selection masks.

The methods use the same ordered predicate as the byte mask. Position i goes into the low-to-high bit sequence of its output byte. Unused high bits in the last byte are cleared.

**What is timed.** Comparison, bit packing, and output stores, including the last partial byte. Compare with the byte mask to see the tradeoff between packing work and output size. [Packed mask details](20-Packed-predicate-masks)

## Float32 copy

<!-- operation: copy -->

Copies float32 values while preserving every bit. Typical uses include staging data in another buffer or retaining a snapshot without changing its representation.

The benchmark compares a source-level loop, Go's built-in copy, and vector methods with partial writes or a scalar tail. All preserve details such as signed zero and NaN payload bits.

**What is timed.** Reading the source, writing the destination, and any checks or tail work. Output hashing happens after timing. Go’s built-in copy may already use optimized runtime code. [Copy details](21-Copy-and-data-movement)

## Float64 copy

<!-- operation: copy64 -->

Copies double-precision values without changing their binary representation. Use this pattern when moving numeric buffers whose exact bits must survive a copy.

The source loop, Go’s built-in copy, and the partial-vector implementation copy the same binary64 input to the output. The vector path uses the actual float64 width.

**What is timed.** One eight-byte read and write per value plus checks and any tail work. Larger copies can stress a different part of the memory hierarchy than shorter ones. [Float64 copy details](31-Exact-Float64-copy)

## Three array multiply add

<!-- operation: muladd -->

Computes a product and adds a third array element with one rounding. This is a building block for weighted sums, signal processing, and numerical kernels.

For each position, the methods compute the fused result of the first two inputs’ product plus the third input. The benchmark records the software or vector path that performs the fused operation.

**What is timed.** The full kernel with three input streams and one output stream. Compare with fused affine to see how reading three arrays changes the workload compared with using shared constants. [Three-array details](24-Three-array-fused-multiply-add)

## Clamp

<!-- operation: clamp -->

Clips values to a minimum and maximum. Applications include keeping normalized pixels or sensor values within an allowed interval.

The scalar and vector strategies replace only values outside the bounds. Values equal to a boundary and NaNs preserve their original bits. The vector min/max strategy repairs cases where straight min/max would break that contract.

**What is timed.** Clipping and stores, including any repairs in the min/max method. Input distributions determine how often the scalar branches take the clipping path. [Clamp details](22-Clamp-and-numerical-contracts)

## Array sum

<!-- operation: sum -->

Adds all the values in one float32 array and returns a total. This is useful for aggregating samples or computing the numerator of an average.

The scalar method adds from left to right. The vector methods keep one, two, or four running vectors, then merge and fold the lanes and handle the remainder. Those methods change the addition order, so rounding can change the total.

**What is timed.** The whole reduction, including accumulator merges, the lane fold, and the tail. The benchmark retains the result bits so timing can be read alongside the numerical contract. [Sum and rounding details](14-Sum-and-reduction-order)

## Dot product

<!-- operation: dot -->

Multiplies corresponding values and sums the products. Dot products are used in similarity scoring, projections, and matrix multiplication.

The scalar and vector implementations round each float32 product before adding it. Vector accumulation changes the fold order. Separate accumulators can reduce the dependency between successive adds.

**What is timed.** Both input reads, products, accumulation, merging, and the scalar tail for one complete dot product. The fused dot benchmark uses a different rounding rule. [Unfused dot details](15-Dot-product-and-unfused-rounding)

## Fused dot product

<!-- operation: dotfma -->

Accumulates products with fused multiply-add. This is a building block for numerical kernels such as matrix multiplication and digital filters.

Each accumulation step rounds once after adding the product to the accumulator. The vector methods combine lane accumulators at the end, and the selected runtime path can use software or vector FMA.

**What is timed.** The full dot product, including the fused updates, merge, lane fold, and tail. Read the recorded execution path and accumulation order before comparing results. [Fused dot details](18-Fused-dot-and-fold-orders)

## Saturating byte addition

<!-- operation: saturatingadd8 -->

Adds unsigned byte values and caps each result at 255. This is useful when combining byte-valued pixels or other signals that should clip at their maximum.

The scalar method widens each pair before adding and capping. The vector method uses saturating byte addition, which produces the same capped result without wrapping to zero.

**What is timed.** Reading the two byte arrays, the addition and cap, and output byte stores. Test lengths around byte-vector boundaries using the lane count recorded for this operation. [Saturating addition details](26-Uint8-saturating-addition)

## Float32 bit round trip

<!-- operation: bitroundtrip -->

Views float32 values as their 32-bit representation and then restores that view. Exact bit access is useful in binary formats and diagnostics.

The scalar method bit-casts each value. The vector version reinterprets all lanes as integers and converts the representation back. Signed zeros and NaN payloads retain every original bit.

**What is timed.** The complete compiled load, reinterpretation, and store path. The compiler may eliminate representation-only instructions. [Bit round-trip details](27-Bit-reinterpretation-round-trip)

## Bitwise XOR

<!-- operation: xorbits -->

Applies XOR to pairs of 32-bit patterns stored in float32 arrays. This kind of operation can combine or detect differences between opaque binary masks.

The methods treat the float32 bits as integers and XOR corresponding pairs. The output uses those same bits as its float32 representation.

**What is timed.** Reading both input arrays, the bitwise work, and output stores. The benchmark treats payloads as opaque bits throughout. [Bitwise details](28-XOR-of-float32-bit-patterns)

## Eight value block sum

<!-- operation: sum8 -->

Returns one sum per block of eight float32 values. This is useful when reducing data into fixed-size windows, such as downsampling by groups of eight.

Both methods form the same four pairs and add those pair sums in the same order. A short final block is padded with positive zero. The vector method carries out the paired additions before folding each block.

**What is timed.** Preparing each block inside the kernel, adding its values, and storing one result per block. The operation and output shape differ from returning one total for the entire array. [Fixed block details](29-Fixed-block-horizontal-sum)

## Carryless multiplication

<!-- operation: clmul64 -->

Multiplies pairs of 64-bit bit patterns as polynomials and keeps the low 64 bits. Carryless products are building blocks for CRC and finite-field arithmetic.

The scalar method combines shifted partial products with XOR. The vector path handles even and odd lanes and assembles the corresponding low words. This operation measures a low-word product, which is one part of wider carryless algorithms.

**What is timed.** The complete kernel, including extraction, scratch work, and stores. The recorded capability describes the runtime path available on the measuring host. [Carryless multiplication details](30-Carryless-polynomial-multiplication)
