package report

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/result"
)

func comparisonFixture() result.Comparison {
	m := func(v float64) result.Metric { return result.Metric{Value: v, Available: true} }
	return result.Comparison{
		Title: "addition", BaselinePolicy: "legacy-serial-scalar/v1",
		Columns: []result.Column{
			{StrategyID: "scalar", Label: "Scalar", Baseline: true},
			{StrategyID: "simd-partial", Label: "SIMDPartial"},
			{StrategyID: "simd-scalar-tail", Label: "SIMDTail"},
		},
		Rows: []result.Row{
			{Elements: 8, Cells: []result.Cell{
				{NanosecondsPerOp: m(25)}, {NanosecondsPerOp: m(8.5), ChangePercent: m(-66)}, {NanosecondsPerOp: m(30), ChangePercent: m(20)},
			}},
			{Elements: 16, Cells: []result.Cell{
				{NanosecondsPerOp: m(20)}, {NanosecondsPerOp: m(10), ChangePercent: m(-50)}, {},
			}},
			{Elements: 32, Cells: []result.Cell{
				{NanosecondsPerOp: m(0)}, {NanosecondsPerOp: m(0.125)}, {},
			}},
		},
	}
}

func TestFormatsFromSameStructuredFixture(t *testing.T) {
	table := comparisonFixture()
	const metadata = "cpu: measured CPU\ncommand: example, with metadata\n"
	const csvWant = "ELEMENTS,Scalar ns/op,SIMDPartial ns/op,SIMDPartial change,SIMDTail ns/op,SIMDTail change\n" +
		"8,25.00,8.50,-66.0,30.00,20.0\n16,20.00,10.00,-50.0,,\n32,0.00,0.12,,,\n"
	const mdWant = "Repeated measurements summarized by median.\n\n" +
		"## addition benchmark comparison (ns/op; change vs scalar)\n\n" +
		"| ELEMENTS | Scalar ns/op | SIMDPartial ns/op | SIMDPartial change | SIMDTail ns/op | SIMDTail change |\n" +
		"| ---: | ---: | ---: | ---: | ---: | ---: |\n" +
		"| 8 | 25.00 | 8.50 | -66.0% | 30.00 | +20.0% |\n" +
		"| 16 | 20.00 | 10.00 | -50.0% | — | — |\n" +
		"| 32 | 0.00 | 0.12 | — | — | — |\n" +
		"\n## Measurement environment\n\n    cpu: measured CPU\n    command: example, with metadata\n"
	const txtWant = "Repeated measurements summarized by median.\n\n" +
		"addition benchmark comparison (ns/op; change vs scalar)\n" +
		"ELEMENTS Scalar ns/op SIMDPartial ns/op SIMDPartial change SIMDTail ns/op SIMDTail change\n" +
		"8 25.00 8.50 -66.0% 30.00 +20.0%\n16 20.00 10.00 -50.0% — —\n32 0.00 0.12 — — —\n" +
		"\nMeasurement environment\ncpu: measured CPU\ncommand: example, with metadata\n"
	for format, want := range map[Format]string{TXT: txtWant, Markdown: mdWant, CSV: csvWant} {
		t.Run(string(format), func(t *testing.T) {
			var out bytes.Buffer
			if err := Write(&out, format, table, metadata); err != nil {
				t.Fatal(err)
			}
			got := out.String()
			if format == TXT {
				normalize := func(s string) string {
					lines := strings.Split(s, "\n")
					for i, line := range lines {
						lines[i] = strings.Join(strings.Fields(line), " ")
					}
					return strings.Join(lines, "\n")
				}
				got, want = normalize(got), normalize(want)
			}
			if got != want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, want)
			}
			if format == CSV {
				records, err := csv.NewReader(strings.NewReader(got)).ReadAll()
				if err != nil || len(records) != 4 || len(records[0]) != 6 {
					t.Fatalf("invalid CSV: %v %v", records, err)
				}
				if strings.Contains(got, "cpu:") || strings.Contains(got, "%") || strings.Contains(got, "—") {
					t.Fatal("CSV contains nonnumeric cells or metadata")
				}
			}
		})
	}
}

func TestFilteredComparisonWithoutScalar(t *testing.T) {
	table := result.Comparison{Title: "filtered", Columns: []result.Column{{StrategyID: "simd-partial", Label: "SIMDPartial"}},
		Rows: []result.Row{{Elements: 8, Cells: []result.Cell{{NanosecondsPerOp: result.Metric{Value: 2, Available: true}}}}}}
	for _, format := range []Format{TXT, Markdown, CSV} {
		var out bytes.Buffer
		if err := Write(&out, format, table, ""); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out.String(), "Scalar") || strings.Contains(out.String(), "Inf") {
			t.Fatalf("invalid filtered report: %s", out.String())
		}
		if format == CSV && out.String() != "ELEMENTS,SIMDPartial ns/op,SIMDPartial change\n8,2.00,\n" {
			t.Fatal(out.String())
		}
	}
}

func TestCSVQuotingAndMarkdownEscaping(t *testing.T) {
	table := comparisonFixture()
	table.Columns[1].Label = "A, \"quoted\" | <label>\nnext"
	var out bytes.Buffer
	if err := Write(&out, CSV, table, ""); err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(&out).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if records[0][2] != table.Columns[1].Label+" ns/op" {
		t.Fatal("CSV header did not round-trip")
	}
	out.Reset()
	if err := Write(&out, Markdown, table, "```\n<script>literal</script>\n"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `\| &lt;label&gt; next`) || !strings.Contains(out.String(), "    ```\n    <script>literal</script>\n") {
		t.Fatalf("Markdown escaping/metadata changed: %s", out.String())
	}
}

func TestParseFormatsAndWriteErrors(t *testing.T) {
	for _, name := range []string{"txt", "md", "csv"} {
		if got, err := ParseFormat(name); err != nil || string(got) != name {
			t.Fatalf("ParseFormat(%q) = %q, %v", name, got, err)
		}
	}
	for _, name := range []string{"", "TXT", "html", " csv"} {
		if _, err := ParseFormat(name); err == nil {
			t.Errorf("accepted unsupported format %q", name)
		}
	}
	want := errors.New("destination failed")
	for _, format := range []Format{TXT, Markdown, CSV} {
		if err := Write(errorWriter{want}, format, comparisonFixture(), ""); !errors.Is(err, want) {
			t.Errorf("%s: write error = %v", format, err)
		}
		if err := Write(shortWriter{}, format, comparisonFixture(), ""); !errors.Is(err, io.ErrShortWrite) {
			t.Errorf("%s: short write error = %v", format, err)
		}
	}
	table := comparisonFixture()
	before := comparisonFixture()
	var out bytes.Buffer
	if err := Write(&out, TXT, table, ""); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(table, before) {
		t.Fatal("rendering mutated results")
	}
	table.Rows[0].Cells = nil
	if err := Write(&out, TXT, table, ""); err == nil {
		t.Fatal("accepted malformed row")
	}
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
