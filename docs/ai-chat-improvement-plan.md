# Plan: explain AI choices and make chat model selection explicit

Status: guidance and configuration implemented, 22 November 2026. Based on the [model evaluation findings](evaluations/2026-11-22/README.md). The design below records the intended behavior; this status table distinguishes delivered work from future evaluation.

| Work | Status |
| --- | --- |
| Shared dialog guidance, dedicated wiki page, navigation, and quick starts | Implemented. |
| Installed local model selector, refresh, assessment metadata, and per-request selection | Implemented. No local browser-chat model is marked recommended. |
| Explicit CLI selection and offline guided fallback | Implemented. Silent `qwen3:4b` selection removed. |
| Evaluation export/configuration improvements | Integrated into test code; the committed historical patch remains with the findings. |
| OpenAI startup model option and API-key guide | Implemented through `-openai-model`; the UI shows the configured model. |
| Larger-model semantic qualification and new local recommendations | Open; needs suitable hardware or a configured inference endpoint. No additional quality claims were made. |
| Additional frontier-provider adapters and hosted inference backend | Future work. Only the existing OpenAI adapter is exposed. |

Validation covers the full Go suite, vet, focused race checks, and real JavaScript DOM tests against the HTTP backend with stubbed Ollama responses. A graphical browser review was blocked by this environment's browser restrictions; DOM tests do not certify pixel layout.

## Intended user experience

When someone clicks either **Ask an agent** button, they should immediately understand their choices, the current state of local-model testing, and where to find setup guidance. They should be able to identify the model that will answer before sending a question. Later recommendations should help them choose a model that has passed the checks for the task they want to perform.

Keep the first explanation short and place it inside the existing chat dialog. Use the wiki for installation, test details, model recommendations, and provider setup. Avoid a second blocking popup or a confirmation checkbox just to read or ask a question.

The immediate message is that **no local model is currently recommended for browser chat**. The tests found specific failures in the evaluated model/configuration combinations. They do not establish that Ollama itself, all local models, or every installation of the tested models is incapable. The 9B model's two successful CLI integration checks are promising but do not qualify it for browser-chat or planning recommendations.

## What already exists

| Area | Current behavior | Change needed |
| --- | --- | --- |
| Explore and Results | Both buttons open the same `agent-dialog` through `openAgent()` | Add the same concise orientation at both entry points, with provider-specific detail after selection. |
| Provider selection | `agent-provider` already offers copy mode, configured OpenAI, and one explicitly configured Ollama model | Extend this control; add a separate local model selector rather than building a parallel chat panel. |
| Local CLI options | `ask`, `ask -preview`, `interpret -preview`, `experiment -preview`, and `ui` already accept `-model` | Make selection and effective defaults consistent and document the options accurately. |
| Local defaults | Several commands and the UI's intent-resolver setup fall back to `qwen3:4b` | Stop presenting that fallback as an endorsed model; subsequently replace silent model selection with explicit configuration. |
| UI local setup | `ui -ollama -model MODEL` registers that model for chat; `-ollama` alone can configure an intent resolver without adding an Ollama chat option | Make enablement, selection, and the scope of the selected model understandable. |
| Installed models | `chatProvider.localModels()` already reads `/api/tags` and filters out cloud/remote entries | Reuse this server-side discovery for the local model dropdown. |
| Remote chat | OpenAI API-key setup works, but the chat model is a Go constant | Preserve the working setup instructions now; add remote model configuration before describing it as selectable. |
| Wiki | Page 33 explains chat, especially existing OpenAI setup; Markdown is embedded into the binary | Keep page 33 as the usage guide and add a dedicated setup/recommendations page. |

## Phase 1 — explain the choices before changing the integration

### 1. Add orientation inside the shared chat dialog

Files:

- `cmd/simdlab/ui/index.html`: add an information area below the dialog description and beside the existing setup link, before the question field.
- `cmd/simdlab/ui/app.js`: populate the area in `openAgent()` and update its provider-specific message in `updateAgentControls()`.
- `cmd/simdlab/ui/style.css`: make the message readable on small screens without making the form unnecessarily tall.

Suggested opening copy:

