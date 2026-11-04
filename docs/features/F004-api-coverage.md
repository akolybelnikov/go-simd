# F004 — Baseline SDK API coverage ledger

Reviewed 2026-11-04 against **Go 1.27.3-X:simd, GOEXPERIMENT=simd, windows/amd64**.
This inventory accounts for every exported type/function/method/constant/variable
declaration in the selected simd package files, once each. It records coverage
status; it does not claim exhaustive operation kernels. F004 is [complete for
the agreed experimental scope](F004-completion.md).

The snapshot contains **381 declarations**, including fourteen vector/mask types.
The [machine-readable snapshot](evidence/f004-sdk-inventory.json) retains exact
symbol names, receiver, declaration kind, source file, status, reason and source
SHA-256. An AST inventory was made from the GoFiles selected by
`GOEXPERIMENT=simd go list -json simd`; exported FuncDecl/TypeSpec/ValueSpec names
were sorted and checked for duplicates. Source hashes pin this local experimental
SDK baseline; another Go version or target must be inventoried independently.

## Status totals

| Status | Declarations | Meaning |
| --- | --- | --- |
| Deferred bridge | 28 | To/from architecture vectors; optional gated extension. |
| Deferred member | 279 | Separate experiment needs an explicit contract. |
| Deferred type | 7 | No input/output kernel for that vector type. |
| Exercised member | 33 | Used in admitted kernels, without claiming isolated primitive benchmarks. |
| Exercised type | 7 | Vector type used by current kernels; not a separate command. |
| Infrastructure used | 7 | Live sizing/runtime metadata, rather than arithmetic experiments. |
| Utility outside scope | 20 | SDK sizing/formatting/capability utility unused in delivered lab paths. |

Float32, float64, uint8 and uint64 are CLI input types; the latter two initially support
addition, saturating addition and low64 carryless products respectively. Int32s is used solely for predicate export
scratch, Uint32s is an internal bit-reinterpretation/logic vector, and Mask32s is
internal comparison state. Uint8s saturation now uses real byte inputs; predicate export remains distinct. FMA availability remains gated by its verified
reference and actual path; SDK emulation is recorded separately.

## Source snapshot

| Selected SDK file | SHA-256 |
| --- | --- |
| doc.go | b577e72ac5e623d98ac313aab5a31867fd9286471af20040ab62422e045d86e9 |
| midway_amd64.go | a187562095fad6d2f885ec7b22db2101c4a0d61586897f6e6b4aa6b44a4da9f2 |
| midway_common.go | d573ef7e915cf5153f4442a801bab1fb58a4144ea48e08d2e158ed4612f559bc |
| simd_stubs.go | c003d32bdcbf168f70dc34e193b333562b0c227b50539af1ceede6c5d1eb12be |
| simd_types.go | f8ff9a05803cdc01a60c1d4f5f9fd4005744455dba1a1c45856f3ed8cd602181 |
| tofrom_amd64.go | 62c56237338ab6258fdc92172d8eef3c3c7288a139a9536abe0674876474a874 |

simd_emulated.go supplies semantic cross-checks but is not a selected compiled
declaration file here; GODEBUG=simd=0 execution is verified in separate processes.
The independent simd/archsimd API is outside this baseline inventory; selected
ToArch/FromArch bridges are explicitly deferred below. No portable Permute,
GetElem, horizontal-pair/Sum/dot primitive or Mask32s.ToBits is invented.

## Evidence and limits

Kernel usage: internal/addition, internal/computation, internal/patterns,
internal/reduction, internal/integer and internal/movement; sizing/context: internal/platform.
Per-operation numerical/traffic/test references live in the [closure audit](F004-completion.md).
This is source-use coverage, not proof that every SDK instruction is accelerated
on every machine, or that all argument combinations of a method are tested.
Deferred families include subtraction/division/sqrt, broader conversions and
bit reinterpretation, integer shifts/saturation/carryless multiplication and
further typed operations and integer input types. No new commands are implied by this ledger.

## Package types and functions

