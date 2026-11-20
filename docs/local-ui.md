# Local browser UI

## Run a comparison

From a terminal with `simdlab` available, run:

```text
simdlab ui
```

The command opens **Explore** at a private address on the machine that
launched it. Search or browse the operations available on this machine.
Select an operation and a method, then enter any whole number from 1 to
16,777,216 for the values per input array. The preset sizes are suggestions.
Choose **Preview comparison** to bring those settings into **Compare**.
If the values do not fill the last vector, the preview explains how the
selected method handles the remainder. For addition, this can be a partial
vector or a scalar tail, depending on the method.

Each operation card explains what it does and gives an example of where it
is used. **Read the measurement guide** opens its wiki section in a new tab.
The **Wiki** link in the header opens the full bundled wiki. Both work offline.

**Compare** also has a ready-made addition starter. Review the preview and
choose **Run locally**. When the comparison finishes, **Results** shows the
difference in median timing and the individual repeats. **Save report**
downloads Markdown.

The worker keeps the report and the original measurement capture in
`artifacts/experiments/` relative to your working directory. Use
`simdlab ui -artifacts DIR` to choose another folder.

For automation, scripts and SSH, use `simdlab run addition` and the
[CLI reference](cli.md). The browser and terminal use the same addition
recipe, validated plan and measuring code.

## Open a saved run or ask for a plan

- **Open a saved run** in **Results** accepts a `.measurements.json` file from the lab. The
  browser shows the host recorded in the file, material caveats, and a
  batch-size selector for a capture with several measured sizes. The browser
  passes the selected file to the local process only.
- **Set up another comparison** in **Compare** offers one-click previews for summing an
  array and adding arrays of 8,192 values. Expand **Describe your own
  comparison** to enter a request in your own words. Review the preview,
  then choose **Run locally** to start measuring. The examples work offline. Start
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
