# M2 observational comparison

Captured 2026-10-27 on Windows/amd64, AMD Ryzen 7 7435HS, Go 1.27.3,
`GOEXPERIMENT=simd`, GOMAXPROCS 16, 256-bit vectors / 8 float32 lanes,
`simd.Emulated() == false`. The quick binary supplies the live SIMD context;
the native files retain their own OS/architecture/CPU headers.

- [Before](m2-before.txt): clean tracked source at
  `542997f8db1f679e8361595a2063579584813ff1`, before M2 edits.
- [After](m2-after.txt): M2 catalogue/registration implementation subsequently
  committed as `0c628fd`, based on that revision. Tracked implementation changes were present;
  documentation was still being completed. Kernel bodies were unchanged.

Both files retain complete command stdout: 90 native observations (five sizes,
six variants, three repetitions) plus all six quick-comparison results. Lines
were normalized to LF and trailing whitespace removed; numeric values and
benchmark names were not changed. A preliminary after capture exceeded the
tool's output limit and was discarded; the archived after command was rerun with
a sufficient output budget. The pre-edit baseline was captured only once.

## Reproduce

From the repository root in PowerShell:

```powershell
$env:GOEXPERIMENT = 'simd'
go version
go test ./internal/addition ./internal/computation -run='^$' -bench='^Benchmark(Addition|Computation)$/^n=(1|8|16|1024|1048576)$' -benchmem -count=3 -benchtime=50ms
go run ./cmd/simdbench
```

Native size selection intentionally uses universally registered sizes. Raw
throughput is descriptor-derived logical bytes per second, not measured DRAM
bandwidth. Quick comparisons retain their existing 1,048,576-element input and
Go's normal `testing.Benchmark` calibration rather than the native 50ms setting.

## Interpretation

Coverage matches: all 30 selected native cases appear three times on each side,
and all six quick variants appear in their existing order. Serial native cases
retain zero B/op and zero allocs/op. Parallel addition retains its per-call
worker allocations; allocator/scheduler bytes can vary. Current traffic remains
`12*n` for both operations.

These short, sequential captures ran during development, without exclusive CPU
control. They show material variability, not performance equivalence: selected
native medians range from roughly -44% to +60%, with addition scalar at n=1024
about +38%. Quick results also differ. They are evidence of measurement/coverage
continuity, not proof of a regression, improvement, or unchanged optimization.
Do not use their timings as test thresholds or pool them into one experiment.
Kernel source and timed call/validation paths were separately reviewed; package
movement can affect compiler layout even when that boundary remains the same.

The local Windows/amd64 addition test binaries were also inspected with
`go tool objdump`: on both sides `benchutil.Measure` calls the benchmark callback
indirectly, and the active 256-bit callback calls the typed kernel indirectly.
There is no added timed descriptor lookup or dispatch layer. Instruction layout
and captured-field offsets differ, so this check does not assert identical
machine code or explain all timing changes. The baseline binary was built from
a temporary archive of the clean pre-edit commit; the source checkout was not
reset or replaced. Reproduce the call-site inspection on each revision with:

```powershell
go test -c -o addition-check.exe ./internal/addition
go tool objdump -s 'internal/benchutil.Measure$' addition-check.exe
# Before M2:
go tool objdump -s 'addition_test.BenchmarkAddition@simd256.func1.1.1$' addition-check.exe
# After M2 (adjust the specialization to the vector width being inspected):
go tool objdump -s 'benchutil.BenchmarkOperation@simd256.func1.1.1$' addition-check.exe
```

See the [M2 record](../m2-runtime-catalog.md) for correctness, compatibility,
capability-listing checks, and the milestone stopping point.
