# M0 baseline evidence

Captured on **2026-10-27, Europe/Athens**, against source revision
`3ce769ca33eec166aa389c33c1de1ed315deca8a`. The tracked worktree was clean when
native capture started. Later changes are tests/documentation; measured production
source and benchmark loops remain those of this revision.

## Files

| File | Purpose |
| --- | --- |
| `2026-10-27-windows-amd64.native.txt` | Complete native output with provenance/environment; 162 measured rows |
| `addition.smoke.raw` / `addition.smoke.txt` | Actual report output and raw evidence for addition |
| `computation.smoke.raw` / `computation.smoke.txt` | Actual report output and raw evidence for computation |

Native output includes outliers. Line endings are normalized to LF and trailing
whitespace trimmed; measurement values are unchanged. Smoke filenames are
shortened; embedded timestamps, commands, metadata, tables, and measurements are
retained. Original generated files remain in the temporary capture directory.

The machine had no controlled affinity, cache, frequency, thermal, or isolation
policy. These observations do not establish crossover guarantees or regression
thresholds. `1x` records verify report/artifact plumbing, not performance.
Synthetic expected tables under command `testdata` are deterministic fixtures.

## Environment

- Windows/amd64, AMD Ryzen 7 7435HS.
- Go 1.27.3, `GOEXPERIMENT=simd`; runtime reports `go1.27.3-X:simd`.
- `GOAMD64=v1`, `GOARM64` unset, `GOTOOLCHAIN=auto`.
- Effective GOMAXPROCS: 16 (native suffix and report metadata).
- Eight `float32` lanes (256 bits), observed by the report runner.
  Emulation status was not captured and is unknown.
- Parallel workers: `min(16, elements)` from the unchanged kernel policy.
- Inputs: `generated-float32-v1`; formula and future hash encoding in the
  [M0 contract](../m0-contract.md).
- Legacy artifacts have no input fingerprint or build revision field; revision
  was recorded separately. Rounded ns/op cannot recover exact total duration.

## Reproduce native capture

Run from the repository root in PowerShell:

```powershell
$env:GOEXPERIMENT = 'simd'
go test ./internal/addition ./internal/computation -run='^$' -bench='^Benchmark(Addition|Computation)$/^n=(1|7|8|9|15|16|17|1024|1048576)$' -benchmem -count=3 -benchtime=100ms
```

Sizes cover tiny input, one/two-vector boundaries on this machine, a medium case,
and the quick CLI's large case. This is a selected baseline. Use `-bench='.'`
for the full existing sweep; expect a longer run and larger working sets.

## Reproduce report smoke

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -group all -bench '^Benchmark(Addition|Computation)$/^n=(8|9)$/^(Scalar|SIMD.*)$' -count 2 -benchtime 1x -output artifacts/m0-smoke
```

The output directory differs from tracked evidence intentionally. Preserve
historical captures instead of replacing them with new timings.

## Verification

Formatting, uncached `go test ./...`, and `go vet ./...` passed. Attempted
`go test -race ./...` failed to build `runtime/cgo`: local `cgo.exe` exited 2
with MSYS2 UCRT GCC 16.1.0, before tests ran, including the unchanged addition
package. Linux CI already runs races. M0 does not claim a successful local race
run or change that CI configuration.
