# M3 smoke evidence

Captured 2026-10-27 during M3 implementation, based on
`471b9faad1f72bfcfb10d0cb738e9a05d4c3c794` plus the uncommitted M3 changes.
Machine: Windows/amd64, Ryzen 7 7435HS, Go 1.27.3, GOEXPERIMENT=simd,
GOMAXPROCS 16, 256-bit SIMD / 8 float32 lanes, emulation false. CPU model in the
compiled metadata is explicitly unknown because model detection is unsupported.
The SDK/native child separately reports the model in developer text.

| Evidence | Contents |
| --- | --- |
| `m3-compiled-addition.measurements.json` | Eight direct observations, four strategies at n=8, two repetitions |
| `m3-compiled-computation.measurements.json` | Four direct observations, two strategies at n=8, two repetitions |
| `m3-developer-addition.raw` | Same selected native cases/repetitions through explicit developer backend |
| `m3-developer-computation.raw` | Same selected native cases/repetitions through explicit developer backend |

All final measurements use 2x mode. Direct evidence retains actual duration/N,
input fingerprints, case/worker identity, runtime/build context and counters.
Go-run build information in this capture has unknown VCS facts; the known source
base and dirty M3 worktree are recorded here rather than invented inside the
capture. The evidence writer was subsequently extracted into a pure package;
the schema and measured kernel/callback paths are the same. No summary files
are duplicated here: export tests compare every saved format against the same
observations and fixture-protected renderer.

## Reproduction

From the checkout in PowerShell:

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdreport -bench='^Benchmark(Addition|Computation)$/^n=8$' -count=2 -benchtime=2x -format=md -output=artifacts/m3-compiled
go run ./cmd/simdreport -backend=go-test -bench='^Benchmark(Addition|Computation)$/^n=8$' -count=2 -benchtime=2x -format=md -output=artifacts/m3-developer
```

Published copies use shortened names; original embedded run IDs, timestamps,
metadata and values are retained. Line endings are normalized to LF and trailing
whitespace trimmed. Original generated artifacts remain ignored under artifacts/.

These are plumbing checks, not performance conclusions. At tiny sizes and 2x,
timer resolution, setup/calibration and scheduling dominate. Differences between
backends are not regressions/equivalence verdicts and must not become test timing
thresholds or pooled samples. Backend/harness identities deliberately differ.

## Standalone acceptance

Automated tests build the actual executable, run from a fresh empty temporary
directory with PATH and GOEXPERIMENT cleared, list without artifacts, then run
both groups at sizes 8/16, count 2 and 2x mode with Markdown exports. They check
exact iterations, duration presence, runtime facts and matching files. The
explicit developer backend fails there instead of falling back.

`make build`, `make help`, `make list`, quick shortcut/help forwarding and a
Make report smoke were exercised. `make install` into a temporary GOBIN installed
both binaries; the installed report command also generated a two-repetition,
2x CSV report from a separate directory with PATH/GOEXPERIMENT empty. Full tests,
vet, pure-package tests without SIMD, and CI cover the remaining acceptance.

See the [M3 record](../m3-standalone.md) for precise scope and limitations.
