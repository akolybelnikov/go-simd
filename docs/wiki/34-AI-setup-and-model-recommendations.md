# AI setup and model recommendations

Choose an assistant before sending your question. SIMD Lab can help you prepare
context for an assistant you already use, connect to local Ollama, or use your
OpenAI API account. Comparisons and the ordinary lab commands work without AI.

## Choose how to ask

| Choice | What you need | What happens |
| --- | --- | --- |
| Copy to another assistant | Your existing assistant | Review and copy the question and lab context, then paste them into that assistant. |
| Local · Ollama | Ollama running with an installed model | Choose a Chat model in the panel. Local chat is experimental. |
| Remote · OpenAI | An API key in the terminal that launches the lab | The configured model receives your question, recent exchanges, and lab context. Your API account is billed. |

Both **Ask an agent** buttons open this same choice. In Results, select the batch
size first. [How to ask and review a suggested comparison](33-Chat-with-an-agent).

## Local Ollama: no browser-chat recommendation yet

The local models tested so far made mistakes when choosing operations,
interpreting evidence, or preparing comparison drafts. Installing a model does
not mean it is recommended. These findings apply to the tested configurations;
they are not a verdict on every local model or every computer.

We screened three smaller models on 12 browser cases each, and Qwen 3.5 9B on
four focused browser cases. The questions covered operation/type matching,
ambiguity, numerical correctness, selected results, and comparison drafts.

| Exact model tag | Findings from 22 November 2026 |
| --- | --- |
| `qwen3:4b` | Unrequested drafts, incorrect explanations, and timeouts. Experimental. |
| `qwen3.5:4b` | Cleaner replies, but wrong operation selection and missing drafts. Experimental. |
| `granite4.2:3b` | Wrong type/operation recommendations and claims about unmeasured sizes. Experimental. |
| `qwen3.5:9b` | Useful correctness explanation, but failed browser planning. Passed two separate CLI tool/citation checks; that is not a browser-chat recommendation. |
| `inkling-2-mini` | Reviewed, not run. The approximately 20 GB published checkpoint exceeded our roughly 10 GB test host's RAM. No recommendation yet. |

The test host was CPU-only. First full-catalog requests often exceeded the
65-second limit. More time improved reply availability but did not fix missing
drafts or all catalog errors. A completed request is not a correct answer.
Our response times do not predict performance on your GPU or another machine.

