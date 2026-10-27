# M2 — Shared runtime catalogue

Completed 2026-10-27. This milestone registers the existing two float32
operations and six implementations in ordinary Go source. It introduces
`simdbench -list` and stops before M3's standalone measurement backend.
Implementation commit: `0c628fd` (`feat(catalog): share existing kernel
descriptors and runtime listing`).

## Architecture and ownership

| Boundary | Responsibility | Consumers |
| --- | --- | --- |
| `internal/catalog` | Operation/type/strategy IDs, historical labels/order, arithmetic semantics, operand/output shapes, scalar parameter description, worker policy, logical bytes | Quick listing, executable bindings, native registration, report groups, legacy importer/comparison ordering |
| `internal/catalog/kernels` | Typed float32 calls and prepared-buffer closures for six existing kernels | Quick CLI and native benchmark helper |
| `internal/benchutil` | Deterministic fixtures, vector-boundary size sweep, native benchmark registration/measurement | `BenchmarkAddition` and `BenchmarkComputation` |
| `internal/result` / `internal/report` | Existing observation, summary, comparison and rendering contracts | Source-based report runner |

Metadata imports neither SIMD nor `testing`. Returned descriptors own their
slices so callers cannot mutate global registration. SIMD dependencies remain
in executable bindings and measurement code. There is no public registration API,
dynamic plugin system, implicit type conversion, or automatic backend fallback.

The catalogue has one typed binding entry per operation/strategy. Each entry
supplies a native function pointer and an explicit quick-CLI closure calling the
same kernel. The explicit closure preserves the quick path's direct kernel call
inside its existing bound callback. Using a generic function-pointer binder would
add another indirect dispatch there. Tests check exact native function bindings,
metadata order, missing/orphan entries, and both execution paths' results.

## Registered coverage and compatibility

| Operation ID | Type / arithmetic ID | Legacy benchmark | Labels / stable strategy IDs |
| --- | --- | --- | --- |
| `add` | `float32` / `go-float32-add-v1` | `BenchmarkAddition` | Scalar / `scalar`; SIMDPartial / `simd-partial`; SIMDTail / `simd-scalar-tail`; SIMDParallel / `simd-parallel-scalar-tail` |
| `compute-expression` | `float32` / `go-float32-quadratic-v1` | `BenchmarkComputation` | Scalar / `scalar`; SIMD / `simd-partial` |

Existing labels, benchmark names, group/variant order, generated inputs, sizes,
scalar references, expression semantics, aliases, report table/export formats,
and M0 identity contracts are retained. Unknown historical names retain their
legacy namespace; recognizing a historical label does not make it executable.
The importer gets known identities and worker policies from catalogue metadata.

All implementations still accept two `[]float32` inputs and one `[]float32`
output. Kernels in `internal/addition/add.go` and
`internal/computation/compute.go` are unchanged. Prepared closures retain typed
buffers; existing sinks retain a float32 result after the measured loop. Buffer
allocation/filling, descriptor resolution, binding and byte accounting are
outside timing. Each kernel's length validation, the existing callback/error
check, and parallel addition's per-call goroutine creation/join remain measured.
Native benchmarks keep their callback followed by typed function-pointer call.
No harness-version change is needed for these unchanged measurement boundaries.

## Work accounting and extension seams

`Operation.LogicalBytes(n)` derives effective bytes per call from declared
inputs/output. Array values count `n` elements; scalar values count one. It
rejects negative sizes, invalid shapes/widths and int64 overflow. Both current
operations therefore retain `12*n` bytes: two 4-byte reads and one 4-byte write.
This is logical work accounting, not hardware memory traffic or bandwidth.

Synthetic tests cover unary, binary, ternary, predicate and scalar-output models.
Scalar parameter metadata describes preparation/identity; parameters are excluded
from array traffic. These are descriptor seams, not executable parameterized
kernels or a generalized output-retention implementation. New types, mask
representations, scalar-output bindings, reduction order and parameter identity
still require explicit work before F004 kernels can use them.

SIMD variants declare portable SIMD use. Serial implementations have one worker;
parallel addition creates `min(n, GOMAXPROCS)` workers per nonempty call and none
for an empty input. Unknown imported GOMAXPROCS remains unknown. No worker pool,
ISA-specific kernel, or new capability-dependent dispatch is added.

## Runtime listing

```powershell
$env:GOEXPERIMENT = 'simd'
go run ./cmd/simdbench -list
go run ./cmd/simdbench -list -group addition
go run ./cmd/simdbench -list -function compute
```

Listing includes this process's Go version, OS/architecture, vector bit width,
float32 lane count, `simd.Emulated()` state and GOMAXPROCS, then registered
operation/type pairs, expressions, semantics, operands, logical traffic, labels,
strategy IDs and worker policies. Existing selectors and conflict rules apply.
`-list=false` retains the normal quick comparison. This is human-readable
diagnostic output, not a versioned machine-readable schema or CPUID inventory.
An emulation flag does not prove acceleration or guarantee a speedup.

The listing branch runs before fixture allocation/binding/calibration. It does
not execute kernels, invoke Go subprocesses, or create report artifacts. A built
quick binary was checked outside the checkout with PATH empty. This verifies
listing only; the report command still needs Go and source until M3.

## Verification and observational evidence

- Full tests and vet with `GOEXPERIMENT=simd`, including real TXT/MD/CSV report
  exports, legacy identity/format fixtures and existing numerical kernel tests.
- Exact six-binding inventory; native and bound calls at empty, small, vector
  boundary, tail and 1025-element sizes; disjoint and exact in-place outputs;
  length failures preserve output.
- Frozen descriptor identities/order, defensive copies, traffic/overflow and
  worker policies; CLI listing filters, aliases, runtime context and writer
  failures. Listing does not contain measurements or proposed operations.
- Metadata/input/result/report tests pass with `GOEXPERIMENT` unset, confirming
  the SIMD-independent metadata boundary.
- Native before/after observations and quick comparisons are retained under
  [baselines](baselines/m2-README.md), with reproduction commands. They check
  benchmark coverage and allocation patterns; timing is observational and noisy,
  with no performance thresholds or equivalence claim. Source movement may
  affect compiler layout even with the same call/validation boundary.
- Local addition test-binary disassembly confirms the original native callback
  and indirect typed-kernel call depth on both sides; reproduction and limits
  are included with the baseline evidence.
- Linux and Windows CI run the repository checks; Linux additionally covers
  race detection. CI outcome is available on the pushed commits.

## Checkpoint and next work

M2 is complete. M3 remains pending: structured compiled measurement, validated
duration/iteration/count controls and report execution without runtime Go/source.
New broadcast, predicate/mask, reduction and movement kernels remain the separate
[F004 backlog](../features/F004-operation-patterns.md). Existing article credit
and proposed-pattern distinctions remain in the Wiki.
