# SIMD Lab model evaluation — 22 November 2026

Source: `86eb973` on `main`. This report distinguishes live inference results from model-card review. One run per case is a screening exercise, not a robust ranking or an estimate of failure probability.

## Findings at a glance

These are reply-completion counts, not semantic pass scores. Each case was run once. Models ran sequentially on this CPU-only host.

| Model | Browser replies within 65 seconds | Main finding |
| --- | --- | --- |
| `qwen3:4b` | 8/12 | Six unsolicited drafts among completed replies; poor clarification and unsupported explanations. |
| `qwen3.5:4b` | 10/12 | Cleaner replies, but substituted addition for float64 sum and omitted the requested draft. |
| `granite4.2:3b` | 10/12 | Wrong type/operation recommendations; invented an unmeasured-size winner and reused timings at that size. |
| `qwen3.5:9b` | 2/4 focused cases | Good correctness explanation; missing draft. Passed both separate CLI tool/citation integration checks. |

Qwen 3.5 9B is worth further CLI testing. None of these runs supports promoting a model as a reliable browser-agent default. Inkling remains an untested candidate requiring different hardware or an accessible inference endpoint. No paid or remote inference endpoint was benchmarked.

## What the agent needs

The browser agent explains SIMD, matches a workload to the catalog, interprets a selected measurement batch, and optionally emits a comparison draft. It does not need a general autonomous coding agent. The CLI additionally needs native function calling, including consuming tool results and citing documentation. A model passing browser JSON chat has not thereby passed CLI tool use.

The existing 12-case browser evaluation covers concepts, clamp selection, unavailable float64 sums, ambiguity, floating-point correctness, the recorded winner, insufficient repetitions, unproven bottlenecks, different hosts, unmeasured sizes, explicit comparison drafts, and follow-up draft changes.

## Reproduction

The commands below require the accompanying [evaluation test patch](evaluation-tests.patch), which adds raw record export, a configurable test-only timeout, a fixture-derived winner rubric, and CLI test model selection. The patch was tested against source revision `86eb973` and changes only test code. Apply it from the repository root before running the commands:

```sh
git apply docs/evaluations/2026-11-22/evaluation-tests.patch
```

The existing browser evaluation already supports model selection; the patch adds the export and CLI-test configuration used in this report.

Install Go 1.27.3, enable `GOEXPERIMENT=simd`, start an up-to-date Ollama, and pull the model. From the repository root on Linux/macOS:

```sh
ollama pull qwen3.5:4b
GOEXPERIMENT=simd GO_SIMD_CHAT_EVAL_MODEL=qwen3.5:4b \
  GO_SIMD_CHAT_EVAL_OUTPUT="$PWD/qwen3.5-4b-evaluation.json" \
  go test ./cmd/simdlab -run TestBrowserChatEvaluation -v -count=1 -timeout=16m
```

PowerShell:

```powershell
ollama pull qwen3.5:4b
$env:GOEXPERIMENT = 'simd'
$env:GO_SIMD_CHAT_EVAL_MODEL = 'qwen3.5:4b'
$env:GO_SIMD_CHAT_EVAL_OUTPUT = "$PWD/qwen3.5-4b-evaluation.json"
go test ./cmd/simdlab -run TestBrowserChatEvaluation -v -count=1 -timeout=16m
```

For CLI function-call integration, set `GO_SIMD_OLLAMA_INTEGRATION=1` and `GO_SIMD_OLLAMA_MODEL` to the candidate, then run:

```sh
GOEXPERIMENT=simd GO_SIMD_OLLAMA_INTEGRATION=1 \
  GO_SIMD_OLLAMA_MODEL=qwen3.5:4b \
  go test ./internal/agent/ollama -run 'TestIntegration_(RuntimeToolCall|DocumentationCitation)' -v -count=1 -timeout=5m
```

Model integration tests now accept `GO_SIMD_OLLAMA_MODEL`; the previous default remains `qwen3:4b`. Browser evaluation now asserts HTTP success and optionally records actual requests, raw provider responses, rendered replies, elapsed time, and the human grading rubric. HTTP success alone does not establish semantic correctness. The original winner rubric was incorrect for the supplied fixture: ordinary sum is 100 ns per batch and vectorized sum is about 166.67 ns, so ordinary sum wins. The rubric now derives the winner from the selected fixture. The baseline run retained the old rubric text in its raw record; all manual grading uses the actual fixture, and no model prompt changed. The raw response is necessary because the UI can substitute fallback evidence when the model returns an unreadable answer.