[Full findings, settings, and raw replies](https://github.com/akolybelnikov/go-simd/tree/main/docs/evaluations/2026-11-22).

### Future recommendations

This page will name exact download tags only as recommendations for the
workflows they have passed. Each recommendation needs a test date, model
revision/quantization, runtime/backend, hardware/memory guidance, generation
settings, and semantic results. Browser chat, planning, and CLI tool use are
evaluated separately. Larger-model testing remains open; no new recommendation
is implied by the model selector.

## Try an installed local model

Install and start [Ollama](https://ollama.com/download), then launch:

```text
simdlab ui -ollama
```

Select **Local · Ollama**, then choose an installed **Chat model**. Use
**Refresh** after installing or removing a model or restarting Ollama. Labels
show **Evaluated — experimental** or **Not evaluated**. No model is downloaded
automatically. If there are no installed models, use copy mode or consult the
model's hardware requirements before choosing an experimental download.

To select an experimental model at launch, replace `INSTALLED_MODEL` with its
exact installed name:

```text
simdlab ui -ollama -model INSTALLED_MODEL
```

This sets the initial chat selection and the separate local intent resolver.
Changing **Chat model** in the panel affects chat only. It starts a new thread
and stops waiting for an old reply. It does not change other tabs, CLI sessions,
or the launch-time intent-resolver model. The question remains available after
connection and model errors.

The lab keeps cloud/remote Ollama entries out of the local dropdown. This
feature talks to the local Ollama service, not an arbitrary browser-supplied
server.

### CLI model selection

Ordinary guided questions and deterministic comparisons do not require a model:

```text
simdlab ask
simdlab run addition
```

For experimental model assistance, select the model explicitly:

```text
simdlab ask -model INSTALLED_MODEL
simdlab ask -preview -model INSTALLED_MODEL "Which addition implementations are available?"
simdlab interpret -preview -model INSTALLED_MODEL CAPTURE_FILE "Explain these results."
simdlab experiment -preview -model INSTALLED_MODEL "Compare scalar and SIMD addition."
```

Put flags before the question or file. There is no automatic `qwen3:4b` fallback.
Model-assisted calls identify the selected model. CLI `-endpoint`, `-think`, and
`-timeout` options vary by command; use the command's `-h`. A nonlocal endpoint
receives the context sent by that command. A CLI timeout or thinking setting does
not change browser chat settings. Browser chat currently uses thinking off and
a 65-second reply budget.

[CLI reference](https://github.com/akolybelnikov/go-simd/blob/main/docs/cli.md).

## Connect OpenAI with an API key

Set the key in the terminal that starts the lab. Choose a model compatible with
the OpenAI Responses API, structured JSON output, and the configured low
reasoning effort. The default remains `gpt-6.1-sol`; the local-model evaluation
does not certify this or any other remote model.

**Windows PowerShell**

```powershell
$env:OPENAI_API_KEY = 'your-api-key'
simdlab ui -openai-model gpt-6.1-sol
```

**macOS or Linux**

```sh
export OPENAI_API_KEY='your-api-key'
simdlab ui -openai-model gpt-6.1-sol
```

Replace the example key with your own key. Change `-openai-model` to configure a
different compatible model, then choose **Remote · OpenAI · MODEL** in the
panel. This option is separate from `-model`, which selects local Ollama.
The remote model is configured by the launching process; the UI displays that
choice. Other frontier-provider adapters are not implemented yet.

The key stays in the launching process. Do not put it in the chat question,
a URL, or a model name. Requests use your API account; an assistant subscription
is not the API configuration. The lab permits up to 24 OpenAI questions per
session. Review **what will be sent** before Send: it includes the question,
recent exchanges, catalog context, and selected measurement summaries. Raw
arrays and file paths are omitted from those summaries. Typed text is sent as
entered. You can return to copy mode at any time.

The current public site offers documentation, downloads, and a sample; it does
not run hosted AI chat. Public hosted inference needs its own backend.

## Repeat the model checks

From a source checkout with Go 1.27.3 and `GOEXPERIMENT=simd`, select the installed
model explicitly. Live checks are opt-in and are not part of routine CI.
For browser evaluation on macOS/Linux:

```sh
GOEXPERIMENT=simd GO_SIMD_CHAT_EVAL_MODEL=INSTALLED_MODEL \
  GO_SIMD_CHAT_EVAL_OUTPUT="$PWD/chat-evaluation.json" \
  go test ./cmd/simdlab -run TestBrowserChatEvaluation -v -count=1 -timeout=16m
```

For CLI tool/citation checks:

```sh
GOEXPERIMENT=simd GO_SIMD_OLLAMA_INTEGRATION=1 \
  GO_SIMD_OLLAMA_MODEL=INSTALLED_MODEL \
  go test ./internal/agent/ollama -run 'TestIntegration_(RuntimeToolCall|DocumentationCitation)' -v -count=1 -timeout=5m
```

In PowerShell, set each variable with `$env:NAME = 'value'` before the `go test`
command. The browser test can export raw replies and rubrics; a Go test PASS
alone is not semantic approval. A test-only `GO_SIMD_CHAT_EVAL_TIMEOUT=3m` override
must be recorded as a different evaluation budget. It does not change the app's
timeout or prove a model suitable for interactive use. Current source includes
the evaluation helpers; the patch accompanying the historical report is only
needed when reproducing that earlier source revision.

## If a model fails

- **Ollama unavailable:** start it, then Refresh. You can continue in copy mode.
- **Model missing:** install the intended model yourself after checking its
  requirements, or choose another installed model. The lab does not pull models.
- **Slow reply:** use Stop. Try another configured model, copy the question and
  context, or create the comparison directly in Explore.
- **Wrong operation or unsupported type:** check the catalog and correct the
  request. Do not treat a similarly named operation as a substitute.
- **No usable draft:** use Explore to set up the comparison. A reply saying
  “prepared” is not enough; a valid draft has an **Edit suggested comparison**
  action and still requires review and **Run locally**.

The lab displays recorded evidence alongside explanations. Timing alone does
not establish numerical equivalence, statistical significance, a bottleneck,
or a production speedup. [Understanding results](06-Understanding-results).
