# Working with arrays

The batch size is the number of values in each input array. A dot product with 1,024 values uses two arrays of 1,024 values each. Its output is one number. Elementwise addition produces an output array of 1,024 values.

## Match the question to the output

| Question | Operation | Output |
| --- | --- | --- |
| How do aligned values combine? | Addition | One result per position |
| Which values pass a cutoff? | Byte or packed predicate mask | One decision per position |
| What is the total? | Array sum | One result for the whole array |
| What are the totals in each window? | Fixed-block sum | One result per group of eight |
| How much work does copying take? | Exact copy | The same bits in another buffer |

## Choose informative sizes

Use one size close to the measured vector width and one substantially larger. If the width is eight, sizes 7, 8, and 9 tell you about a short vector, a complete vector, and a vector followed by a remainder. A width of eight float32 lanes does not imply eight float64 lanes.

A larger batch can change where the data fits in the memory hierarchy. Keep the method and fixture fixed when exploring size alone. The [experiment recipes](07-Experiment-recipes) offer more comparisons.

## Map your data to a benchmark

| Your data | Comparable setup in the lab |
| --- | --- |
| Two 480-sample audio buffers that start at the same time | **Addition** at 480 values per input array. Each result is the sum of two samples at the same position. |
| 12,000 sensor readings with a cutoff | **Threshold** keeps a 12,000-value output. A byte mask records 12,000 decisions as bytes. A packed mask records the same decisions in 1,500 bytes. |
| A row of 64 features and 64 weights | **Dot product** at 64 values per input array. The batch returns one score for the row. |

Use the number type your application uses. For comparisons or clipping,
think about how often values pass and whether they arrive in groups. A
clustered or shuffled fixture can help explore that difference. The lab
generates its own inputs, and your application may also spend time loading,
converting, or moving its data. [Fixture options](03-Command-registry#fixture-options) · [Input and output examples](32-Operation-measurement-guide)
