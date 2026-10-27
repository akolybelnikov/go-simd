# 11 — Glossary and capability ledger

Reviewed against M1 / Feature 1 on 2026-10-27. This is a ledger of implemented
capabilities, not a promise to deliver every possible extension.

## Capability ledger

| Capability | Current state |
| --- | --- |
| Input type | `float32` slices |
| Addition | Scalar, partial-vector, full-vector/scalar-tail, and parallel SIMD strategies |
| Compute expression | Scalar and SIMD `a*a + b*b + a*b` |
| Caller-supplied slices | Supported by module-local kernel calls |
| CLI datasets | Deterministic generated fixtures only |
| Input CSV/JSON inference | Not implemented |
| Other numeric types / comprehensive operation catalog | Not implemented |
| MulAdd/FMA mode | Not implemented; numerical semantics need an explicit contract |
| Quick comparison | Fixed 1,048,576 elements; no saved report |
| Multi-size report | Powers of two and vector-boundary sizes; configurable selection/count/benchtime |
| Saved summaries | TXT, Markdown, CSV; terminal stays readable text |
| Retained observations | Legacy raw Go benchmark text plus environment context |
| Typed internal model | Case/run/environment identity, repetitions, iterations, optional counters, provenance |
| Public structured raw schema | Not published; internal identities are not all exported |
| Statistics | Median summaries and ratios of medians; no inference/significance verdict |
| Concurrency control | GOMAXPROCS environment or native Go `-cpu`; no project worker flag |
| Parallel scalar control | Not implemented |
| Platform execution evidence | Current Windows/amd64 and Linux/amd64 CI; other targets require execution evidence |
| ISA/emulation claims | Experimental; vector API success alone does not establish native acceleration, and raw metadata has no explicit emulation status |
| Installed standalone report runner | Not implemented; report still invokes Go from the checkout |
| HTML/TUI/advisor/cost model | Not implemented |
| Process CPU / RSS / resource attribution | Not implemented; Go allocation counters are different metrics |
| Cloud provisioning / automatic spend | Not implemented |
| Offline saved-run analysis or re-render command | Not implemented |

M2 is shared runtime catalog work; M3 is the standalone benchmark backend.
Their migration order does not change current command availability. Broader
workers/resources/frontends remain conditional on observed laboratory needs.

## Glossary

| Term | Meaning in this lab |
| --- | --- |
| Scalar | The ordinary per-element source-loop reference |
| SIMD | Single instruction, multiple data; explicit vector operations in these kernels |
| Lane | One element position in the vector representation |
| Vector width | Lane count times element bits; eight float32 lanes would be 256 bits |
| Partial vector | A vector operation whose load/store handles fewer valid elements than a full vector |
| Tail | Remaining elements after full-vector processing; the scalar-tail strategy uses a scalar loop for these |
| Payload size | Number of elements in each input/output slice, not the number of bytes |
| Working set | Buffers used by an experiment; three float32 slices occupy about `12 * elements` bytes |
| Kernel | The implemented numerical function called inside measurement |
| Iteration | One repeated kernel call within a reported benchmark measurement |
| Repetition / sample | A separately reported measurement for the same complete case; not every calibration call |
| Calibration | Go's work to choose/execute a measurement schedule |
| Median | Middle repeated timing, or mean of the middle pair for an even count |
| Time change | `(variant median - scalar median) / scalar median * 100` |
| Speedup ratio | Scalar time divided by variant time; reciprocal of relative time |
| GOMAXPROCS | Go runtime bound on concurrent execution; used by this parallel kernel's worker policy |
| Goroutine | A Go runtime task; starting more of them does not create more physical cores |
| Case identity | Operation, type, strategy, size, semantics, parameters, input, workers, and execution mode |
| Backend | Measurement mechanism; currently legacy Go benchmark text for report import |
| Harness | The timed dispatch/setup/accounting contract around the kernel |
| Fixture fingerprint | SHA-256 over the declared input-bit encoding, computed outside measurement for fresh generated inputs |
| Emulation | Vector semantics implemented without the native acceleration a reader might assume; status must be observed, not guessed |
| FMA | Fused multiply-add, whose rounding semantics need deliberate comparison with unfused arithmetic |
| Throughput | Effective bytes per second from the benchmark's declared work accounting |
| Raw companion | The `.raw` file retaining individual Go benchmark observations and tool output |

Return to [Home](https://github.com/akolybelnikov/go-simd/wiki), or use the
[command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry)
to choose an implemented workflow.
