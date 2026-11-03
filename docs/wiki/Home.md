# go-simd user knowledge guide

Learn to run, understand, and extend a small Go SIMD experiment laboratory.
Start with a smoke measurement, learn the comparison boundaries, then design a
repeatable experiment on your own toolchain and machine.

**Guide baseline:** architecture M3 / F004 fused-dot checkpoint, reviewed 2026-11-03.
Command names and examples describe implemented behavior. Future work is marked
explicitly. This guide grows alongside each delivered feature.

## Choose your starting point

| You want to… | Read |
| --- | --- |
| Install prerequisites and verify the checkout | [01 — Getting started](https://github.com/akolybelnikov/go-simd/wiki/01-Getting-started) |
| Complete a small experiment from start to finish | [02 — Your first experiment](https://github.com/akolybelnikov/go-simd/wiki/02-First-experiment) |
| Look up every project command, flag, and Make target | [03 — Command registry](https://github.com/akolybelnikov/go-simd/wiki/03-Command-registry) |
| Understand the implemented algorithms and functions | [04 — Operation guide](https://github.com/akolybelnikov/go-simd/wiki/04-Operation-guide) |
| Save TXT, Markdown, CSV, and raw evidence | [05 — Reports and evidence](https://github.com/akolybelnikov/go-simd/wiki/05-Reports-and-evidence) |
| Explain timings, percentages, tails, and concurrency | [06 — Understanding results](https://github.com/akolybelnikov/go-simd/wiki/06-Understanding-results) |
| Run focused, repeatable experiments | [07 — Experiment recipes](https://github.com/akolybelnikov/go-simd/wiki/07-Experiment-recipes) |
| Call the kernels from Go code inside this module | [08 — Using the kernels](https://github.com/akolybelnikov/go-simd/wiki/08-Using-the-kernels) |
| Diagnose setup, flags, selection, and report failures | [09 — Troubleshooting](https://github.com/akolybelnikov/go-simd/wiki/09-Troubleshooting) |
| Contribute and keep this Wiki current | [10 — Contribution and Wiki maintenance](https://github.com/akolybelnikov/go-simd/wiki/10-Contributing-and-wiki-maintenance) |
| Check terminology and current capability limits | [11 — Glossary and capabilities](https://github.com/akolybelnikov/go-simd/wiki/11-Glossary-and-capabilities) |
| Explore missing SIMD patterns and credited reading | [12 — SIMD patterns and reading](https://github.com/akolybelnikov/go-simd/wiki/12-SIMD-patterns-and-reading) |
| Try broadcast constants, masks and predicate arrangements | [13 — Broadcast and predicate experiments](https://github.com/akolybelnikov/go-simd/wiki/13-Broadcast-and-predicate-experiments) |
| Sum an array and understand order-sensitive rounding | [14 — Sum and reduction order](https://github.com/akolybelnikov/go-simd/wiki/14-Sum-and-reduction-order) |
| Compute a dot product with explicit unfused rounding | [15 — Dot product and unfused rounding](https://github.com/akolybelnikov/go-simd/wiki/15-Dot-product-and-unfused-rounding) |
| Compare one, two and four independent vector accumulators | [16 — Accumulator-count experiments](https://github.com/akolybelnikov/go-simd/wiki/16-Accumulator-count-experiments) |
| Compare fused affine and inspect true float32 rounding | [17 — Fused affine and rounding](https://github.com/akolybelnikov/go-simd/wiki/17-Fused-affine-and-rounding) |
| Compare fused dot updates, accumulator orders and actual execution paths | [18 — Fused dot and fold orders](https://github.com/akolybelnikov/go-simd/wiki/18-Fused-dot-and-fold-orders) |

## What you can do today

- Compare four addition strategies and two implementations of `a*a + b*b + a*b`
  over generated `float32` slices.
- Compare scalar/SIMD scale, affine, threshold, selection and range masking,
  with exact scalar parameters and controlled predicate distributions.
- Compare left-to-right and striped float32 sum with balanced, cancellation and
  mixed-magnitude inputs, retaining scalar result bits and fold identity.
- Compare unfused dot strategies over two generated arrays, with checked scalar
  returns, product/fold semantics and four fixture choices shared with fused dot.
- Compare 1/2/4-vector accumulation strategies for sum/dot with explicit merge
  orders and retained accumulator-count/result identities.
- Compare fused affine with a corrected software baseline and a gated vector
  or explicitly recorded software path.
- Compare fused dot with Scalar and 1/2/4-vector accumulators, explicit merges,
  software tails and per-case execution paths.
- Run a quick comparison at 1,048,576 elements or a configurable multi-size sweep.
- Inspect registered operations and live SIMD context with either CLI's `-list`.
- Run installed commands on a compatible local or remote machine without Go.
- Select benchmark groups, sizes, and variants using Go benchmark expressions.
- Choose repetitions and a target measurement duration or fixed iteration count.
- Save median tables as TXT, Markdown, or CSV, with companion raw observations.
- Run correctness checks and inspect the actual Go kernels and measurement code.
- Explore serial vector work, partial loads/stores, scalar tails, and per-call
  goroutine overhead. Results can show regressions as well as improvements.

The packages currently live under `internal/`. This is a source-checkout lab,
not a stable public import API. Installed commands run without Go or source;
`go run`, Make and the explicit developer backend need the checkout. See the
[capability ledger](https://github.com/akolybelnikov/go-simd/wiki/11-Glossary-and-capabilities)
before planning file imports, additional numeric types, or installed tooling.

## A one-command first measurement

After setting `GOEXPERIMENT=simd` in your shell, run from the repository root:

```powershell
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1$/^(Scalar|SIMDPartial)$' -count 1 -benchtime 1x
```

This checks setup and report generation. One iteration is a smoke check, not a
sound performance conclusion. Continue with the first-experiment guide for a
useful repeated measurement.

## Guide history

| Date | Delivered knowledge |
| --- | --- |
| 2026-10-27 | Initial progressive guide, full command registry, six-kernel reference, report conventions, experiment recipes, troubleshooting, and runnable Go example |
| 2026-10-27 | F004 pattern-gap review, portable/architecture API checks, and credit to Jesús Espino's *SIMD and Go* |
| 2026-10-27 | M2 shared catalogue, typed bindings, descriptor traffic, runtime listing and contributor workflow |
| 2026-10-27 | M3 standalone reports, exact structured evidence, consistent flags/Make launchers, installation and remote-use guide |
| 2026-10-27 | F004 first slice: five broadcast/predicate operations, scalar bits, seeded arrangements, compact catalogue and worked examples |
| 2026-11-03 | F004 sum: typed scalar results, specified fold orders, three input recipes, exact result-bit evidence and cancellation examples; eight operations/eighteen implementations |
| 2026-11-03 | F004 dot: unfused products, checked scalar returns, both-input fingerprints and compiled/native evidence parity; nine operations/twenty implementations |
| 2026-11-03 | F004 accumulators: sum/dot with 1/2/4 vector accumulators, declared merge orders, count-sensitive examples and result identity; nine operations/twenty-four implementations |
| 2026-11-03 | F004 fused affine: one-rounding reference, double-rounding counterexamples and explicit vector/software evidence; ten operations/twenty-six implementations |
| 2026-11-03 | F004 fused dot: corrected FMA updates, four specified folds, actual vector/software/tail paths and shared fusion-sensitive fixture; eleven operations/thirty implementations |

Project links: [source](https://github.com/akolybelnikov/go-simd),
[changelog](https://github.com/akolybelnikov/go-simd/blob/main/CHANGELOG.md),
[migration checkpoints](https://github.com/akolybelnikov/go-simd/blob/main/docs/migration/README.md),
[issues](https://github.com/akolybelnikov/go-simd/issues).
