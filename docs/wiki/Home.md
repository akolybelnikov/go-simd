# go-simd user knowledge guide

## What is SIMD?

A CPU can apply the same operation to several values at once. Start with
[What is SIMD?](00-What-is-SIMD) for a simple example, where it helps, and
why it can sometimes be slower.

[Install the CLI](https://github.com/akolybelnikov/go-simd/blob/main/docs/cli.md#install-the-cli-from-source) to start a comparison.

## Open the local browser UI

Run `simdlab ui` to open the browser on the measuring machine. **Explore**
lists supported operations and methods, **Compare** previews and runs a
comparison, and **Results** shows timings or opens a saved capture. The CLI
remains available for automation and SSH.
[Browser guide](https://github.com/akolybelnikov/go-simd/blob/main/docs/local-ui.md)

## Run directly

```text
simdlab run addition
```

This compares ordinary and vectorized addition on two generated arrays of
1,024 float32 values, with three repeats per method. The terminal displays
the settings, runs the comparison, and prints the timing difference and the
paths to the Markdown report and measurement file. This path works without
a model. Press Ctrl+C to cancel.

## Ask an agent

The examples below work without Ollama. Other wording may use the optional local model.

```text
simdlab ask
```

Enter **Compare ordinary and vectorized addition for 1,024 numbers.** Review
the methods, batch size, and repeats, then enter **yes** to measure or **no**
to decline. After the run, ask **Which took less time, and by how much?** or
**Were the repeats consistent?** The answer uses measurements saved during
the session. Enter **quit** to leave.

Follow the [first experiment walkthrough](02-First-experiment) or see the
[CLI reference](https://github.com/akolybelnikov/go-simd/blob/main/docs/cli.md) for Ollama setup and exact flags.

## Choose your starting point

| You want to… | Read |
| --- | --- |
| Open the lab and prepare your machine | [Getting started](01-Getting-started) |
| Run a comparison and interpret the outcome | [Your first addition comparison](02-First-experiment) |
| Find terminal commands and options | [Command reference](03-Command-registry) |
| Understand the methods and why their timings differ | [Operation guide](04-Operation-guide) |
| See what each benchmark times and where the operation is used | [How the operation benchmarks work](32-Operation-measurement-guide) |
| Save and reopen measurements | [Reports and evidence](05-Reports-and-evidence) |
| Read medians, variability, and timing percentages | [Understanding results](06-Understanding-results) |
| Design a focused comparison | [Experiment recipes](07-Experiment-recipes) |
| Learn how arrays, batches, and tails fit together | [Working with arrays](08-Using-the-kernels) |
| Diagnose problems opening the lab or comparing results | [Troubleshooting](09-Troubleshooting) |
| Look up terms and what the lab can measure | [Glossary and capabilities](11-Glossary-and-capabilities) |
| Follow recommended reading on SIMD | [SIMD patterns and reading](12-SIMD-patterns-and-reading) |

## Explore by topic

- [Broadcasts and predicates](13-Broadcast-and-predicate-experiments) · [Byte masks](19-Predicate-export-and-byte-masks) · [Packed masks](20-Packed-predicate-masks) · [Clamp](22-Clamp-and-numerical-contracts)
- [Array sums](14-Sum-and-reduction-order) · [Dot products](15-Dot-product-and-unfused-rounding) · [Accumulator counts](16-Accumulator-count-experiments) · [Block sums](29-Fixed-block-horizontal-sum)
- [Fused affine transforms](17-Fused-affine-and-rounding) · [Fused dot products](18-Fused-dot-and-fold-orders) · [Three-array multiply-add](24-Three-array-fused-multiply-add)
- [Float32 copy](21-Copy-and-data-movement) · [Float64 addition](25-Float64-addition) · [Float64 copy](31-Exact-Float64-copy) · [Saturating byte addition](26-Uint8-saturating-addition)
- [Bit reinterpretation](27-Bit-reinterpretation-round-trip) · [Bitwise XOR](28-XOR-of-float32-bit-patterns) · [Carryless multiplication](30-Carryless-polynomial-multiplication)