| Symbol | Status | Reason |
| --- | --- | --- |
| `BroadcastFloat32s` | Exercised member | Loads or per-call broadcasts inside admitted float32 kernels. |
| `BroadcastFloat64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `BroadcastInt16s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `BroadcastInt32s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `BroadcastInt64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `BroadcastInt8s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `BroadcastUint16s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `BroadcastUint32s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `BroadcastUint64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `BroadcastUint8s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Emulated` | Infrastructure used | Runtime metadata in internal/platform; no throughput comparison. |
| `Float32s` | Exercised type | Primary generated input/output and arithmetic vector. |
| `Float32sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Float64s` | Exercised type | True float64 addition input/output vector; no implicit float32 widening. |
| `Float64sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `HasHardwareCarrylessMultiply` | Infrastructure used | Record SDK carryless availability separately from instruction claims. |
| `Int16s` | Deferred type | No admitted input/output kernel for this SDK vector type; no implicit cast advertised. |
| `Int16sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Int32s` | Exercised type | Intermediate canonical predicate conversion/extraction only; no integer input experiment. |
| `Int32sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Int64s` | Deferred type | No admitted input/output kernel for this SDK vector type; no implicit cast advertised. |
| `Int64sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Int8s` | Deferred type | No admitted input/output kernel for this SDK vector type; no implicit cast advertised. |
| `Int8sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `LoadFloat32s` | Exercised member | Loads or per-call broadcasts inside admitted float32 kernels. |
| `LoadFloat32sPart` | Exercised member | Loads or per-call broadcasts inside admitted float32 kernels. |
| `LoadFloat64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadFloat64sPart` | Exercised member | Timed bounded Float64 loads; emulated count workaround included in kernel. |
| `LoadInt16s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadInt16sPart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadInt32s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadInt32sPart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadInt64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadInt64sPart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadInt8s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadInt8sPart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadUint16s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadUint16sPart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadUint32s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadUint32sPart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadUint64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadUint64sPart` | Exercised member | Bounded uint64 loads inside clmul64. |
| `LoadUint8s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `LoadUint8sPart` | Exercised member | Timed bounded true byte-input loads. |
| `Mask16s` | Deferred type | No admitted input/output kernel for this SDK vector type; no implicit cast advertised. |
| `Mask16sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Mask32s` | Exercised type | Internal comparison/composition/selection only; no raw mask representation exported. |
| `Mask32sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Mask64s` | Deferred type | No admitted input/output kernel for this SDK vector type; no implicit cast advertised. |
| `Mask64sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Mask8s` | Deferred type | No admitted input/output kernel for this SDK vector type; no implicit cast advertised. |
| `Mask8sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Uint16s` | Deferred type | No admitted input/output kernel for this SDK vector type; no implicit cast advertised. |
| `Uint16sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Uint32s` | Exercised type | Intermediate opaque bit representation in Float32 roundtrip; no uint32 CLI input family. |
| `Uint32sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Uint64s` | Exercised type | True uint64 input/output carryless kernel. |
| `Uint64sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Uint8s` | Exercised type | Real uint8 input/output vectors in saturatingadd8. |
| `Uint8sFromArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `VectorBitSize` | Infrastructure used | Runtime metadata in internal/platform; no throughput comparison. |

