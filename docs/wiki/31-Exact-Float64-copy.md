# Exact Float64 copy

Copying a double-precision array is a movement operation. A faithful copy preserves the data’s exact representation, including signed zeros, infinities, and NaN payloads.

## How the methods differ

The scalar loop assigns each value. The built-in Go copy uses the compiler and runtime’s chosen method. The vector path divides the array into chunks at the measured float64 width, with bounded loads and stores at the end.

## What the benchmark measures

The timed work includes length checks, loads, stores, and partial-vector handling. Each value accounts for eight logical bytes read and eight written. Fixture preparation and output verification happen outside timing.

## Try it

Select **Exact Float64 copy** in **Explore**. Try the built-in and vector methods against the same scalar reference at a small size and at a larger working set.

See also [Float32 copy and data movement](21-Copy-and-data-movement).
