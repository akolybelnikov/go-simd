# Changelog

All notable changes are recorded here. Migration milestones are separate from
product roadmap phases; M0 is the architecture behavior-freeze milestone.

## Unreleased

### F004 extension — Exact bit reinterpretation round trip — 2026-11-04

- Add bitroundtrip / bit-roundtrip-f32 / BenchmarkBitRoundTrip with Scalar and
  SIMDPartial: nineteen operations/forty-nine implementations. Preserve every
  opaque float32 bit through Float32bits/frombits or ToBits/BitsToFloat32,
  including signed zeros and quiet/signaling NaN payloads; no numeric conversion.
- Time validation, width setup, bounded loads, representation path and stores
  with 8*n logical bytes. Retain all input/output bits, actual lanes and dedicated
  harnesses; keep copy-specific identity separate and historical unknowns unknown.
- Use fixed raw-bit fixtures with a distinct generator namespace and no relevant
  scalar/fixture flags. Propagate typed unary execution and complete output
  evidence through quick/native/compiled/agent/worker paths and standalone use.
- Test random/special bits, tails, aliases/canaries/unchanged input, errors and
  allocations, strict native facts and normal/emulated/backend/worker parity.
  Keep frozen legacy tests explicitly focused on their original two operations.
- Update README/registry, Wiki chapter27, example, contract, retained captures,
  SDK source-use ledger, CI and local tracker statuses/links. Explain compiler
  bitcast elimination: timings alone establish no emitted instruction or latency.
  Numeric conversions and further representation studies remain future work.

### Native shared-package group isolation — 2026-11-04

- Resolve broad native selectors to the selected group's registered leaves when
  multiple operations share a Go benchmark package. Keep the original selector
  in raw metadata and retain configured fixture propagation separately.
- Verify that addition/addition64 developer reports each contain only the chosen
  operation under a broad Benchmark.* selector. Kernel timing callbacks are unchanged.

### F004 extension — Uint8 saturating addition — 2026-11-04

- Add saturatingadd8 / add-sat-u8 / BenchmarkSaturatingAdd8 with Scalar and
  SIMDPartial: eighteen operations/forty-seven implementations. Compute the
  widened sum clipped at 255, with no modulo wrapping or fabricated float input.
- Add true byte-array bindings/fixtures/fingerprints through quick/compiled/
  native/agent/worker paths. Time validation, width setup, bounded loads,
  AddSaturated and stores with 3*n logical bytes; retain every output byte and
  actual Uint8 lanes. Keep Case.Predicate absent for numeric output.
- Import strict native width/digest facts under dedicated saturation harnesses;
  retain known generated input and preserve unknown historical contexts.
- Exhaust all 65,536 input pairs in normal/emulated modes. Test native-width tails,
  aliases/canaries/unchanged inputs, errors/allocations, typed bindings/hashes,
  backend/worker parity, no-Go standalone use and testing/report compatibility.
- Update SDK coverage, README/registry, Wiki chapter26, example, contract,
  retained normal/emulated observations, CI and local status/navigation trackers.
  Other widths, wrapping, conversion and additional integer operations remain
  separately scoped; each continuation operation has its own signed commit.

### F004 extension — True Float64 addition — 2026-11-04

- Add addition64 / add-f64 / BenchmarkAddition64 with Scalar/SIMDPartial:
  seventeen operations/forty-five implementations. Keep existing float32
  addition identities, fixtures and timed callbacks unchanged.
- Add real []float64 inputs/output and typed bindings, directly generated
  binary64 fixtures, all-input fingerprints and complete little-endian output
  digests. Include validation, width setup, bounded loads/addition/stores in
  timing with 24*n logical bytes; retain actual Float64 lanes in Case.Vector.
- Bound partial vectors to actual lanes to avoid the selected SDK's emulated
  Float64 loader over-reporting slice length. Test that mode explicitly and
  keep all bounding work timed. Preserve IEEE zero/subnormal/overflow behavior;
  NaN payloads remain unspecified.
- Make native registration resolve the descriptor's actual type, retain strict
  native width/digest evidence and worker policy for all known types, and carry
  true arrays/bytes/widths through quick/compiled/native/plan/worker paths.
- Test independent exact-rational/special/random bits, tails/canaries/aliases,
  errors/allocations, low binary64 fingerprint bits, adapter flag restoration,
  compiled/native/emulated/worker parity and no-Go standalone operation.
