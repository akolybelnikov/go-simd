# Local browser UI

## Run a comparison

From a terminal with `simdlab` available, run:

```text
simdlab ui
```

The command opens the browser on a private address on the machine that
launched it. Review the addition preview, then choose **Run locally**. The
result leads with the difference in median timing. The plot and table show
the individual repeats, and **Save report** downloads Markdown.

The worker keeps the report and the original measurement capture in
`artifacts/experiments/` relative to your working directory. Use
`simdlab ui -artifacts DIR` to choose another folder.

For automation, scripts and SSH, use `simdlab run addition` and the
[CLI reference](cli.md). The browser and terminal use the same addition
recipe, validated plan and measuring code.

## Open a saved run or ask for a plan

- **Open a saved run** accepts a `.measurements.json` file from the lab. The
  browser shows the host recorded in the file, material caveats, and a
  batch-size selector for a capture with several measured sizes. The browser
  passes the selected file to the local process only.
- **Ask a question** accepts prompts such as “Compare summing an array with
  a loop and with vectors.” The question fills in the preview. **Run locally**
  separately approves it. The example works offline. Start
  `simdlab ui -ollama` to use a local Ollama server for additional wording.
  The default model is `qwen3:4b`. `-model NAME` selects another installed model.

`simdlab ui -no-open` prints the private URL for manual use. Keep the
terminal running during the session. Ctrl+C stops the local app. Each launch
opens a separate loopback session. The session expires after 30 minutes and a
plan preview expires after five minutes. Use the starter button or ask again
for a fresh plan.

## Distribution

The [package workflow](../.github/workflows/release.yml) builds Windows amd64
and Linux amd64 archives and checksums. It unpacks each archive and runs the
starter on its platform with a stripped environment. A manual workflow run
produces CI artifacts. A version tag stages the packages in a draft
[GitHub Release](https://github.com/akolybelnikov/go-simd/releases) for review.
Release users extract the archive and run the bundled executable. See the
[archive quick start](binary-quickstart.txt) for the exact commands.

The public [site source](../site/index.html) serves download links, docs, and a
sample backed by a saved sum capture. In Cloudflare Pages, connect this
repository, set the production branch, leave the build command empty, and
set the output directory to `site`. Verify the release assets and checksums
before announcing downloads. [Cloudflare static HTML guide](https://developers.cloudflare.com/pages/framework-guides/deploy-anything/)

See the [delivery roadmap](local-ui-roadmap.md) for acceptance checks and
deferred work.
