# Sum and reduction order

An array sum reduces many values to one total. Adding readings to obtain a total or calculate an average uses this pattern.

## How the methods differ

The scalar loop adds from left to right starting at positive zero. A vector sum adds into separate lanes, folds the accumulated lanes from low to high, then handles the scalar tail. Each float32 addition rounds, so these orders can return different totals.

For example, in float32, adding 1 to 2²⁴ rounds back to 2²⁴. If the opposite large terms cancel first, the same mathematical sum can retain the 1.

## What the benchmark measures

The timed batch is one complete sum, including vector accumulation, folding the lanes, and processing remaining values. Reading the returned value alongside the timings helps distinguish a change in speed from a change in arithmetic order.

## A recorded example

An included capture used 64 float32 values with large positive and negative
terms that cancel. Its exact mathematical total was 1.

| Method | Returned total | Two repeats, ns per whole sum |
| --- | ---: | ---: |
| Ordinary sum | 0 | 100, 100 |
| Vectorized sum with one accumulator | 1 | 166.67, 166.67 |

The returned totals differ because the methods add the values in a
different order. The timing comes from one host with eight float32 lanes
and only three iterations per repeat. Treat it as an example of reading
a result. [The saved measurement capture](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/evidence/sum-cancellation.measurements.json)

## Try it

In **Explore**, select **Float32 sum**. Compare one vector accumulator with the scalar loop. Try a batch near your lane width and another large batch. The [fixture options](03-Command-registry#fixture-options) describe balanced, cancellation, and mixed-magnitude inputs.

See also [accumulator counts](16-Accumulator-count-experiments).