- Update README/registry, Wiki chapter25, example, contract, retained captures,
  SDK member/type status ledger, CI checks and local tracker status/navigation.
  Other Float64 operations, integer families and imported datasets remain separate.

### F004 extension — Three-array fused multiply-add — 2026-11-04

- Add muladd-f32 / muladd / BenchmarkMulAdd with Scalar software FMA32 and
  verified SIMDPartial vector/software paths: sixteen operations/forty-three
  implementations. Preserve one-rounding binary32 semantics, length errors before
  writes, valid empty input and exact aliases; NaN payloads remain unspecified.
- Extend typed bindings with a real third array; generate balanced/rounding/mixed
  triples and stream fingerprints across all operands. Keep older recipes/timing
  callbacks intact and explicitly reject ternary use of the two-input helper.
- Include validation/dispatch/loads/FMA/stores in timing, account for 16*n work,
  and retain complete output SHA-256 after timing with dedicated harnesses and
  existing FMA policy/gate/path/width identity. Strict native import preserves
  unknown historical input and rejects incomplete or conflicting metrics.
- Propagate fixture configuration through both backends, both quick/report CLIs,
  agent catalogue/plans and isolated workers. Update help, README, Wiki chapter24,
  kernel examples, feature/evidence records and local status/navigation trackers.
- Verify independent rational oracles, special/random bits, tails, exact aliases,
  canaries, errors and allocations; test all-input fingerprints, every fixture's
  native/compiled/emulated parity, worker propagation and no-Go standalone runs.
  Extend CI native/emulation coverage; keep follow-on operations separate.

### F004 slice 8 and closure — Ordered clamp — 2026-11-04

- Added clamp / clamp-f32 / BenchmarkClamp with Scalar branches, SIMDPartial
  comparisons/selection and SIMDMinMax with timed outside/tie/NaN bit repairs.
  The shared catalogue now contains fifteen operations / forty-one implementations.
- Define inclusive ordered bounds, exact clipping-bound bits, original equality/
  signed-zero/quiet/signaling-NaN payload bits, length-error precedence and
  invalid-bounds rejection before writes, including empty input. Support disjoint
  and exact alias; partial overlap is outside the contract.
- Reuse lower/upper/pattern/seed through both CLIs, installed binaries, Go and Make.
  Patterns select clipped positions; reject impossible distributions and NaN/
  reversed bounds before artifacts. Retain exact parameter/generator/input identity.
- Include validation, broadcasts, selection/MinMax repairs and stores in timing;
  account for 8*n logical bytes. Reuse complete post-timing output retention and
  ArrayResultSHA256, with dedicated clamp harnesses and Case.Clamp policy/width.
  Strict native digest/lanes import and comparison context reject incompatible facts.
- Test independent bitwise references, random/special values, equal/infinite/zero
  bounds, tails/canaries/aliases/errors/allocations; compiled/native parity covers
  every pattern and exact outputs, also in SDK emulation. Extend all-group no-Go
  standalone verification and Linux/Windows CI emulation/native checks.
- Add examples/clamp, retained normal/emulated captures, detailed clamp contract,
  README/registry updates and Wiki chapters 22/23. Audit F004 acceptance and account
  for all 381 selected Go 1.27.3 SDK declarations with explicit status/reasons and
  source hashes. Preserve earlier checkpoint history and Jesús Espino article credit.
- Close F004 for the agreed current scope. Three-array MulAdd, broader input types,
  reordering/conversion studies, M4 and HTML remain separately scoped future work.

### F004 slice 7 — Exact copy and movement controls — 2026-11-04

- Added copy / copy-f32 / BenchmarkCopy with Scalar source loop, GoCopy built-in
  control, SIMDPartial and SIMDTail: fourteen operations/thirty-eight implementations.
- Preserve all float32 bits, including signed zeros, subnormals, infinities and
  quiet/signaling NaN payloads. Require equal lengths before writes; empty/nil
  succeeds. Support disjoint storage/exact alias, with partial overlap outside
  the shared contract. Benchmarks use disjoint input/output.
- Include validation, movement, loop control and tails in timing. Retain the
  complete output after timing under dedicated copy callbacks/harnesses; original
  numerical/predicate callbacks retain their measurement boundaries.
- Add typed BindUnary preparation with one real input and output, no fabricated
  second input or parameters. Fixed generated-copy-float32-bits-v1 fixtures
  preserve raw special-value bits; other operation options leave them unchanged.
