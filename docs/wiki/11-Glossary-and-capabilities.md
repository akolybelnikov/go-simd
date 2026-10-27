# 11 — Glossary and capability ledger

Reviewed against M3 / F004 slice 1 on 2026-10-27. This is a ledger of implemented
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
| Broadcast scale/affine kernels | Implemented, Scalar/SIMDPartial, exact float32 scalar parameters; broadcasts per call |
| Predicate/mask/selection kernels | Implemented threshold, selection and rangefilter; strict comparisons, bit-preserving selection, Masked/IfElse/And |
| Sum/dot reductions | Planned composed kernels; scalar-result and numerical-order contracts pending |
| Permutation/horizontal primitive studies | Separate, gated `archsimd` extension; not portable methods assumed by the lab |
| Quick comparison | Fixed 1,048,576 elements; no saved report |
| Shared runtime catalogue | Seven operations / sixteen implementations; shared descriptors, configured and legacy typed bindings |
| Runtime listing | Both CLIs share compact `-list`, detailed `-list -group`, aliases/help and parameter controls; no fixtures/benchmarks |
| Multi-size report | Powers of two and vector-boundary sizes; configurable selection/count/benchtime |
| Saved summaries | TXT, Markdown, CSV; terminal stays readable text |
| Retained observations | Default direct exact-duration/iteration JSON evidence; explicit developer backend retains legacy text |
| Typed internal model | Case/run/environment identity, repetitions, iterations, optional counters, provenance |
| Structured evidence schema | `go-simd.measurements/v1` in `.measurements.json`; complete run/observations and runtime/build facts; worker manifest deferred |
| Statistics | Median summaries and ratios of medians; no inference/significance verdict |
| Concurrency control | GOMAXPROCS environment or native Go `-cpu`; no project worker flag |
| Parallel scalar control | Not implemented |
| Platform execution evidence | Current Windows/amd64 and Linux/amd64 CI; other targets require execution evidence |
| ISA/emulation claims | Both listing and compiled evidence retain live emulation/vector context; this does not establish acceleration |
| Installed standalone report runner | Implemented; default compiled backend needs no Go/source; `go install` or copy a compatible built binary |
| CPU/build metadata | CPU model currently unknown; Go/VCS/build facts come from the executable, with missing VCS explicitly unknown |
| HTML/TUI/advisor/cost model | Not implemented |
| Process CPU / RSS / resource attribution | Not implemented; Go allocation counters are different metrics |
| Cloud provisioning / automatic spend | Not implemented |
| Offline saved-run analysis or re-render command | Not implemented |

M0-M3 and F004's first broadcast/predicate slice are complete. Reductions, FMA,
clamp, mask export, conversions and new types remain planned.
Worker isolation, cancellation recovery and advanced manifests remain deferred. Broader
workers/resources/frontends remain conditional on observed laboratory needs.

## Glossary

| Term | Meaning in this lab |
| --- | --- |
| Scalar | The ordinary per-element source-loop reference |
| SIMD | Single instruction, multiple data; explicit vector operations in these kernels |
| Lane | One element position in the vector representation |
| Vector width | Lane count times element bits; eight float32 lanes would be 256 bits |
| Partial vector | A vector operation whose load/store handles fewer valid elements than a full vector |
| Broadcast | Filling vector lanes from one scalar parameter |
| Mask | Logical per-lane predicate results; exported representation must be declared |
| Selection / blend | Choosing one of two values per lane according to a mask |
| Reduction | Combining multiple input values into fewer results, often one scalar |
| Permutation | Rearranging lane positions according to an index contract |
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
| Backend | Measurement mechanism; compiled testing.Benchmark by default, or explicit go-test developer comparison |
| Harness | The timed dispatch/setup/accounting contract around the kernel |
| Fixture fingerprint | SHA-256 over the declared input-bit encoding, computed outside measurement for fresh generated inputs |
| Emulation | Vector semantics implemented without the native acceleration a reader might assume; status must be observed, not guessed |
| FMA | Fused multiply-add, whose rounding semantics need deliberate comparison with unfused arithmetic |
| Throughput | Effective bytes per second from the benchmark's declared work accounting |
| Evidence companion | `.measurements.json` for compiled observations, or `.raw` for explicit developer/native Go text |

Return to [Home](https://github.com/akolybelnikov/go-simd/wiki), or use the
[command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry)
to choose an implemented workflow.
The [pattern review and credited reading](https://github.com/akolybelnikov/go-simd/wiki/12-SIMD-patterns-and-reading)
explain the F004 backlog without presenting proposed patterns as implemented.
