# Float64 addition

Float64 addition combines corresponding values from two double-precision arrays. It is useful when a calculation needs more precision or range than float32 provides.

## How the methods differ

Each result uses normal binary64 rounding. The scalar path handles one pair at a time and the vector path handles pairs in lanes. At the same hardware vector width, there are half as many float64 lanes as float32 lanes.

## What the benchmark measures

The timing covers length checks, loads, addition, stores, and bounded tail work. There are 24 logical bytes per result from two eight-byte reads and one eight-byte write.

## Try it

Use **Float64 addition** in **Explore**. Choose sizes below, at, and above the measured float64 lane count. Compare float32 and float64 as separate experiments because both precision and logical traffic change.

See also [addition methods](04-Operation-guide#scalar-addition) and [working with arrays](08-Using-the-kernels).