- Account for 8*n logical bytes; record Case.Copy's exact-bit policy and actual
  SDK lanes. Retain ArrayResultSHA256 over every little-endian float32 output bit
  after measurement. Import strict child SDK width and eight digest words; keep
  historical unknown input unknown and reject incompatible bit policies/widths.
- Keep Scalar as the source-loop baseline, copy-source-loop/v1, with GoCopy as an
  additional optimized control. Label compiler/runtime-selected behavior clearly;
  SDK emulation does not disable Go built-in-copy optimization.
- Verify random/raw special bits, offset/native/tail/large lengths, exact aliases,
  input/canary preservation, length errors before writes and zero allocations.
  Cover unary bindings, streamed fingerprints, irrelevant flags, strict native
  facts, partial captures and timing restoration.
- Verify four-label compiled/native identity/output parity at vector boundaries
  and larger sizes, including SDK emulation. Include GoCopy in all-format exports;
  keep compact-list validation at one row per registered operation.
- Pass full tests/vet, pure packages, no-Go/all-group standalone execution and
  Go/installed/Make examples. Extend Windows/Linux native copy smoke checks and
  Linux SDK-emulation parity. Add examples/copy, feature/evidence records and
  Wiki chapter 21 with consistent commands, controls and interpretation.
- Stop after copy. Clamp, reordering, three-array MulAdd, standalone conversions,
  other input types and architecture M4 remain separate work.

### F004 slice 6 — Packed predicate masks — 2026-11-03

- Added predicatepack / predicate-gt-packed-f32 / BenchmarkPredicatePacked
  with Scalar/SIMDPartial: thirteen operations/thirty-four implementations.
  Input i maps to bit i%8 in byte i/8, least-significant bit first.
- Require exactly ceil(n/8) output bytes; clear unused high bits and overwrite
  all bytes on every call. Ordered equality/NaNs produce false. Invalid length
  fails before writes, empty/nil succeeds and input storage remains disjoint.
- Scalar builds each byte; SIMD normalizes vector comparisons to explicit 0/1,
  converts/extracts and packs valid lanes, carrying partial bytes across vector
  boundaries. Validation, conversion, packing and stores are timed.
- Add PackedBits/OutputElements to distinguish logical input size from physical
  output storage with overflow-safe rounding. Logical work is 4*n+ceil(n/8).
  Preserve original array/scalar traffic and byte-mask kernel/callback bodies.
- Reuse typed byte output and strict complete output-digest retention, with
  uint8-packed-lsb0/v1 and dedicated native/compiled harnesses. Keep packed and
  byte-mask arithmetic/output identities and timing series distinct.
- Reuse threshold/pattern/seed recipes and exact input identity. Add sorted,
  deduplicated sizes 7,9,63,65 around packing boundaries to both backends;
  existing operations keep their prior sweeps.
- Test independent bitwise truth, exact LSB bytes, unused padding, stale-output
  overwrite, negative-threshold tails, NaNs/zeros/subnormals/infinities, random
  float32 bits, canaries, input preservation, errors and current zero allocations.
  Verify physical extents/work overflow, typed bindings and unchanged inputs.
- Verify all five distributions through compiled/native backends at byte/vector
  boundaries, exact counts and complete output hashes, including forced emulation.
  Pass partial-retention, all-group/no-Go standalone, pure-package, full tests/vet
  and checked Go/installed/Make examples.
- Extend Windows/Linux CI with packed-boundary native smoke checks and Linux
  emulated backend parity. Add examples/predicatepack, feature/evidence records,
  Wiki chapter 20 and consistent readable command/operation/capability guides.
  Make examples use n=64; direct anchored selectors cover exact n=65.
- Stop after packed export. Clamp, copy/reorder, three-array MulAdd, standalone
  conversions, other input types and architecture M4 remain separate work.

### F004 slice 5 — Canonical uint8 predicate export — 2026-11-03

- Added predicate / predicate-gt-f32 / BenchmarkPredicate with Scalar and
  SIMDPartial, bringing the catalogue to twelve operations/thirty-two
  implementations. Input remains float32; output is a real uint8 array.
- Define ordered strict x > threshold as canonical byte 1/0. Equality, either
  zero at threshold zero and NaNs produce 0. Length errors precede writes,
  equal empty/nil succeeds, input is preserved and storage must be disjoint.
- SIMD selects explicit float32 0/1, converts/extracts through int32 scratch
  and writes valid bytes only. Raw SDK mask representation never escapes;
  padded zero lanes cannot corrupt a negative-threshold tail.