The optional test-only `GO_SIMD_CHAT_EVAL_TIMEOUT=3m` override can investigate a slow model with a longer budget; record it as a separate configuration and do not compare its completion rate directly with the 65-second runs. Production timeouts are unchanged.

## Environment and controls

See [environment.json](environment.json) for model manifests and runtime settings. Linux amd64, CPU-only virtual AMD EPYC 9V74, about 9.7 GiB RAM; Ollama 0.40.17 and Go 1.27.3. Models run sequentially with one loaded model and one request at a time. Browser requests use the unmodified production configuration: `think=false`, `temperature=0`, `num_ctx=16384`, `num_predict=900`, and a 65-second HTTP client timeout. The first request includes cold loading/prompt processing; subsequent requests may reuse prompt state. Download time is excluded from per-case timing. The saved sum capture describes a different measuring machine; the current host is the inference host only. No new benchmark measurements are made by the chat evaluation.

## Live baseline result

`qwen3:4b`: 8/12 requests returned HTTP 200; four exceeded the client timeout (concept, clamp selection, unsupported float64 sum, and follow-up draft). Six of the eight completed responses emitted unsolicited drafts; two of those used the scalar reference as the candidate and were rejected by the existing UI validator. The explicitly requested 80,000-value, 30-repeat sum draft had the correct fields, although the prose also asserted an unmeasured overhead explanation. The ambiguous-workload answer described an intention to ask a clarifying question without asking it. This baseline is unsuitable for promotion in the present configuration.

The completed responses were not all substantively wrong: the correctness reply recognized floating-point order effects; the variability reply rejected significance; and the winner answer identified the scalar method correctly. Their draft behavior and, in some cases, contradictory or unsupported explanations still violated the requested interaction.

See [raw baseline records](qwen3-4b.json). Test `PASS` in the log means HTTP success and no pending execution approval, not a semantic rubric pass.

## Live Qwen 3.5 4B result

`qwen3.5:4b`: 10/12 HTTP replies, with the concept and follow-up draft cases timing out. It emitted no unsolicited draft objects, a clear improvement in interaction shape over the baseline. It selected clamp correctly and rejected statistical significance and a proven bottleneck. However:

- It recommended `addition64` for a float64 **sum**, changing the requested operation.
- Instead of asking what “cleaning” meant, it recommended sum and guessed at the user’s intent.
- It stated that the requested comparison draft was prepared, but omitted the structured draft. That response also suggested an unavailable `simd-striped-8` method (the catalog has `simd-striped-2` and `simd-striped-4`).
- It speculated about the winner at 80,000 values even though that size was not measured.
- Some otherwise useful responses added unsupported or overstated explanations in `hypothesis`.

It is a better candidate for further investigation, but this configuration is not reliable enough to select as the lab’s default. See [raw Qwen 3.5 4B records](qwen3.5-4b.json).

## Live Granite 4.2 3B result

`granite4.2:3b`: 10/12 HTTP replies; concept and follow-up draft timed out. It selected clamp correctly, but recommended float32 sum for a float64 request, confidently selected copy for an ambiguous “cleaning” workload, and returned only “No.” without the requested explanation for correctness and significance. It named scalar as the winner at the unmeasured size and reused the recorded 64-value timings in an answer presented as an 80,000-value comparison. It did not supply the requested draft object. Its shorter/faster replies did not resolve the grounding or workflow problems. See [raw Granite records](granite4.2-3b.json).

## Focused Qwen 3.5 9B result

The 9B model was screened on four browser cases, not the full 12-case suite. With the same 65-second budget, clamp selection and unsupported float64 sum timed out during full-catalog requests. Correctness returned in 58.7 seconds with a clear explanation of fold order and rounding. The requested draft returned in 36.4 seconds but was missing: the model incorrectly treated the saved 64-value fixture as a limit on which comparisons can be planned and confused `3x` fixed iterations with repeat count.

It passed both live CLI integration tests: documentation retrieval with the exact source path and section (61.94 seconds), and the required native function call with the expected arguments (23.57 seconds). These are narrow integration successes, not coverage of every CLI workflow. The CLI uses different prompts, native tools, thinking enabled, a 120-second turn budget, a 4,096-token server context in this run, and the model’s default sampling settings (observed temperature 1); they cannot isolate the effect of model, context, or thinking individually.

