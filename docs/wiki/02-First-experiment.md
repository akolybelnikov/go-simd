# Your first addition comparison

[Install the CLI](https://github.com/akolybelnikov/go-simd/blob/main/docs/cli.md#install-the-cli-from-source).

## Run directly

```text
simdlab run addition
```

This compares ordinary and vectorized addition on two generated arrays of
1,024 float32 values, with three repeats per method. The terminal displays
the settings, runs the comparison, and prints the timing difference and the
paths to the Markdown report and measurement file. This path works without
a model. Press Ctrl+C to cancel.

## Ask an agent

```text
simdlab ask
```

Enter **Compare ordinary and vectorized addition for 1,024 numbers.** Review
the methods, batch size, and repeats, then enter **yes** to measure or **no**
to decline. After the run, ask **Which took less time, and by how much?** or
**Were the repeats consistent?** The answer uses measurements saved during
the session. Enter **quit** to leave.

## Read and share the result

For example, with medians of 80 ns for vectorized addition and 100 ns for
ordinary addition, vectorized addition took 20% less time per batch. The
Markdown report includes the observed repeat ranges and measuring host.
Use `simdlab inspect -text` with the printed measurement path to see details.

## Custom comparison with three methods

Question: for a modest input size on this machine, how do serial partial-vector
addition and scalar-tail addition compare with the scalar loop?

Prerequisites: finish [getting started](https://github.com/akolybelnikov/go-simd/wiki/01-Getting-started).
For these `go run` examples use the repository root and `GOEXPERIMENT=simd`.
With an installed command, replace `go run ./cmd/simdreport` with `simdreport`;
all flags stay the same and neither Go nor the checkout is needed at runtime.

## Step 1: measure one size and three serial strategies

```powershell
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1024$/^(Scalar|SIMDPartial|SIMDTail)$' -count 5 -benchtime 100ms -format md
```

The expression selects exactly 1,024 elements and the scalar, partial-vector,
and scalar-tail variants. `100ms` is a per-measurement target handled by Go's
calibration, not the duration of the whole command. Five repetitions produce
five observations for each selected case.

This is a manageable first repeated experiment. Prefer a quiet machine and
repeat later if the pattern matters to a decision.

## Step 2: inspect the table and evidence

The terminal remains text even though the saved report is Markdown. Open the
paths printed by the command. The `.md` contains the table and environment
metadata; `.measurements.json` contains individual direct measurements, exact
duration/iterations, identities, runtime/build context and allocation counters.

An **illustrative**, invented row might be:

| Elements | Scalar ns/op | SIMDPartial ns/op | SIMDPartial change | SIMDTail ns/op | SIMDTail change |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1024 | 100.00 | 80.00 | -20.0% | 110.00 | +10.0% |

This means the medians were 100, 80, and 110 nanoseconds for a whole 1,024-element
kernel call. Partial-vector took 20% less time; scalar-tail took 10% more time.
The invented values are a reading exercise, not expected output.

## Step 3: expand the size comparison

```powershell
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=(128|1024|8192)$/^(Scalar|SIMDPartial|SIMDTail)$' -count 5 -benchtime 100ms -format csv
```

Now compare whether the pattern changes across 128, 1,024, and 8,192 elements.
CSV uses numeric changes (`-20.0`, not `-20.0%`) and empty fields for unavailable
values. Keep the raw companion when sharing or charting the CSV.

## Step 4: isolate the compute expression

```powershell
go run ./cmd/simdreport -group computation -bench '^BenchmarkComputation$/^n=(128|1024|8192)$/^(Scalar|SIMD)$' -count 5 -benchtime 100ms -format md
```

This measures `a*a + b*b + a*b`, not addition. Compare strategies within each
operation first. Different operations perform different work, even when they
use the same number of input elements.

## Step 5: write a bounded conclusion

Use a template such as:

> On [CPU / OS / Go version], with generated float32 inputs and [size], the
> median time change of [strategy] against scalar was [change]. The raw samples
> [did/did not] show a consistent pattern across repeated runs. This experiment
> includes kernel-call overhead and buffer reuse; it does not isolate a resource
> bottleneck or establish application-wide speedup.

Record source revision (`git rev-parse HEAD`) and local changes
(`git status --short`) with the report; the legacy raw format does not capture
complete build/source identity automatically.

Next: [command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry)
for exact options, or [understanding results](https://github.com/akolybelnikov/go-simd/wiki/06-Understanding-results)
for the measurement model.
