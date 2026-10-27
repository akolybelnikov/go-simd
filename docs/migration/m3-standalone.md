# M3 — Standalone compiled reporting and consistent command use

Completed 2026-10-27. The default `simdreport` runs the current kernels directly
through compiled `testing.Benchmark`, without runtime Go tools or source paths.
Both commands can be built/installed and copied to a compatible remote machine.
The same flags apply through installed binaries, `go run` and Make.
Implementation: `33cc372` (`feat(engine): run reports as a standalone compiled CLI`).

## Boundaries

| Package / surface | Responsibility |
| --- | --- |
| `internal/catalog` and `kernels` | Existing descriptors and six typed kernel implementations; no arithmetic changes |
| `internal/engine` | Validated timing/configuration, Go-compatible case selection, prepared typed fixtures, backend interface and direct observations |
| `internal/evidence` | SIMD-independent versioned capture types and JSON writer |
| `engine.Testing` | Serialized documented testing flag API, calibration/counters, exact T/N and kernel-error propagation |
| `internal/platform` | Measuring executable's runtime/vector/build facts; no Go subprocess |
| `internal/cli` | Common group/alias/list definitions, group conflict validation, readable help and catalogue rendering |
| `cmd/simdreport` | Options, chosen backend, evidence lifecycle, summary/export orchestration |
| `cmd/simdbench` | Existing fixed-size quick measurements and display; shared CLI surface |

The pure catalogue, input, evidence, result and rendering packages still test with
GOEXPERIMENT unset. SIMD stays in executable kernel/platform paths. The engine
interface is internal; no public extension API or general workload platform is
introduced. Only the two current float32 operations/six implementations run.

## Measurement and identity

The adapter uses documented `testing.Init`, the registered `test.benchtime`
flag's `Value.Set`, and `testing.Benchmark`. It does not use unsafe or private
testing internals. A package mutex serializes initialization, setting, each
measurement and restoration of the caller's timing flag. Unrelated concurrent
testing.Benchmark callers in the same process are unsupported. CLI commands run
measurements serially; no worker process is added.

Positive Go durations and positive native-int `Nx` iteration counts are validated
before artifacts or fixtures. Go retains its GC and calibration policy; setup
can invoke a kernel beyond the final reported N. Count means independent final
measurements per case, not each calibration invocation. Actual N and duration T
are retained directly, and ns/op is `float64(T.Nanoseconds())/float64(N)`, without
integer NsPerOp or benchmark-text rounding. Allocation metrics likewise divide
total counters by N. No process CPU/RSS attribution is claimed.

The sweep/labels/variant order, fixture generator, scalar references, numerical
semantics and kernels remain unchanged. Typed fixtures are allocated/filled and
fingerprinted once per size, then reused across strategies/repetitions. Lookup,
binding, allocation and fingerprints happen before measurement. The callback
retains kernel validation/error checks, typed function-pointer invocation and
parallel addition's worker creation/join; a float32 sink consumes output once
after the repeated loop. No full checksum or extra warmup is added.

Unexpected kernel failures return adapter errors rather than `b.Fatal`. A failed
callback stops invoking the kernel; any remaining Go calibration callbacks do
no work and their results are discarded. Completed earlier observations are
retained. An injected backend-failure test covers this partial-result path.

Standalone results use backend `testing-benchmark/v1`, harness `standalone/v1`
as reserved by M0. Existing quick results retain `legacy-quick/v1`; developer
text retains `go-test-text/v1` / `legacy-report/v1`. Do not pool series across
these contexts. The default-backend change and precision change are explicit;
summary shape, median/ratio arithmetic, missing values and TXT/MD/CSV formatting
remain fixture protected.

## Selection and validation

The compiled selector matches registered `BenchmarkOperation/n=size/Label`
leaves. It implements Go's slash components and top-level alternatives while
respecting parentheses, character classes, escapes, whitespace rewriting and
unanchored regex. Empty expressions select nothing; paths below registered
leaves have no cases. It never substitutes substring matching or source fallback.
Differential tests compare native Go's actual selected cases with compiled plans
for both groups, anchored alternatives, a slash inside a class and a too-deep
path. Native `-cpu` and arbitrary testing flags remain developer controls, not
project CLI flags.