See [focused browser records](qwen3.5-9b-focused.json) and [CLI test log](qwen3.5-9b-cli.log). The separate 180-second browser diagnostics both completed. The unsupported-float64-sum answer took **165.39 seconds** and correctly said that summation was unavailable, but incorrectly claimed `copy64` was the only float64 group (the supplied catalog also included `addition64`). The requested draft took **37.36 seconds** and repeated the same incorrect refusal as the default-budget run. More time improved availability but did not repair planning or eliminate false catalog details. See [longer-budget records](qwen3.5-9b-long-budget.json).

## Reviewed candidates that were not run

| Candidate | Published download | Assessment for this project |
| --- | --- | --- |
| Inkling 2 Mini | About 20 GB | Priority candidate on suitable hardware or a hosted endpoint. The published checkpoint exceeds this host’s RAM. |
| Qwen 3.8 27B | About 18 GB | A current larger reasoning/agent candidate; exceeds this host’s RAM. |
| Qwen3-Coder 30B | About 19 GB | More relevant if repository exploration/coding becomes central; exceeds this host’s RAM. |
| Gemma 4 E2B/E4B | About 4.6/6.6 GB and up, depending on backend | Alternative family for a later comparison; not tested in this screening. |

Inkling 2 Mini was released on 20 November 2026. Its model card describes 26.4B total parameters, 3.4B active, and 5.4B effective active parameters when accounting for repeated layers. The full weights must still reside in memory: “5.4B active” does not mean a 5.4B-sized download. Published hardware support includes Apple silicon with at least 32GB memory and DGX Spark; other GPU platforms are possible with 32GB+ VRAM recommended. The model card lists Ollama for Apple silicon, and other inference frameworks for NVIDIA. Verify the actual backend for a Linux/Windows installation before committing to it.

Published coding/tool benchmark scores justify testing Inkling, but do not show that it can interpret this lab’s evidence or honor its draft schema. The lab’s own rubric remains the selection criterion.

## Integration findings and next steps

1. Retain deterministic catalog, evidence rendering, draft validation, and run approval. Model selection should improve explanation and intent parsing without becoming the source of numerical facts.
2. Reduce unnecessary context for small local models. Explore currently sends the full catalog even for a conceptual question. The baseline’s first prompt had 5,067 tokens. Measure retrieval/intent routing separately and preserve unsupported-operation detection rather than blindly dropping catalog entries.
3. Test a clearer draft contract: explicitly distinguish measured sizes from allowable draft sizes, require `draft` to be either null or an object, and verify that a requested draft is returned as structured fields rather than prose saying it was prepared. Keep this as an evaluated prompt/schema change rather than assuming a bigger model resolves it.
4. Treat model and generation settings as a configuration profile. Browser chat fixes thinking off, while several CLI previews default to thinking on. Assess both within a declared latency budget; do not silently compare different settings.
5. Keep browser structured-answer evaluation separate from CLI native-tool integration. Extend the browser rubric with prompt injection, invalid method IDs, and numerical edge cases before promoting a candidate.
6. For a public website, add a server-side remote provider using the same context and validation contract. The current browser code has a fixed OpenAI model and explicitly rejects Ollama cloud/remote tags from the local model list. Add an explicit remote configuration rather than weakening that distinction. Keep API credentials on the server. Test remote candidates against the same fixtures before choosing one.
7. Repeat the promising candidate on the user’s actual hardware, including cold and warm requests and several runs per case. An observed timeout on this virtual CPU does not establish that the model is incapable.

## Sources checked on 22 November 2026

- [Inkling 2 Mini model card](https://thinkingmachines.ai/model-card/inkling-2-mini/)
- [Inkling announcement](https://thinkingmachines.ai/news/inkling-2-mini/)
- [Ollama Inkling listing](https://ollama.com/library/inkling-2-mini)
- [Qwen 3.5 tags](https://ollama.com/library/qwen3.5/tags)
- [Granite 4.2 3B](https://ollama.com/library/granite4.2:3b)
- [Qwen 3.8](https://ollama.com/library/qwen3.8)
- [Qwen3-Coder](https://ollama.com/library/qwen3-coder)
- [Gemma 4](https://ollama.com/library/gemma4)

## Validation and evidence

The relevant deterministic tests passed with `GOEXPERIMENT=simd`:

```sh
go test ./cmd/simdlab ./internal/agent/... ./internal/labeval
```

Live browser test failures at the 65-second setting are retained as findings, not hidden as skipped cases. The longer-budget Go test reports PASS because HTTP requests completed; its draft answer still failed the task. The CLI tests assert the required native call and documentation attribution.

This directory contains the report, raw JSON request/response records, per-suite logs, runtime/model manifests, and the evaluation test patch. Production provider defaults and prompts were used for the browser comparisons.