- Add typed BindUint8/Prepared.ByteOutput preparation and separate compiled/
  native predicate-export harnesses. Include validation, broadcasts, comparison,
  normalization, conversion, extraction and stores in timing. Logical work is
  5*n bytes; preparation/retention/output hashing stay outside timing.
- Retain full ByteResultSHA256 per observation and Case.Predicate's canonical
  encoding/actual lanes. Reject mixed widths/encodings. Import strict child
  width and eight exact uint32 digest-word metrics; missing/duplicate/fractional/
  invalid words fail. Keep unrelated v1 fields omitted and existing callbacks.
- Reuse shared threshold/pattern/seed recipes and exact parameter/input identity;
  do not allocate dummy float32 output or introduce new CLI flags.
- Verify exact bytes over random raw float32 values, signed zeros, infinities,
  subnormals/NaNs, length errors, input preservation, output canaries, lane/tail
  boundaries and current zero allocations. Cover bindings, fixture identity,
  mixed-context rejection, partial captures and timing-flag restoration.
- Verify all five distributions through compiled/native backends at 1,w-1,w,w+1,
  exact counts and complete output hashes, including forced emulation. Full
  tests/vet, experiment-independent packages, standalone all-group/no-Go tests,
  Go/installed/Make examples and hardware/software captures pass.
- Extend Windows/Linux CI with native predicate smoke checks and Linux
  forced-emulation backend parity. Add examples/predicate, the detailed feature
  record, Wiki chapter 19 and synchronized human command/capability guides.
- Stop after byte-mask export. Packed masks, clamp, three-array MulAdd, movement,
  standalone conversions, other input types and architecture M4 stay pending.

### F004 slice 4 — Fused dot with specified folds — 2026-11-03

- Added dotfma / dot-fma-f32 / BenchmarkDotFMA with Scalar and explicitly
  unrolled SIMDStriped/2/4 strategies. Catalogue: eleven operations/thirty
  implementations. Existing sum, unfused dot and affine bodies/IDs stay intact.
- Fuse each binary32 multiply-accumulate using the corrected software reference
  or gated native MulAdd. Specify round-robin full-vector assignment,
  left-to-right ordinary rounded merges, low-to-high lane fold and software FMA
  tails. This does not round the complete mathematical dot only once.
- Preserve each striped order in software/emulated execution. Retain per-case
  reference, software-striped, vector-full, vector/software-tail or tail-only
  paths separately from the process gate, width, fold order and accumulator count.
- Use checked scalar bindings and dedicated compiled/native dot-FMA harnesses.
  Keep length validation, dispatch, updates, extraction/folds and tails timed;
  retain uint32 scalar result bits once afterward. No output slice; logical
  work is 8*n+4 bytes. Errors preserve earlier successful measurements.
- Add strict child fma-vector/fma-path/result-bits/simd-lanes validation;
  dot-specific path codes 0..4 do not change affine's 0/1 contract. Reject
  inconsistent size/gate/path facts and incompatible comparison contexts.
- Extend shared -dot-input with fusion-sensitive for dot and dotfma, retaining
  the three original recipes. The new fused generator has its own identity;
  both groups fingerprint identical x/y fixture values independently.
- Verify independent exact-rational per-update rounding and specified folds,
  finite forward bounds, random bit patterns, NaNs/infinities, signed underflow,
  overflow cancellation, nil/empty/mismatched inputs, overlap, input preservation,
  current zero allocations, unroll boundaries and wider software scratch.
- Check all four fixtures/strategies through compiled/native backends at
  tail-only, full-vector and hybrid sizes, with exact counts and identity parity;
  include failure retention and all-group/no-Go standalone tests.
- Pass full tests/vet, experiment-independent model packages, forced-software
  numerical/backend checks, Go/installed/Make commands and retained hardware/
  software captures. Extend Windows/Linux CI with fused-dot native smoke checks
  and Linux forced-emulation backend parity.
- Add examples/dotfma, detailed feature/evidence records and progressive Wiki
  chapter 18; align README, catalogue, shared flag values and capability ledger.
  Preserve the credited SIMD reading and prior checkpoint history.
- Stop after fused dot. Three-array MulAdd, other pattern families and M4 remain
  separate pending checkpoints.

### F004 slice 3 — Correctly rounded fused affine — 2026-11-03

- Added affinefma / affine-fma-f32 / BenchmarkAffineFMA with software Scalar
  and SIMDPartial under one binary32 round-to-nearest ties-to-even contract.
  Existing affine, computation, dot and reduction kernels remain unchanged.
