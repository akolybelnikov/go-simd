# 07 — Focused experiment recipes

All commands use the repository root and `GOEXPERIMENT=simd`. Keep raw companions
and record source revision/local changes. Examples use bounded selections so you
can learn one behavior before running the full sweep.

## A. Verify plumbing without drawing conclusions

```powershell
go run ./cmd/simdreport -group all -bench '^Benchmark(Addition|Computation)$/^n=(8|16)$/^(Scalar|SIMD.*)$' -count 1 -benchtime 1x -format md
```

Expect two pairs of raw/Markdown files, one per group. This checks selection,
execution, and saving. Both sizes are powers of two present in every current
sweep; they are not guaranteed vector boundaries on every environment. `1x`
timings are not performance evidence.

## B. Study vector boundaries

Read `SIMD float32 lanes` from a report. If it is **8**, compare:

```powershell
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=(7|8|9|15|16|17)$/^(Scalar|SIMDPartial|SIMDTail)$' -count 5 -benchtime 100ms -format md
```

For another width `w`, replace the sizes with `w-1`, `w`, `w+1`, `2w-1`, `2w`,
and `2w+1`, omitting nonpositive values. The automatic sweep includes these
boundaries. Ask whether partial-vector and scalar-tail patterns change locally;
do not label a crossover universal.

## C. Compare operations across larger working sets

```powershell
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=(1024|65536|1048576)$/^(Scalar|SIMDPartial|SIMDTail)$' -count 5 -benchtime 100ms -format csv
go run ./cmd/simdreport -group computation -bench '^BenchmarkComputation$/^n=(1024|65536|1048576)$/^(Scalar|SIMD)$' -count 5 -benchtime 100ms -format csv
```

Compare strategy changes within each operation. Discuss the different arithmetic
per effective byte, while keeping cache/memory explanations as hypotheses.

## D. Explore concurrency in separate report runs

PowerShell, preserving your previous environment setting:

```powershell
$previousProcs = $env:GOMAXPROCS
try {
    $env:GOMAXPROCS = '1'
    go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1048576$/^(Scalar|SIMDParallel)$' -count 5 -benchtime 100ms -format md -output artifacts/procs-1
    $env:GOMAXPROCS = '4'
    go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1048576$/^(Scalar|SIMDParallel)$' -count 5 -benchtime 100ms -format md -output artifacts/procs-4
} finally {
    $env:GOMAXPROCS = $previousProcs
}
```

Both report roots retain their own group directory. A larger GOMAXPROCS bound
is not a promise of more dedicated physical cores. SIMDParallel still creates
goroutines at GOMAXPROCS=1 and retains its setup overhead. These comparisons
explore concurrency; they do not replace a parallel scalar control.

In a POSIX shell, use a per-command environment prefix to avoid changing the
parent shell:

```bash
GOMAXPROCS=1 go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1048576$/^(Scalar|SIMDParallel)$' -count 5 -benchtime 100ms -format md -output artifacts/procs-1
```

## E. Use native Go concurrency sweeps

```powershell
go test ./internal/addition -run='^$' -bench='^BenchmarkAddition$/^n=1048576$/^(Scalar|SIMDParallel)$' -benchmem -count=5 -benchtime=100ms -cpu=1,4
```

This produces native benchmark lines for each concurrency setting. `-cpu` is
not a `simdreport` flag. Keep different concurrency contexts separate during
analysis; the typed comparison model rejects mixed contexts in one table.

## F. Measure only one strategy

```powershell
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1024$/^SIMDPartial$' -count 5 -benchtime 100ms -format csv
```

This is valid when you need its timing. Percentage change stays empty because
there is no measured scalar baseline. Add `Scalar` to the selection if you need
a baseline ratio.

## G. Check iteration mode

```powershell
go run ./cmd/simdreport -group computation -bench '^BenchmarkComputation$/^n=1024$/^(Scalar|SIMD)$' -count 3 -benchtime 10x -format txt
```

Inspect the reported iteration count in raw output. Fixed iterations can check
runner behavior, but a short batch may be too brief for a reliable performance
decision. Go can execute calibration invocations in addition to the final
reported measurement.

## H. Keep a before/after experiment honest

1. Record the question, revision, local changes, toolchain, machine, and settings.
2. Run a selected repeated comparison into `artifacts/before`.
3. Make the single intended change and record its revision/diff.
4. Run the same selection/configuration into `artifacts/after`.
5. Compare raw variability and summaries without silently pooling contexts.
6. Preserve regressions, missing evidence, and the scope of the conclusion.

The tool currently has no saved-run comparison command, statistical verdict, or
automatic regression threshold. Do not replace that missing analysis with a
claim that every small median change is meaningful.
