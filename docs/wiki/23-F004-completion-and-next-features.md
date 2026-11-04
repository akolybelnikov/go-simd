# 23 — F004 completion and future extensions

Current catalogue: **23 operations / 58 implementations**. Counts and stop
statements below describe this chapter’s historical delivery checkpoint. See
[the current contract index](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-completion.md).

**F004 is complete for the agreed current scope, 2026-11-04.** Clamp is the final
slice. The shared catalogue contains fifteen operations / forty-one implementations.
No additional slice is needed for this checkpoint.

The delivered experiments cover broadcasting, strict predicates, mask And/Or,
selection, sums/dots and accumulator counts, fused affine/dot, canonical and
packed mask export, exact-copy controls and ordered clipping. Each has a scalar
control, explicit numerical/alias rules, typed output, declared timed work,
logical byte model, boundary tests and CLI/Wiki examples. Exact operation sizes
and all repetitions remain available in reports/evidence.

The [acceptance audit](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-completion.md)
maps the requirements to their evidence. The [baseline SDK ledger](https://github.com/akolybelnikov/go-simd/blob/main/docs/features/F004-api-coverage.md)
accounts for every exported declaration in the selected Go 1.27.3 simd package,
with exercised/helper/deferred status and source hashes. It does not turn every
available SDK method into an implemented lab command. Use -list for commands.

## What remains optional

| Extension | Why it is separate |
| --- | --- |
| Three-array element-wise MulAdd | Subsequently implemented; see chapter 24 for its distinct 16*n operand/traffic contract |
| Float64 or integer inputs | New typed fixtures, numerical policies and bindings |
| Standalone conversion, bit reinterpretation, saturation | Different correctness and representation questions |
| Fixed-width permutations / horizontal primitives | Architecture/width/capability gates and partial-block rules |
| Imported datasets and type inference | Conditional F002 feature; current CLI uses generated fixtures |
| Worker isolation and resource accounting | Architecture M4; not required to add clamp |
| Interactive/HTML reports | F008 frontend work; TXT/Markdown/CSV plus raw evidence are available |

These are future choices, rather than work silently started by closing F004.
Current tests cover Linux/Windows CI and SDK emulation; other platforms require
execution evidence. Performance still depends on workload, size, compiler and
hardware. The library is experimental and its kernel packages remain internal.

Reading credit: [Jesús Espino, SIMD and Go](https://internals-for-interns.com/posts/go-simd/)
informed the pattern review. See [SIMD patterns and reading](https://github.com/akolybelnikov/go-simd/wiki/12-SIMD-patterns-and-reading)
for its context and the distinction between portable and architecture-specific APIs.