## Float32s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Float32s.Abs` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float32s.Add` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.ConvertToInt32` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.Div` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float32s.Equal` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.Greater` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float32s.IfElse` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.Len` | Infrastructure used | Actual lanes for kernel loops and evidence; setup cost follows each kernel contract. |
| `Float32s.Less` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float32s.Masked` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.Max` | Exercised member | Clamp SIMDMinMax, with all bit-preserving repairs included in timing. |
| `Float32s.Min` | Exercised member | Clamp SIMDMinMax, with all bit-preserving repairs included in timing. |
| `Float32s.Mul` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.MulAdd` | Exercised member | Gated native FMA paths only; software fallback and actual path retained. |
| `Float32s.Neg` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float32s.NotEqual` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.Sqrt` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float32s.Store` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.StorePart` | Exercised member | Used inside admitted float32 kernels; composed semantics, not isolated instruction latency. |
| `Float32s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Float32s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float32s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Float32s.ToBits` | Exercised member | Timed source representation path in exact bitroundtrip; machine instruction emission not assumed. |

## Float64s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Float64s.Abs` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Add` | Exercised member | Timed binary64 addition in addition64. |
| `Float64s.Div` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Greater` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Len` | Infrastructure used | Actual Float64 lanes for bounded kernel loads/stores and evidence. |
| `Float64s.Less` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Max` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Min` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Mul` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.MulAdd` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Neg` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Sqrt` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.Store` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.StorePart` | Exercised member | Timed bounded Float64 output stores in addition64. |
| `Float64s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Float64s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Float64s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Float64s.ToBits` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Int16s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Int16s.Abs` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Add` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.AddSaturated` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.AndNot` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.ConvertToUint16` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Greater` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Len` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Int16s.Less` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Max` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Min` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Mul` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Neg` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Not` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.RotateAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.RotateAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.ShiftAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.ShiftAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Store` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.StorePart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Int16s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.SubSaturated` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Int16s.ToBits` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.ToMask` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int16s.Xor` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Int32s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Int32s.Abs` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Add` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.AndNot` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.ConvertToFloat32` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.ConvertToUint32` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Greater` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Len` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Int32s.Less` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Max` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Min` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Mul` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Neg` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Not` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.RotateAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.RotateAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.ShiftAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.ShiftAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Store` | Exercised member | Timed predicate conversion/extraction scratch store only. |
| `Int32s.StorePart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Int32s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Int32s.ToBits` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.ToMask` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int32s.Xor` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Int64s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Int64s.Add` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.AndNot` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.ConvertToUint64` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Greater` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Len` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Int64s.Less` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Neg` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Not` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.RotateAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.RotateAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.ShiftAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Store` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.StorePart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Int64s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Int64s.ToBits` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.ToMask` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int64s.Xor` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Int8s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Int8s.Abs` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Add` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.AddSaturated` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.AndNot` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.ConvertToUint8` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Greater` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Len` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Int8s.Less` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Max` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Min` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Mul` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Neg` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Not` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Store` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.StorePart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Int8s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.SubSaturated` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Int8s.ToBits` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.ToMask` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Int8s.Xor` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Mask16s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Mask16s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Mask16s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Mask16s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Mask16s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Mask16s.ToInt16s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Mask32s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Mask32s.And` | Exercised member | Timed strict-range And or clamp tie/NaN Or composition. |
| `Mask32s.Or` | Exercised member | Timed strict-range And or clamp tie/NaN Or composition. |
| `Mask32s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Mask32s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Mask32s.ToInt32s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Mask64s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Mask64s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Mask64s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Mask64s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Mask64s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Mask64s.ToInt64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Mask8s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Mask8s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Mask8s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Mask8s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Mask8s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Mask8s.ToInt8s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Uint16s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Uint16s.Add` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.AddSaturated` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.AndNot` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Average` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.BitsToInt16` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.ConvertToInt16` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Greater` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Len` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Uint16s.Less` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Max` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Min` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Mul` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Not` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.ReshapeToUint32s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.ReshapeToUint64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.ReshapeToUint8s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.RotateAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.RotateAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.ShiftAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.ShiftAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.Store` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.StorePart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Uint16s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.SubSaturated` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint16s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Uint16s.Xor` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Uint32s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Uint32s.Add` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.AndNot` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.BitsToFloat32` | Exercised member | Exact return from opaque Uint32 representation to Float32 bits. |
| `Uint32s.BitsToInt32` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.ConvertToInt32` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Greater` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Len` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Uint32s.Less` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Max` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Min` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Mul` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Not` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.ReshapeToUint16s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.ReshapeToUint64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.ReshapeToUint8s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.RotateAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.RotateAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.ShiftAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.ShiftAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.Store` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.StorePart` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Uint32s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint32s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Uint32s.Xor` | Exercised member | Timed opaque32-bit XOR in xorbits, through Float32/Uint32 reinterpretation. |

## Uint64s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Uint64s.Add` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.AndNot` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.BitsToFloat64` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.BitsToInt64` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.CarrylessMultiplyEven` | Exercised member | Even input products in timed low64 polynomial kernel. |
| `Uint64s.CarrylessMultiplyOdd` | Exercised member | Odd input products in timed low64 polynomial kernel. |
| `Uint64s.ConvertToInt64` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.Greater` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.GreaterEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.Len` | Infrastructure used | Actual uint64 width in kernels and retained evidence. |
| `Uint64s.Less` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.LessEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.Not` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.ReshapeToUint16s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.ReshapeToUint32s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.ReshapeToUint8s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.RotateAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.RotateAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.ShiftAllLeft` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.ShiftAllRight` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.Store` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.StorePart` | Exercised member | Timed uint64 scratch stores for low-word extraction. |
| `Uint64s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Uint64s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint64s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Uint64s.Xor` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |

## Uint8s members

| Symbol | Status | Reason |
| --- | --- | --- |
| `Uint8s.Add` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.AddSaturated` | Exercised member | Timed byte saturating addition, distinct from wrapping. |
| `Uint8s.And` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.AndNot` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Average` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.BitsToInt8` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.ConvertToInt8` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Equal` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.IfElse` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Len` | Infrastructure used | Actual Uint8 lanes for bounded kernels and evidence. |
| `Uint8s.Masked` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Max` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Min` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Mul` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Not` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.NotEqual` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Or` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.ReshapeToUint16s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.ReshapeToUint32s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.ReshapeToUint64s` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.Store` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.StorePart` | Exercised member | Timed bounded numeric byte-output stores. |
| `Uint8s.String` | Utility outside scope | SDK utility accounted for; unused in delivered lab paths, not an arithmetic throughput operation. |
| `Uint8s.Sub` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.SubSaturated` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
| `Uint8s.ToArch` | Deferred bridge | Architecture bridge; a separately gated fixed-width experiment is outside current scope. |
| `Uint8s.Xor` | Deferred member | No admitted standalone experiment; choose numerical/representation/type contract before adding a kernel. |