> Choose an assistant below. Local Ollama chat is experimental: the models tested so far made mistakes when choosing operations and preparing comparisons. We do not yet recommend a local model for browser chat. See AI setup and model recommendations.

Keep this short orientation visible when the dialog opens from either Explore or Results. When someone selects local Ollama, show the selected model's specific status below the model control. Copy mode should explain that the user can paste the lab context into their existing assistant. OpenAI should retain its relevant sharing and API-account information; do not apply a local-test verdict to an untested remote model.

Use an ordinary informational region with an accessible heading and link. The user can continue immediately. Do not show benchmark tables, raw failure traces, or long hardware explanations in the dialog.

### 2. Add a dedicated wiki page

New file: `docs/wiki/34-AI-setup-and-model-recommendations.md`.

Proposed contents:

1. **Choose how to ask**: copy to an existing assistant, local Ollama, or currently supported remote OpenAI chat. Explain that measurement and deterministic lab commands remain available without a model.
2. **Local Ollama: current status**: no recommended local browser-chat model yet; evaluated candidates remain experimental.
3. **What we tested**: a short account of operation selection, numerical correctness, interpreting the selected result, and preparing/following up on drafts. Include a compact model/status table and link to the full findings in `docs/evaluations/2026-11-22/`.
4. **What the results mean**: wrong operations, unsupported numerical types, missing or unsolicited drafts, and unsupported timing explanations. Explain that a completed HTTP request is not a correct answer, and that CPU-only timing does not predict a user's GPU performance.
5. **Choosing and installing a local model**: explain the installed-model selector and explicit CLI selection as those features ship. Separate install instructions for approved recommendations from instructions for experimenting with an arbitrary model.
6. **Recommended local models**: initially state that no browser-chat recommendation is available. Later add exact model tags, supported workflows, tested hardware/backend, relevant memory requirements, settings, test date, and an evidence link. Keep Qwen 3.5 9B's limited CLI results separate from chat recommendations. Mark Inkling as reviewed but not run in our environment.
7. **Remote models and API keys**: link to the working OpenAI instructions initially. Expand this section when configurable remote models and additional provider adapters are implemented. Label future integrations as planned rather than supported.
8. **When a reply fails or takes too long**: show the available recovery actions, such as Stop, copy the question/context, choose another configured model, or create the comparison directly in Explore.

The public wiki summary should be short enough for a user choosing an assistant. The full evaluation remains the technical evidence source, rather than being duplicated throughout the documentation.

### 3. Connect the guide to all relevant entry points

| File | Planned change |
| --- | --- |
| `docs/wiki/33-Chat-with-an-agent.md` | Add an early link to page 34 and a short local-model status note. Keep this page focused on asking, following up, selecting a result, and reviewing drafts. Preserve existing OpenAI setup until the dedicated setup guide incorporates it. |
| `docs/wiki/Home.md` | Link the AI setup guide from the chat feature description and the guide table. |
| `docs/wiki/_Sidebar.md` | Add the new page next to Chat with an agent. |
| `cmd/simdlab/ui_wiki.go` | Add AI setup to the embedded wiki's navigation so users can return to it from other documentation pages. |
| `README.md` | Add a brief status/setup link beside Ask an agent; avoid advertising a tested model as a default recommendation. |
| `docs/local-ui.md` | Describe the opening notice, the available assistant choices, the selected model, and the guide link. |
| `docs/cli.md` | Explain that historical `qwen3:4b` defaults are not recommendations. Replace the unqualified instruction to download the default model with explicit experimental selection and a link to page 34. |
| `docs/wiki/03-Command-registry.md` | Show how model selection relates to the existing commands, using syntax supported by the implementation. |
| `docs/wiki/09-Troubleshooting.md` | Add short guidance for no model installed, Ollama unavailable, timeout, and no usable draft, with links back to setup. |
| `site/index.html` | Add an AI setup link alongside relevant documentation/download links. The current static site should not claim to provide hosted model inference. |

`docs/wiki/wiki.go` already embeds `*.md`, so the new page does not require a new page-specific route. Rebuild the binary to distribute the updated embedded docs. Verify links in the embedded wiki and repository Markdown. Publishing the separate GitHub Wiki, if used, is a distinct step: the workflows inspected do not automatically synchronize `docs/wiki` into the `.wiki.git` repository.

