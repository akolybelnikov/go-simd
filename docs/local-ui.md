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
Set **Repeats per method** from 1 to 50. The default is 3, with a 100 ms
measurement target per repeat. For a follow-up comparison of float32 addition,
select **Vectorized addition (partial vectors)** and enter **4,096** values
(type `4096`) and **30** repeats.
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
downloads Markdown with medians, observed ranges, and a row per repeat.
Repeats are collected in one benchmark run. Check numerical correctness and
generated instructions separately from these timings.

The worker keeps the report and the original measurement capture in
`artifacts/experiments/` relative to your working directory. Use
`simdlab ui -artifacts DIR` to choose another folder.

For automation, scripts and SSH, use `simdlab run addition` and the
[CLI reference](cli.md). The browser and terminal use the same addition
recipe, validated plan and measuring code.

## Open a saved run

- **Open a saved run** in **Results** accepts a `.measurements.json` file from the lab. The
  browser shows the host recorded in the file, material caveats, and a
  batch-size selector for a capture with several measured sizes. The browser
  passes the selected file to the local process only.

## Ask an agent

Use **Ask an agent about your use case** in **Explore** to describe the type
of data you have, the operation you need, and any precision requirements.
For example, ask which operation fits clipping float32 sensor readings to a
fixed range.

In **Results**, select a batch size if the capture has several, then use
**Ask an agent about these results**. Ask what the repeats suggest or what
comparison would help answer a follow-up question. The copied context contains
the selected size, measured medians and ranges, repeat counts, and caveats.

Review the text, copy it, and paste it into an AI assistant you choose.
SIMD Lab prepares the text locally. It copies a summary without file paths or
raw input arrays. Use the local controls to run any follow-up comparison.

`simdlab ui -no-open` prints the private URL for manual use. Keep the
terminal running during the session. Ctrl+C stops the local app. Each launch
opens a separate loopback session. The session expires after 30 minutes and a
plan preview expires after five minutes. Use Explore or the starter button
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
