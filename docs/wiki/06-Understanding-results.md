# 06 — Understanding results

The unit of work is one call over the complete selected slice. A measurement
for `n=1024` is not the latency of one float addition.

## Read each number correctly

| Field | Meaning | Common mistake |
| --- | --- | --- |
| `ELEMENTS` | Values per input/output slice | Treating it as bytes |
| `ns/op` | Nanoseconds per complete kernel call, averaged over Go's reported iterations | Treating batch-average timing as a request tail percentile |
| `change` | Time change relative to the measured scalar median | Calling it percent speedup without checking the formula |
| `B/op` | Allocated bytes per operation | Treating it as total memory traffic or peak resident memory |
| `allocs/op` | Allocations per operation | Assuming lower allocation alone proves lower latency |
| `MB/s` | Effective byte rate based on `12 * elements` | Treating it as measured DRAM bandwidth |
| Native `-16` suffix | Go concurrency annotation | Reading it as vector width or worker count for every strategy |

## Median and percentage arithmetic

Each case's repeated ns/op observations are sorted. With an odd count, the
middle value is the median; with an even count, the two middle values are
averaged. The report then computes:

```text
change (%) = (median_variant - median_scalar) / median_scalar * 100
```

For invented scalar/variant medians of 100 and 80 ns/op:

- Time change is `-20.0%`: the variant takes 20% less time.
- Speedup ratio is `100 / 80 = 1.25x`.
- The reciprocal operation-rate increase is 25%, not 20%.

For 100 and 120 ns/op, time change is `+20.0%`, meaning more time was required.
Missing or nonpositive scalar baselines produce unavailable changes. The
calculation uses the unrounded medians, not the displayed two-decimal cells.

These are ratios of per-case medians, not medians of pairwise percentages.
There are currently no confidence intervals, significance verdicts, or automated
"meaningful improvement" labels. Keep raw variability visible when judging a
small difference.

## Why tiny inputs can lose

Length validation, dispatch, vector setup, partial loads/stores, and tail checks
can dominate very small slices. Parallel work also pays goroutine startup and
waiting costs. A tiny scalar baseline makes percentage changes numerically large.

The useful question is what happens across nearby sizes and repeated runs on
the actual toolchain and machine, not whether SIMD always wins.

## Vector boundaries and tails

With width `w`, compare `w-1`, `w`, `w+1`, `2w-1`, `2w`, and `2w+1`.
The lab automatically includes those sizes. They help distinguish full-vector
work from remainder handling. A width-eight result at 8 and 9 elements is a
local example; do not transfer those boundaries to a different vector width.

Use the report's recorded lane count and inspect the actual selected rows.
Do not infer native hardware vector acceleration solely because a vector API
call succeeds; lowering and emulation depend on the experimental environment.

## Serial versus parallel comparisons

`SIMDPartial` and `SIMDTail` are serial addition strategies. `SIMDParallel` adds
concurrency, chunk boundaries, allocation, startup, and synchronization.
Changing GOMAXPROCS changes this experiment.

The current scalar baseline is serial. A 2x parallel-SIMD gain against it is
not evidence of a 2x vectorization gain. A matched parallel scalar control is
needed to isolate that contribution and is not implemented yet.

## Data movement and arithmetic intensity

Addition performs little arithmetic per effective byte. The compute expression
performs more arithmetic over the same two-input/one-output shape. Different
curves are useful observations, but these tables alone do not prove that an
operation is CPU-bound, cache-bound, or memory-bandwidth-bound.

Inputs/output are reused. Cache residency is neither controlled nor directly
measured. Throughput can reflect cached data rather than main-memory traffic.
Use a deliberately controlled experiment and suitable independent measurements
before assigning a bottleneck cause.

## Repeatability checklist

1. Ask one concrete question and choose the relevant operation/strategies.
2. Retain scalar whenever a percentage comparison is needed.
3. Use repeated duration-based measurements after an initial `1x` smoke check.
4. Keep source, toolchain, CPU, concurrency, selected sizes, and fixture policy
   comparable; record the thing you deliberately changed.
5. Inspect raw spread and rerun on a quiet machine if the difference matters.
6. Preserve negative results and explain what the comparison does not isolate.

The lab measures its Go kernels. It does not measure an unrelated application's
end-to-end latency, another language runtime, production request percentiles,
CPU resource counters, or cloud billing savings.
