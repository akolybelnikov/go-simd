# go-simd user knowledge guide

Learn to run, understand, and extend a small Go SIMD experiment laboratory.
Start with a smoke measurement, learn the comparison boundaries, then design a
repeatable experiment on your own toolchain and machine.

**Guide baseline:** architecture M3 / Feature 1, reviewed 2026-10-27.
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

## What you can do today

- Compare four addition strategies and two implementations of `a*a + b*b + a*b`
  over generated `float32` slices.
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

Project links: [source](https://github.com/akolybelnikov/go-simd),
[changelog](https://github.com/akolybelnikov/go-simd/blob/main/CHANGELOG.md),
[migration checkpoints](https://github.com/akolybelnikov/go-simd/blob/main/docs/migration/README.md),
[issues](https://github.com/akolybelnikov/go-simd/issues).
