# Experiment recipes

Start each comparison with a question about one operation, one piece of work, or one measuring condition. Keep the other settings the same and save the repeats.

## Study vector boundaries

Read the measured lane count for the element type. If the width is eight, use **Explore** to preview 7, 8, 9, 15, 16, and 17 values in separate runs. Compare partial vectors and full vectors with a scalar tail using the same scalar reference. The sizes near each boundary help show the cost of handling a remainder.

## Compare working sets

Measure addition at 1,024, 65,536, and 1,048,576 values. Then try the compute-heavy expression at the same sizes. Compare the change from each operation's own scalar reference. The expression performs more arithmetic over a similar amount of data, but the timings alone do not identify a bottleneck.

## Explore concurrency

At a large size, compare a serial addition method and the parallel vector method in separate runs. Use the same machine and keep its concurrency setting fixed first. Then repeat with a different setting if concurrency is your question. Parallel timing includes worker startup and coordination, so state that when describing a difference.

## Compare fixture patterns

Hold the batch size and thresholds or bounds fixed while changing the arrangement of values. Alternating, clustered, and shuffled inputs can expose different costs for a scalar branch. For clamp, the pattern describes the positions that get clipped. See the [fixture options](03-Command-registry#fixture-options).

## Check repeatability

Use several repeats and keep the individual timings. Short fixed-iteration runs are useful for checking the setup. Use duration-based repeats when comparing performance. If two runs disagree, preserve both and check whether the data, settings, host, or background activity changed. [Understanding results](06-Understanding-results) explains how to report the spread.
