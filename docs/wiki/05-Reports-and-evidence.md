# Reports and evidence

A report is a readable summary. The measurement capture keeps the repeats and the context needed to reopen or interpret that summary.

## Save and reopen a run

In **Results**, choose **Save report** to download Markdown. The lab also saves a report and a measurement file under the output directory. Use **Open a saved run** to reopen a measurement file. It retains the recorded host even if you open it on another machine.

From a terminal, use:

```text
simdlab inspect -text "path/to/measurements.json"
```

For terminal sweeps, choose plain text with `-format txt`, Markdown with `-format md`, or CSV with `-format csv`. CSV is useful for plotting or comparing rows in a spreadsheet.

## Know what the evidence says

- The median describes the middle repeat for one method and one batch size. The individual repeats show how much the observations vary.
- The recorded machine and vector width tell you where and how the batch was measured.
- The fixture and numerical settings explain what data and arithmetic the comparison used.
- If a run only saved some repeats before it stopped, describe it as an incomplete run and keep that limitation with the numbers.

For a reproducible comparison, share both the readable report and the original measurement capture. See [Understanding results](06-Understanding-results) for the units and percentage calculation.
