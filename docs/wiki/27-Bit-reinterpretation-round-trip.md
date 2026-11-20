# Bit reinterpretation round trip

A bit reinterpretation reads the bits of a float32 as a 32-bit word, then restores the same float32 representation. It is useful for bit inspection, diagnostics, and binary formats.

## How the methods differ

The scalar method changes the view of each value separately. The vector method changes the view of a group of lanes. This preserves every bit, including signed zero, subnormal representations, and NaN payloads.

## What the benchmark measures

The benchmark times the compiled load, reinterpretation, and store path. A compiler may remove work that changes only the type view of the same bits. Compare the complete method with exact copying to help interpret the result.

## Try it

Choose **Float32 exact bit reinterpretation round trip** in **Explore**. Vary array size while keeping the same data. Compare with **Float32 copy** to see how close the measured costs are on your machine.

See also [bitwise XOR](28-XOR-of-float32-bit-patterns).
