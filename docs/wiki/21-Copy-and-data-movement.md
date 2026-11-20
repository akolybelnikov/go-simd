# Float32 copy and data movement

Moving an array into another buffer is a useful comparison for workloads that do little arithmetic. An exact copy preserves every input bit.

## How the methods differ

Compare a scalar source loop, Go’s built-in copy, and vector methods with bounded partial writes or a scalar tail. Each preserves signed zero and NaN payload bits. A source-level loop or built-in copy can itself be optimized by the compiler or runtime.

## What the benchmark measures

The timed work is reading inputs, writing outputs, and handling any setup or tail. For float32, the logical work is four bytes read and four written per value. Cache residency and physical memory traffic can differ from that figure.

## Try it

In **Explore**, select **Float32 copy**. Compare the built-in with the scalar loop, then compare a vector method at the same size. Try small and large buffers before drawing a conclusion about copying.

See also [Float64 copy](31-Exact-Float64-copy).