- Added an ordinary-source software FMA32 primitive: exact binary64 product,
  error-free TwoSum and round-to-odd correction before narrowing. This avoids
  widened math.FMA double rounding at normal, subnormal and overflow boundaries;
  tests use an independent rational oracle and explicit special/signed-zero rules.
- Inspected the SDK's unfused emulated MulAdd. Added an out-of-timing startup
  gate for native fused rounding; emulation/failed probes use an explicitly
  recorded corrected software path. Scalar always uses that software reference.
  Listing and quick/report output expose actual vector/software execution.
- Added separate FMA harness IDs, machine.fma32_path, Case.FMA path/rounding/
  process-width facts and optional last-output uint32 bits to v1 evidence. Native
  metrics establish child facts, reject malformed/conflicting paths and avoid
  parent capability inference. Mixed vector paths/widths cannot pool.
- Shared -fma-input balanced/rounding/mixed-magnitude plus existing alpha/beta
  through installed/Go/Make launchers. One input plus output accounts for 8*n
  logical bytes; broadcasts are not extra arrays. Catalogue: ten operations,
  twenty-six implementations. Historical absent FMAInput normalizes to balanced.
- Preserved original native pattern callbacks by isolating FMA metric reporting.
  Length checks, dispatch, correction or vector broadcasts/arithmetic/output and
  original array retention remain timed; gate/preparation/metric encoding do not.
- Tested 30,000 random triples/sign inversions, special values and midpoint
  counterexamples; fused/unfused distinctions, vector/tail boundaries, aliases,
  canaries, length errors and zero per-call allocations. Compiled/native fixture
  parity, exact path/last-result retention, invalid-option/no-artifact behavior,
  nonfinite JSON and no-Go/source standalone checks pass.
- Full tests/vet, pure packages and forced-emulation paths pass; CI checks native
  FMA smoke and software-path report parity. Added examples, vector/software
  smoke captures, numerical contract and Wiki chapter with arithmetic references
  while retaining Jesús Espino's original article credit.
- Stopped after fused affine. Fused dot, three-array MulAdd and M4 remain pending.

### F004 slice 2c — Multiple reduction accumulators — 2026-11-03

- Added SIMDStriped2/4 for sum and unfused dot, with explicitly unrolled 2/4-vector
  loops. Full vectors are assigned round-robin; remaining full vectors go to
  accumulators 0/1/2, followed by a left-to-right lane-wise merge, low-to-high
  lane fold and scalar tail. Dot products still round before addition, without FMA.
- Preserved original scalar/one-vector kernel bodies and typed callback/harness
  boundaries. New whole-kernel timings include validation, unrolling, merging,
  extraction/fold and tail. Current widths allocate zero per call; the documented
  larger-width scratch fallback remains visible in allocation counters.
- Added strategy/order IDs and VectorAccumulators=2/4 through existing v1 fields.
  Counts describe the algorithm, not measured physical registers. Native metrics,
  result-bit retention, width compatibility and historical fold IDs remain intact.
- Shared catalogue/listing/quick/report/native paths expose all four strategies
  per reduction through existing benchmark selectors; no new flag or byte model.
  Catalogue totals: nine operations and twenty-four implementations.
- Tested independent rational assignment/product/merge/fold oracles, unroll and
  tail boundaries, remaining full vectors, exceptional/subnormal values, finite
  error bounds, dot length/overlap contracts, immutability, nonfusion and zero
  current allocations. Count-sensitive cancellation cases verify real binding
  identity; reports retain compatible four-strategy ordering and reject mixed widths.
- Full tests/vet, pure packages and forced emulation pass. Exports, standalone
  binary execution, compiled/native parity, examples and Make/CLI workflows cover
  new variants. CI adds explicit native two/four-accumulator smoke cases.
- Added the detailed contract, retained sum/dot smoke evidence, runnable example
  and Wiki chapter; updated README/command/capability guides and retained reading credit.
- Stopped after accumulator comparisons. FMA and M4 remain separate pending phases.

### F004 slice 2b — Unfused float32 dot product — 2026-11-03

- Added dot-f32 / BenchmarkDot with Scalar and SIMDStriped. Products round to
  float32 before accumulation, with explicit left-to-right versus one-vector,
  low-to-high lane-fold and scalar-tail orders. No FMA is used.