Phase 1 is complete when both chat buttons lead to the explanation and working setup links, and no quick-start documentation presents an unsuccessful candidate as recommended.

## Phase 2 — explicit local model selection and shared recommendation metadata

### 4. Maintain model status in one place

Introduce a small internal model catalog, for example `internal/modelcatalog/`, shared by the chat backend and UI metadata. Keep the first implementation limited to what the UI and recommendation guide need.

Each entry should identify the provider, exact model tag, workflow, assessment status, evaluation date/evidence, and any tested configuration/hardware note. Separate assessment from availability:

- Assessment: **recommended for this workflow**, **evaluated — experimental**, or **not evaluated**.
- Availability: installed, not installed, or local service unavailable.

There are initially no recommended local browser-chat entries. The registry may describe known candidates and failures without endorsing them. A model can be promising for CLI tool use while remaining experimental for browser chat.

Avoid independent hardcoded recommendation lists in JavaScript, CLI help, and wiki examples. Use the catalog for machine-readable status, and check user-facing tables/links against it when recommendations change.

### 5. Add a local model selector to the existing chat form

Relevant files: `cmd/simdlab/ui/index.html`, `ui/app.js`, `ui/style.css`, `ui_chat.go`, `ui_chat_provider.go`, `ui_http.go`, and `ui.go`.

Proposed interaction:

1. **How to ask** chooses Copy, Local Ollama, or a configured remote provider.
2. Selecting Local Ollama reveals a **Chat model** dropdown and a Refresh action.
3. The server discovers installed local models using the existing `/api/tags` integration, then combines availability with assessment metadata.
4. Show a Recommended group only when qualifying recommendations exist. Today, show “No local browser-chat model is recommended yet,” followed by installed experimental models. Unknown installed models are clearly marked Not evaluated.
5. For a future recommended model that is missing, show setup instructions and a copyable installation command. Keep it unavailable for sending until installed; selecting it must not trigger a large automatic download.
6. Show the chosen tag and its workflow-specific status before Send. If a model is removed or Ollama stops, preserve the typed question and offer Refresh, setup guidance, and copy mode.

Use the server as the authority for allowed local selections. Extend the authenticated catalog response or add a bounded authenticated model-list endpoint; do not have browser JavaScript talk directly to arbitrary Ollama endpoints. Keep cloud/remote tags outside the Local category. Changing model must cancel an in-flight reply and start a new chat thread, as provider changes already do.

Scope the first dropdown explicitly to chat. `app.model` also exists for a separate `/api/ask` intent-resolver path: do not accidentally change it globally or imply that selecting a chat model has configured every AI feature. Any subsequent unification must carry an explicit model choice per request and must not silently restore the old `qwen3:4b` fallback.

### 6. Clarify CLI flags and remove silent selection of an unendorsed model

Files: `cmd/simdlab/ui.go`, `guide.go`, `ask.go`, `interpret.go`, `experiment.go`, and `main.go`.

Retain the existing local `-model` syntax. Suggested semantics after this phase ships:

```text
simdlab ui -ollama
simdlab ui -ollama -model INSTALLED_MODEL
simdlab ask -model INSTALLED_MODEL
simdlab ask -preview -model INSTALLED_MODEL "Which addition implementations are available?"
simdlab interpret -preview -model INSTALLED_MODEL CAPTURE_FILE "Explain these results."
```

- `ui -ollama` enables local discovery; chat still opens on Copy until a model is deliberately selected. It must not quietly select an experimental model.
- `ui -ollama -model MODEL` supplies the initial local chat selection and clearly reports which local features it configures.
- Ordinary deterministic `ask` interactions should work without a model. Construct the model client only when needed; if no model was selected, explain how to select one instead of silently contacting a default.
- Model-dependent preview commands should identify a missing or unavailable model and print a useful configuration example. Deterministic experiment paths must not acquire a model requirement.
- Print the effective model for model-assisted CLI calls and keep help/output consistent across commands. Keep thinking and timeout settings explicit where supported; document that different commands currently use different defaults.

This is a behavior change, not merely a documentation edit. Cover existing offline starter flows and include a release note explaining the replacement of the historical fallback. Do not publish proposed commands as current behavior before implementing them.

