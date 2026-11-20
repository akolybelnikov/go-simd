# CLI setup and reference

## Install the CLI from source

For a fresh checkout, run `git clone https://github.com/akolybelnikov/go-simd.git`
and `cd go-simd`.

Use Go **1.27.3** or a compatible patch release with `GOEXPERIMENT=simd`. Run
the setup for your shell from the checkout. CI verifies Linux/amd64 and
Windows/amd64. macOS source builds are outside the current CI matrix.

**macOS and Linux (sh, Bash, or Zsh)**

```sh
export GOEXPERIMENT=simd
go install ./cmd/simdlab ./cmd/simdbench ./cmd/simdreport
go_bin=$(go env GOBIN)
if [ -z "$go_bin" ]; then
  go_path=$(go env GOPATH)
  go_bin="${go_path%%:*}/bin"
fi
export PATH="$go_bin:$PATH"
simdlab help
```

**Windows (PowerShell)**

```powershell
$env:GOEXPERIMENT = 'simd'
go install ./cmd/simdlab ./cmd/simdbench ./cmd/simdreport
$goBin = go env GOBIN
if (-not $goBin) { $goBin = Join-Path ((go env GOPATH).Split([IO.Path]::PathSeparator)[0]) 'bin' }
$env:Path = "$goBin;$env:Path"
simdlab help
```

The PATH change applies to the current terminal. Add the same directory to
your shell startup file or Windows user PATH to keep it. Once installed, the
commands below are the same across platforms and run outside the checkout.
See [build or install](../README.md#build-or-install) for local builds and copying a binary
to a measuring host.

## Starter commands

Both starter paths use a 100 ms measurement target and three repeats per method.
Use `-artifacts DIR` before the recipe, if any, to choose where to save results.
The default is `artifacts/experiments/` in the current working directory.

| Command | Purpose |
| --- | --- |
| `simdlab ui` | Open the [local browser UI](local-ui.md) for interactive comparisons. |
| `simdlab run addition` | Run the starter with generated data. |
| `simdlab ask` | Explore operations, approve a comparison, or explain saved results. |
| `simdlab ask -evidence "path/to/measurements.json"` | Start a session with saved measurements. |
| `simdlab inspect -text "path/to/measurements.json"` | Show medians, ranges, and measuring-host details. |
| `simdbench -group addition` | Terminal demo. Use `simdlab run addition` to save a starter report. |

The quick-start prompts work without Ollama. Other wording may use the optional
local model. `simdlab ask` uses `qwen3:4b` at `http://127.0.0.1:11434` by default.
Use `-model`, `-endpoint`, or `-timeout` (default 30 seconds) to change these
settings. A remote endpoint receives the question and catalog choices for
model-assisted requests.

The guided session compares the ordinary scalar method with one alternative
at one size. It uses 1,024 values if you omit the size and asks for a supported
size if needed. Type `details` for the measuring host, timings, and caveats.
Use the exact plan flags below for custom repeat counts or size sweeps.
For a result at a particular size, ask “Which method was faster at 1,024
values?” To see the saved run separately, ask “Explain the saved result.”

## Other ways to launch the tools

Flags are the same for installed commands, `go run` and Make. Only the launcher
changes. Make and `go run` need Go and the checkout. Installed binaries run from
any working directory.

| Task | Installed / on PATH | Go from checkout | Make from checkout |
| --- | --- | --- | --- |
| List operations and runtime context | `simdreport -list` | `go run ./cmd/simdreport -list` | `make list` |
| Quick comparison, all groups | `simdbench` | `go run ./cmd/simdbench` | `make run` |
| Quick addition comparison | `simdbench -group addition` | `go run ./cmd/simdbench -group addition` | `make run-addition` |
| Save a Markdown sweep | `simdreport -format md` | `go run ./cmd/simdreport -format md` | `make report ARGS="-format md"` |
| Report help | `simdreport -h` | `go run ./cmd/simdreport -h` | `make report ARGS="-h"` |
| Quick help | `simdbench -h` | `go run ./cmd/simdbench -h` | `make run ARGS="-h"` |

## Review a custom benchmark without Ollama

1. Find the strategy IDs and suggested sweep sizes for addition. A lab plan
   accepts any whole number from 1 to 16,777,216 values per input array.

   ```text
   simdlab catalog -text -group addition
   ```

2. Request a small, measured comparison with two strategies and three repeats
   each. Review the displayed plan and measuring host. Copy the exact
   `run sha256:…` line from the prompt within five minutes to run it once, or
   press Ctrl+C to decline. `-preview` opts in to the experiment CLI.

   ```text
   simdlab experiment -preview -group addition -variants scalar,simd-partial -sizes 1024 -count 3 -benchtime 100ms
   ```

3. Copy the `evidence` path from the final JSON output and replace the example
   path below. By default, experiment evidence is saved under
   `artifacts/experiments/` relative to the working directory. Use `-artifacts`
   to choose another location.

   ```text
   simdlab inspect -text "path/to/measurements.json"
   ```

For scripts, omit `-text` to get JSON from `catalog` and `inspect`. `simdlab plan`
reads an explicit JSON request from stdin and prints the resolved plan without
measuring. For a direct Markdown report, use the report CLI:

```text
simdreport -group addition -bench "^BenchmarkAddition$/^n=1024$" -count 3 -benchtime 100ms -format md
```

The report writes its output and companion capture under `artifacts/addition/`.
The full default sweep can take several minutes.

## Try the optional Ollama previews

Install and start Ollama for [macOS](https://docs.ollama.com/macos),
[Linux](https://docs.ollama.com/linux), or [Windows](https://docs.ollama.com/windows).
You can also use an existing server. Download the default model on that server.

```text
ollama pull qwen3:4b
ollama ls
```

Try one of these commands after downloading the model. Replace the capture path
with the evidence saved by your run.

```text
simdlab ask -preview "Which addition implementations are available?"
simdlab interpret -preview "path/to/measurements.json" "How did SIMD compare with scalar?"
```

The default endpoint is `http://127.0.0.1:11434`. Put `-endpoint` or `-model`
before the question or file path to use a different server or installed model.
A remote endpoint receives the prompt and any catalog or evidence summaries
used by the selected command. Model-assisted steps can take tens of seconds,
can return no usable proposal, and are cancellable with Ctrl+C. Exact flags
provide a fallback for experiments. Each model-proposed plan still requires the
same human confirmation. Interpret prints verified observations and caveats as
plain text. Use `interpret -preview -json` for structured output with fact IDs.
See the [preview checks](agent-preview.md) and
[model evaluation](interpret-evaluation.md) for current limitations.

Use `-group computation` for the expression experiment. Every Make run, list
and report target forwards `ARGS`, including group shortcuts:

```bash
make help
make list ARGS="-group addition"
make run-computation ARGS="-list"
make report-addition COUNT=3 BENCHTIME=100ms ARGS="-format md -bench BenchmarkAddition/n=1024/Scalar"
```

Prefer direct commands for complex regex or paths needing careful shell quoting.