- Added checked `(float32,error)` kernels/bindings and distinct quick/compiled/
  native execution paths. Unequal lengths return +0,ErrLengthMismatch before
  reading inputs; equal empty inputs return +0,nil. Inputs and read-only overlaps
  are preserved. No output array is allocated.
- Timed length validation, multiplication, extraction/fold and tail. Propagated
  failures while retaining earlier observations; successful returned values are
  retained once after timing. Sum and legacy array call paths stay intact.
- Added separate dot harnesses, arithmetic/order identities and strict native
  result-bits/simd-lanes retention. Existing v1 bit/fold fields preserve results
  including NaN/Inf; mixed widths cannot pool. Historical option records without
  DotInput resolve to balanced without changing prior kernel identities.
- Shared -dot-input balanced/cancellation/mixed-magnitude through both CLIs and
  installed/Go/Make launchers. Fixtures allocate/fingerprint both inputs, account
  for 8*n+4 logical bytes and select only dot leaves in the shared native package.
  Catalogue totals: nine operations and twenty implementations.
- Tested independent rational product/fold oracles, eligible finite error bounds,
  nonfusion at vector/tail boundaries, both rounding-loss directions, exceptional
  values/subnormals, length errors, read-only overlap, immutability and zero
  current-width allocations. Checked error/partial-result paths, all-fixture
  backend parity, exports and standalone result/fold retention without Go/source.
- Added a runnable dot example, detailed contract, retained smoke evidence and
  Wiki chapter; synchronized command/capability guides and preserved article credit.
- Stop after unfused dot. Multiple accumulators, FMA and M4 remain future phases.

### F004 slice 2a — Ordered float32 sum — 2026-11-03

- Added sum-f32 with Scalar left-to-right and SIMDStriped full-vector accumulation,
  low-to-high lane fold and scalar tail. Each addition rounds to float32; empty
  input returns +0. Input is unchanged and no caller output slice is required.
- Implemented typed float32-returning bindings and separate quick/compiled/native
  reduction callbacks. Final extraction/fold/tail stays timed; the last returned
  value is retained once after timing. Existing array kernel paths stay intact.
- Added additive scalar-result uint32 bits and fold identity (order, observed
  process lanes, vector accumulator count, tail) to v1 evidence. New reduction
  harness IDs separate these timings; native metrics retain exact bits/width and
  reject missing or malformed facts. Mixed-width comparisons are rejected.
- Shared -sum-input balanced/cancellation/mixed-magnitude across both CLIs and
  Go/Make/installed launchers. Fixtures allocate only x, fingerprint its bits,
  and account for 4*n+4 logical bytes per call. Catalogue totals are eight
  operations and eighteen implementations.
- Tested independent rational float32 fold oracles, two directions of rounding
  loss, eligible finite error bounds, vector/tail boundaries, exceptional values,
  signed zero, immutability, zero current-width allocations, raw-bit retention,
  compiled/native parity, exports and no-Go/no-source standalone execution.
  Full tests/vet, pure packages and forced emulation pass; CI now includes sum.
- Added runnable cancellation examples, the numerical contract, retained smoke
  evidence and a worked Wiki chapter; updated README/command/capability guides
  and preserved credit to Jesús Espino's SIMD and Go article.
- Stopped after sum. Dot, multiple vector accumulators, FMA and M4 remain pending.

### F004 slice 1 — Broadcast and predicate catalogue — 2026-10-27

- Added scalar/SIMDPartial scale, explicitly unfused affine, threshold-to-zero,
  bit-preserving conditional selection and strict range-mask And composition.
  The catalogue now has seven operations and sixteen implementations.
- Added shared float32 alpha/beta/threshold/lower/upper flags, predicate
  none/all/alternating/clustered/shuffled arrangements and a deterministic seed.
  Relevant parameter bits and fixture identities/fingerprints survive evidence;
  impossible true distributions fail before preparation/artifacts.
- Added unary fixture preparation and descriptor-derived 8*n accounting;
  two-array selection retains 12*n logical work. Constants broadcast per kernel
  call; parameter extraction and buffer binding remain outside timing.
- Preserved existing addition/computation kernels and timed bindings. New
  configured callbacks have explicitly distinct pattern harness identities.
  The capture's optional config.input stores scalar uint32 bits safely for NaNs.
- Supported the same operations/options through quick, standalone report, native
  developer backend, Go/Make and installed launchers. Shared-package native runs
  select exact requested leaves; raw metadata retains configuration, while absent
  historical context remains unknown.
- Made default -list compact, with parameter/strategy/semantic details behind
  -list -group. Added examples/patterns, the complete README catalogue, a worked
  Wiki chapter and a detailed implementation contract retaining article credit.