## Phase 3 — earn local-model recommendations through broader testing

Use the committed evaluation as the baseline. The export/configuration improvements are currently available as a [test patch](evaluations/2026-11-22/evaluation-tests.patch); apply or integrate that patch before relying on its additional environment variables.

Before giving a model a browser-chat recommendation:

- Exercise the full browser corpus and review answer correctness, clarification, supported operation/type matching, selected-size fidelity, and structured drafts. Record transport availability separately.
- Require valid requested drafts and no unsolicited drafts in the relevant cases. Check that measured sizes do not become artificial limits on new draft sizes.
- Test follow-up changes, unsupported requests, and the difference between hypothesis and measured evidence.
- Run CLI native-tool and citation checks separately. Two successful integration tests cannot establish that an entire CLI workflow or browser chat is reliable.
- Repeat important cases and measure cold and warm requests. Record model revision/quantization, runtime/backend, RAM/GPU, context, thinking, output limits, timeout, and test date.
- Test larger candidates on appropriate hardware or an explicitly configured hosted endpoint. Keep a local recommendation and its hardware guidance distinct from hosted inference results.

Publish a short recommendation entry only after the relevant workflow passes its semantic review. Retain an honest “no recommendation yet” state if it does not. For local explanation/planning quality, separately evaluate smaller context, clearer draft rules, and the existing typed-tool approach; larger weights alone did not solve the observed planning failures.

## Phase 4 — configurable frontier models and expanded API guidance

Start with the provider already implemented: OpenAI. Its key setup works today; its chat model is hardcoded in `cmd/simdlab/ui_chat_provider.go`.

1. Add a provider-specific startup option such as `simdlab ui -openai-model MODEL`. Keep existing `-model` meaning Ollama for compatibility; document the distinction in help.
2. Pass the configured remote model through provider construction and expose the actual selection in UI metadata instead of a duplicated model-name string.
3. If remote model selection is offered in the UI, populate it from server-configured model choices. Add another provider only when its API adapter and the lab's response contract are implemented and evaluated.
4. Expand page 34 with provider-specific setup, supported model choices, environment-variable examples for supported shells, API-account billing, what context is sent, and how to switch back to copy mode. Replace the fixed-name text in page 33 and `docs/local-ui.md`.
5. Keep keys in the launching process/server configuration, out of browser storage, URLs, model-list responses, and evaluation records. Display provider/model and sharing information before Send.
6. Validate each adapter against the same catalog, evidence, and draft checks. Do not transfer a local-model recommendation or an OpenAI result to a different provider automatically.

A future public website with in-app API inference needs a backend and credential handling appropriate to that deployment. Treat that as a separate implementation from adding model flags to the local CLI; the current static site only links to documentation/downloads and shows a sample.

## Acceptance checks and delivery order

| Deliverable | Evidence of completion |
| --- | --- |
| 1. Dialog explanation and wiki guidance | Both entry points display the short orientation; local selection shows the current assessment; links render in the embedded wiki; current OpenAI setup remains discoverable. |
| 2. Model selection/configuration | Exercise installed, missing, unknown, and disconnected local models; verify CLI initial selection, UI changes, conversation reset, and no silent use of a different model. |
| 3. Recommendations backed by tests | Every recommended workflow has dated semantic results and recorded configuration/hardware; incomplete, reviewed-only, or timed-out cases are labeled accurately. |
| 4. Configurable remote model and later adapters | Configured model is actually sent and displayed; supported setup instructions match implemented flags/providers; existing evidence/draft handling still applies. |

Use focused backend tests with stubbed provider/model-list responses for selection and failure states, plus browser checks for both dialog entry points, the model dropdown, mobile layout, and keyboard access. Check Markdown links and embedded page rendering. Keep real-model evaluation opt-in rather than downloading models or spending API credits in routine CI.

The first implementation should deliver **the opening explanation, page 34, navigation links, and cleanup of misleading model-download/default wording**. Model discovery/configuration comes next, followed by evidence-backed local recommendations and expanded remote-provider setup. This gives users useful guidance immediately without announcing unimplemented integrations or premature recommendations.