Both commands resolve canonical groups, add/compute aliases, deprecated function
aliases and boolean shortcuts through one definition. Matching selectors succeed;
conflicts and positionals fail consistently. Help now shares usage/options/example
structure. Both support identical `-list`; simdreport also validates its report
options before listing. Listing still creates no fixtures or artifacts.

Invalid configuration is rejected before artifact creation for either backend.
An empty selection retains failed evidence and returns no successful group
summary. With both groups, a completed first group remains if the second fails.

## Evidence and metadata

Compiled runs save `benchmark-<UTC>-<unique>.measurements.json` plus the chosen
TXT/MD/CSV summary. Schema `go-simd.measurements/v1` is a minimal versioned
capture: status/error, group/configuration, machine context, typed run and
individual observations. The existing model's Go field names are preserved
under `run`; top-level, configuration and machine fields have explicit JSON
names. It is not a worker manifest. The [Wiki schema guide](../wiki/05-Reports-and-evidence.md)
documents the fields and inspection examples.

Runtime facts include OS/architecture, Go version, effective GOMAXPROCS, vector
bits/lanes and actual emulation state. Embedded `debug.ReadBuildInfo` supplies
build settings/module/VCS information when available. Source revision and dirty
state remain explicitly unknown when absent, rather than calling Git or Go at
runtime. CPU model is currently unknown; no inherited environment string is
misrepresented as detected CPU hardware. The measuring process supplies all
compiled metadata. Input SHA-256, generator, actual workers, case identities,
backend/harness, repetitions, duration/iterations and available counters survive
serialization. Display rounding happens only in the summary renderer.

Explicit `-backend go-test` retains Go environment capture, source-relative native
execution and `.raw` text for development comparisons. Existing raw files and
the legacy importer are preserved. This path intentionally needs Go and source;
its failure never invokes the compiled backend. No existing `.raw` extension is
silently reused for JSON. I/O failures can leave incomplete files; atomic bundles
and abrupt-termination recovery remain outside M3.

## Usability and installation

`make help`, `make list`, `make build` and `make install` expose the current
workflows. All run/list/report targets forward ARGS, including group shortcuts.
Report controls remain COUNT/BENCHTIME/ARTIFACT_ROOT; BIN_DIR controls build
output. Build and install compile both existing commands. Installation uses Go's
standard GOBIN/workspace-bin rules. No package manager or prebuilt release assets
are added.

README and Wiki start with a compact task/installed/Go/Make matrix, then common
flags, report-only flags, bounded examples and a complete developer reference.
The Wiki covers cross-compilation, copying compatible binaries and SSH execution.
Reports remain on the execution machine. A target binary does not require Go,
Make, source or GOEXPERIMENT at runtime. This supports running the lab on a VM;
automatic provisioning, remote orchestration and cloud cost attribution are not
part of the CLI.

## Verification and stopping point

- Full tests/vet and existing numerical/format/identity fixtures.
- Adapter duration/Nx modes, timing-flag restoration and concurrent serialization;
  direct fractional timing/counters and failure retention.
- Both groups and all formats produce summaries consistent with their own direct
  evidence; explicit developer backend still produces parsable legacy raw text.
- Native/compiled selector differential coverage; unknown/invalid options fail
  before artifacts, and empty selections retain failed captures.
- A test builds the actual binary, runs it from an empty temporary directory with
  PATH and GOEXPERIMENT cleared, lists without artifacts, then measures both
  groups with exact 2x iterations and writes reports. Explicit developer execution
  fails there, proving no silent fallback. This test runs in Windows/Linux CI.
- Make help/list/build, shortcut option forwarding and install into temporary
  GOBIN; isolated installed report smoke checks.
- Complete Wiki/source link, flag/Make coverage, table/fence and whitespace checks.
- Linux CI race coverage; Windows/amd64 and Linux/amd64 execution evidence only.

See [M3 retained smoke evidence](baselines/m3-README.md) for actual direct and
developer captures. Timings are observational, not regression thresholds or proof
of backend equivalence. Stop after M3: worker isolation, cancellation, manifest
bundles and broader command migration are later conditional work; new F004
operation patterns remain a separate backlog.