- Tested independent scalar/SIMD oracles, aliases, invalid lengths/canaries,
  vector/tail boundaries, signed zero, NaNs/Inf, unfused rounding, fixture truth
  distributions and fingerprints, context parity, exports and standalone use.
  Full tests/vet and pure packages pass; forced portable emulation passes and is
  now checked by CI alongside normal execution and Linux race coverage.
- Stopped after this slice. Reductions/dot, FMA, clamp, mask export, conversions,
  new numeric types and architecture-specific movement remain planned.

### Architecture migration M3 — 2026-10-27

- Made compiled `testing.Benchmark` the default report backend. Built/installed
  commands run from arbitrary directories without Go or source at runtime.
  Retained `-backend go-test` as an explicit developer option, with no fallback.
- Added an internal engine boundary, validated positive duration/Nx/count and
  Go-compatible benchmark selection, typed prepared calls, exact duration/N and
  fractional ns/op/counters, plus returned kernel errors and partial observations.
- Serialized the documented testing flag API and restored timing configuration;
  retained Go calibration/GC, existing fixtures, kernels and per-call workers.
- Added runtime/vector/emulation/embedded-build capture without runtime tools;
  missing VCS facts and unsupported CPU model detection remain explicitly unknown.
- Added versioned `go-simd.measurements/v1` JSON evidence under a distinct
  `.measurements.json` extension; explicit developer runs retain `.raw` text.
  Existing table formats, medians and missing-value conventions are unchanged.
- Shared group/alias/list/help definitions between both CLIs. Report now accepts
  the same aliases and listing; diagnostics and help examples are consistent.
- Added Make help/list/build/install, BIN_DIR, and ARGS forwarding for quick group
  shortcuts. Reorganized README/Wiki around installed/Go/Make equivalents, a clear
  flag catalogue, bounded experiments and remote installation/use examples.
- Verified full tests/vet, format/identity compatibility, native-selector parity,
  adapter precision/error/settings tests, actual isolated binary reporting with
  no Go/source, temporary installation and Make workflows. Updated smoke evidence
  and guide sources. Worker manifests/cancellation and new F004 kernels remain
  outside this milestone.

### Architecture migration M2 — 2026-10-27

- Added a shared ordinary-source catalogue for the existing two float32
  operations/six implementations: stable identities, labels/order, arithmetic
  semantics, operand/output shapes, parameter seams and worker policies.
- Added typed native calls and explicit prepared-buffer bindings consumed by
  both simdbench and native benchmarks. Report groups and legacy result identity
  mapping now consume the same metadata. Metadata remains SIMD/testing-free.
- Derived logical byte accounting from descriptors; current operations retain
  `12*n`. Added shape/width/overflow checks and synthetic future-shape tests.
- Added `simdbench -list`, with existing group/alias rules, implemented-operation
  details and live runtime SIMD/emulation context, without measuring or creating
  fixtures/artifacts. Verified the built listing outside the checkout without Go
  on PATH. The source-based report runner remains unchanged until M3.
- Preserved all kernel bodies, timed validation/call boundaries, benchmark names,
  sizes, labels, M0 identities, quick tables and M1 report formats.
- Added binding/registry, vector-boundary/in-place/error, metadata/traffic,
  listing/filter/writer and compatibility checks; full tests/vet and pure-package
  tests pass. Retained before/after native observations and quick comparisons.
- Updated the migration record and progressive user Wiki, including command
  registry and contribution instructions. New F004 kernels remain planned.

### F004 operation-pattern review — 2026-10-27

- Added a tracked F004 gap ledger and staged pattern backlog: broadcast
  parameters, comparisons/masks/selection, clamp, reductions/dot/accumulators,
  explicit FMA, and movement/representation experiments.
- Verified portable versus amd64 `archsimd` availability against Go 1.27.3;
  documented absent portable permutation, horizontal and packed-mask methods.
- Specified operand sources, scalar parameters, output representations, tail
  validity, numerical rules, predicate fixtures, capability gates and per-pattern
  traffic models. Three vector operands do not imply three input arrays.
- Credited Jesús Espino's *SIMD and Go* in the Wiki and added a pattern review
  distinguishing proposed experiments from the six implemented kernels.
- Updated the local F4 tracker to prioritize useful patterns before broad type
  coverage. No kernels, selectors or architecture milestones are implemented
  by this planning/documentation change.
