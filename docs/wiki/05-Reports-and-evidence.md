# 05 — Reports and retained evidence

The report command measures first, summarizes its individual observations by
median, and renders the same comparison model in the selected format.

## Artifact layout

```text
artifacts/
  addition/
    benchmark-<UTC timestamp>-<unique suffix>.raw
    benchmark-<UTC timestamp>-<unique suffix>.md
  computation/
    benchmark-<UTC timestamp>-<unique suffix>.raw
    benchmark-<UTC timestamp>-<unique suffix>.csv
```

The report extension depends on that invocation's `-format`. Matching stems tie
the summary to its raw companion. `-output` changes the root; the command adds
the group directory. Unique names preserve previous successful captures.

## Format reference

| Property | TXT | Markdown | CSV |
| --- | --- | --- | --- |
| Selection | Default or `-format txt` | `-format md` | `-format csv` |
| Saved table | Aligned text | Header, alignment row, data rows | Header and numeric records |
| Timing precision | Two decimals, ns/op | Two decimals, ns/op | Two decimals, ns/op |
| Changes | Signed one decimal and `%` | Signed one decimal and `%` | Numeric percentage points, one decimal |
| Missing value | Em dash | Em dash | Empty field |
| Environment metadata | Below table | Literal text below table | Raw companion only |
| Terminal comparison | Readable text | Readable text | Readable text |

CSV uses a proper encoder, including quoting for commas, quotes, and embedded
newlines. A value such as `-20.0` means 20% less time than scalar. It is not the
ratio `-0.2`; apply a spreadsheet percent style only after converting units as
appropriate for that spreadsheet.

### Illustrative CSV

```csv
ELEMENTS,Scalar ns/op,SIMDPartial ns/op,SIMDPartial change
1024,100.00,80.00,-20.0
```

These values are invented to demonstrate the schema. An unmeasured scalar
baseline does not become zero and does not produce an infinite speedup.

## What raw files retain

- UTC capture date and the report process's Go version and platform.
- Parent GOMAXPROCS and observed float32 lane count.
- Exact child `go test` command, selection, repetition count, and benchtime.
- `GOEXPERIMENT`, `GOAMD64`, `GOARM64`, and `GOTOOLCHAIN` from `go env`.
- Child Go benchmark OS/architecture/CPU headers where available.
- Every measured repetition, iteration count, displayed ns/op, throughput,
  allocation counters, and tool output.

Keep the distinction between parent metadata and child-reported facts. A parent
lane count is context, not proof of a different process's capabilities. The
current raw format does not provide complete source identity, an emulation flag,
or an exact total duration recoverable from rounded ns/op. Record source revision
and local changes separately; do not fabricate missing evidence.

## Internal observations versus exported files

M1 introduces typed observations with case/run/environment identity, individual
repetitions, optional metrics, and backend/harness provenance. Fresh generated
inputs have bit-level fingerprints in that internal model; historical text
imports do not assume those fingerprints.

The current TXT/MD/CSV exports are wide summary tables, not a public raw-data
schema. The complete internal identity model is not serialized into these
formats. JSON/NDJSON export, a saved-run analyzer, and an offline re-render CLI
are not available yet. Changing `-format` starts a new measurement run.

## Inspect reports in PowerShell

```powershell
Get-ChildItem ./artifacts/addition

# Pick a path printed by your completed command.
$report = Get-ChildItem ./artifacts/addition/*.csv | Sort-Object LastWriteTime -Descending | Select-Object -First 1
Import-Csv $report.FullName | Format-Table
```

An empty CSV field is missing, not a measured zero. Use explicit numeric parsing
if doing arithmetic; `Import-Csv` starts with text-valued fields. Preserve the
raw companion when making a chart or copying a summary into an issue.

## Failure behavior

Unsupported `-format`, unknown group, nonpositive count, and unexpected
positionals fail before report artifacts are created. Invalid Go timing/regex
arguments can fail in the child runner after the raw file exists.

Benchmark failure leaves raw output and reports its path. Empty selections also
retain raw and produce no successful summary. With `-group all`, an earlier
group can succeed before a later group fails; inspect each group's artifacts
and the command's final exit status.

Disk/output failures can prevent a report from being saved; retain any raw path
reported in the error. There is no finalized versioned manifest or general
process-cancellation artifact protocol at this milestone.

## Share an evidence bundle

Include the summary and raw companion, source revision/local-change notes,
your experiment question, important environment settings, and limitations.
The laboratory does not silently upload reports or input data. Sharing a Wiki
example is separate from publishing your own machine metadata.
