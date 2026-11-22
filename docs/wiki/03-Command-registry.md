# Command reference

Run these commands from a terminal where the lab executables are on your PATH. The browser is a convenient way to choose methods and a size interactively.

| Task | Command |
| --- | --- |
| Open the browser lab | `simdlab ui` |
| Print the private browser address | `simdlab ui -no-open` |
| Choose an installed experimental local chat model | `simdlab ui -ollama` |
| Configure OpenAI chat model (API key required) | `simdlab ui -openai-model MODEL` |
| Test an installed Ollama model in browser chat | `simdlab ui -ollama -model MODEL` |
| Compare two ways of adding 1,024 values | `simdlab run addition` |
| Browse operations and methods in the terminal | `simdlab catalog -text` |
| Open the guided terminal session | `simdlab ask` |
| Inspect a saved capture | `simdlab inspect -text "path/to/measurements.json"` |
| Run a quick addition timing in the terminal | `simdbench -group addition` |
| Save a Markdown report for an addition sweep | `simdreport -group addition -format md` |

Use `-h` with a command to see its options. The [CLI guide](https://github.com/akolybelnikov/go-simd/blob/main/docs/cli.md) includes advanced comparison examples.
For OpenAI chat setup and the browser controls, see [Chat with an agent](33-Chat-with-an-agent).

## Reporting options

| Option | Effect |
| --- | --- |
| `-group` | Choose an operation for the terminal sweep. Find names with `simdreport -list`. |
| `-count` | Choose how many repeats to save for each measured case. |
| `-benchtime` | Set a time target such as `100ms` or a fixed iteration count such as `10x`. |
| `-format` | Choose `txt`, `md`, or `csv` for the formatted report. |
| `-output` | Choose the output folder for `simdreport`. |
| `-artifacts` | Choose where `simdlab ui` and `simdlab run` save measurements. |

## Fixture options

Fixtures generate data that exposes specific differences. Keep fixture settings the same between methods.

| Option | What it controls |
| --- | --- |
| `-sum-input` | Balanced, cancellation, or mixed-magnitude input for sums. |
| `-dot-input` | Balanced, cancellation, mixed-magnitude, or fusion-sensitive input for dot products. |
| `-fma-input` | Balanced, rounding-sensitive (`rounding`), or mixed-magnitude input for fused multiply-add. |
| `-alpha`, `-beta` | The shared scale and offset where the operation uses them. |
| `-threshold` | The strict cutoff for thresholding and selection. |
| `-lower`, `-upper` | The two bounds for range filtering or clipping. |
| `-pattern`, `-seed` | The arrangement of passing or clipped values. Choose none, all, alternating, clustered, or shuffled. The seed repeats a particular shuffle. |

## Build, install and run remotely

Use the [installation guide](https://github.com/akolybelnikov/go-simd/blob/main/docs/cli.md#install-the-cli-from-source) for the source build or choose an archive for your platform. To compare on a remote machine, start the terminal command there. The timings describe that machine, and the reports are saved there.

Local model selection is explicit; no local browser-chat model is currently recommended. [AI setup and testing](34-AI-setup-and-model-recommendations).