- Validated local SDK symbols, current-kernel coverage, Wiki/source links,
  Markdown tables/fences, and unchanged runtime-source scope.

### User knowledge guide — 2026-10-27

- Added a progressive GitHub Wiki guide with tracked sources under `docs/wiki`,
  sidebar/footer navigation, and an incremental maintenance/publishing workflow.
- Documented both CLIs and every flag, all Make targets/variables, native Go
  benchmark controls, build/test commands, current capabilities, and failures.
- Added operation/alias/numerical-contract references, report/evidence guidance,
  interpretation lessons, focused experiment recipes, and troubleshooting.
- Added `examples/basic`, a runnable module-local demonstration of all six
  kernels, exact in-place aliases, unchanged output on length mismatch, and nil
  input. Measurement kernels and migration scope are unchanged.
- Verified example output, all project flag/Make/kernel reference coverage,
  Wiki/source links, Markdown fences, formatting, full tests, and vet.
- Corrected the M1 export smoke test to select universal sizes 8/16 after CI
  exposed its earlier assumption that the width-dependent size 9 always exists.
  General Wiki smoke recipes now use those portable sizes too.

### Architecture migration M1 / Feature 1 — 2026-10-27

#### Added

- `simdreport -format txt|md|csv`; TXT remains the default and terminal output
  remains a readable text comparison for every saved format.
- Markdown tables with literal environment metadata, and encoded CSV containing
  numeric timing/percentage values with empty missing cells and metadata in the
  raw companion file. Unsupported formats fail before execution or artifacts.
- Typed runs, environments, complete case identities, individual observations,
  repetitions, iteration counts, optional counters, and source provenance in
  `internal/result`, with pure median aggregation and scalar comparisons.
- `internal/input` for shared deterministic preparation and streaming SHA-256
  input fingerprints; historical input identity stays explicitly unknown.
- Pure TXT/MD/CSV renderers in `internal/report`, plus identity, encoding,
  cross-format fixture, escaping, writer-failure, and real CLI export tests.

#### Changed

- Removed parsing/aggregation from `cmd/simdreport` and table formatting from the
  testing-dependent `benchutil` surface. The CLI now orchestrates these packages.
- Preserved M0 table arithmetic/order, filtered selections, raw output and
  failed-run evidence. Mixed or ambiguous measurement contexts now fail rather
  than silently pooling incompatible cases. Hyphenated legacy strategy names
  are preserved when stripping numeric GOMAXPROCS suffixes.
- Updated CLI help, README, Make help, and the tracked migration record.

#### Validation and scope

- Formatting, full uncached tests, and vet pass on Windows/amd64. Core result,
  report, and input packages also pass without `GOEXPERIMENT=simd`.
- Real CLI tests cover both groups, default/explicit formats, selected raw
  companions, readable terminal output, empty selections, and benchmark failures.
- Local full race execution retains the M0 `runtime/cgo` limitation; Linux CI
  remains the full race check. Kernel/timed-loop behavior is unchanged.
- M1 is complete; M2 catalog and M3 standalone-backend work remain pending.

### Architecture migration M0 — 2026-10-27

#### Added

- Compatibility fixtures for both command help screens and multi-size reports,
  covering odd/even medians, ordered variants, positive/negative/zero changes,
  fractional timings, filtered variants, and missing/zero scalar baselines.
- Regression coverage for quick-comparison size, variant bindings, allocation
  columns, aliases, the full legacy size sweep, generated-input wraparound,
  and report validation before artifact creation.
- Tracked M0 migration contract defining operation/type/strategy identity,
  arithmetic and harness version boundaries, input fingerprint encoding,
  selector compatibility, and preservation of legacy raw artifacts.
- Native Windows/amd64 baseline with 162 observations and complete command,
  toolchain, hardware, source-revision, and measurement-boundary context.
- Actual two-group report smoke artifacts and reproduction instructions.

#### Validation

- Formatting, uncached tests, vet, and filtered report execution passed on
  Windows/amd64 with Go 1.27.3 and `GOEXPERIMENT=simd`.
- Local race tests could not build `runtime/cgo` (`cgo.exe` exits 2). This also
  affects the unchanged addition package; Linux CI remains the configured race
  check. No local Linux/arm64 execution is claimed.

#### Compatibility

- M0 characterizes existing behavior. Kernel implementations, timed loops,
  command defaults, report formats, and measurement algorithms are unchanged.
- M0 is complete. M1 result/rendering extraction and Feature 1 remain pending.
