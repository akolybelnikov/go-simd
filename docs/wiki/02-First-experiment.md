# Your first addition comparison

[Install the CLI](https://github.com/akolybelnikov/go-simd/blob/main/docs/cli.md#install-the-cli-from-source).

## Run in the browser

Open the lab with `simdlab ui`. In **Compare**, review the addition starter and choose **Run locally**. **Results** opens when the comparison finishes. Use **Save report** to download Markdown.

## Run from a terminal

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

## Read and share the result

Here is an **illustrative** result for two arrays of 1,024 values. Every
timing is for adding the two whole arrays.

| Method | Repeat 1 | Repeat 2 | Repeat 3 | Median |
| --- | ---: | ---: | ---: | ---: |
| Ordinary addition | 101 ns | 100 ns | 99 ns | 100 ns |
| Vectorized addition | 82 ns | 78 ns | 80 ns | 80 ns |

In this example, vectorized addition took 20% less median time per batch.
Both methods produce one result per input position. The [addition data
example](32-Operation-measurement-guide#addition) shows what those values
look like. The Markdown report includes the repeat ranges and measuring
host. Use `simdlab inspect -text` with the printed measurement path to see details.

## Compare another method or size

In **Explore**, choose **Addition**, select a method and size, then preview and approve the comparison. Run the same size again with a different method to compare partial vectors with full vectors and a scalar tail. Then keep the method fixed and vary the size.

Write down which method took less time, the ratio or percentage difference, and how the repeats varied. Include the measuring host and batch size. See [Understanding results](06-Understanding-results) and [Experiment recipes](07-Experiment-recipes).
