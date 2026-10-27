// Package report renders structured comparisons without running measurements
// or calculating medians and baseline ratios.
package report

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/akolybelnikov/go-simd/internal/result"
)

type Format string

const (
	TXT      Format = "txt"
	Markdown Format = "md"
	CSV      Format = "csv"
)

func ParseFormat(value string) (Format, error) {
	switch Format(value) {
	case TXT, Markdown, CSV:
		return Format(value), nil
	default:
		return "", fmt.Errorf("unsupported format %q (choose txt, md, or csv)", value)
	}
}

// Write uses the same columns/cells for every format. CSV contains only tabular
// data; readable metadata belongs in TXT/Markdown and the companion raw file.
// Empty metadata renders just the table, as used for terminal output.
func Write(w io.Writer, format Format, table result.Comparison, metadata string) error {
	if _, err := ParseFormat(string(format)); err != nil {
		return err
	}
	for _, row := range table.Rows {
		if len(row.Cells) != len(table.Columns) {
			return fmt.Errorf("report row does not match columns")
		}
	}
	var out bytes.Buffer
	header, rows := records(table, format == CSV)
	switch format {
	case TXT:
		fmt.Fprintln(&out, "Repeated measurements summarized by median.")
		fmt.Fprintf(&out, "\n%s benchmark comparison (ns/op; change vs scalar)\n", table.Title)
		tabs := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
		for _, record := range append([][]string{header}, rows...) {
			fmt.Fprintln(tabs, strings.Join(record, "\t")+"\t")
		}
		if err := tabs.Flush(); err != nil {
			return err
		}
		if metadata != "" {
			fmt.Fprint(&out, "\nMeasurement environment\n"+metadata)
		}
	case Markdown:
		fmt.Fprintln(&out, "Repeated measurements summarized by median.")
		fmt.Fprintf(&out, "\n## %s benchmark comparison (ns/op; change vs scalar)\n\n", markdownCell(table.Title))
		markdownRow(&out, header)
		alignment := make([]string, len(header))
		for i := range alignment {
			alignment[i] = "---:"
		}
		markdownRow(&out, alignment)
		for _, record := range rows {
			markdownRow(&out, record)
		}
		if metadata != "" {
			fmt.Fprint(&out, "\n## Measurement environment\n\n")
			// Indentation keeps arbitrary metadata literal without a fence that
			// input text could accidentally close.
			for _, line := range strings.Split(strings.TrimRight(metadata, "\r\n"), "\n") {
				fmt.Fprintln(&out, "    "+strings.TrimSuffix(line, "\r"))
			}
		}
	case CSV:
		writer := csv.NewWriter(&out)
		if err := writer.WriteAll(append([][]string{header}, rows...)); err != nil {
			return err
		}
	}
	n, err := w.Write(out.Bytes())
	if err == nil && n != out.Len() {
		return io.ErrShortWrite
	}
	return err
}

func records(table result.Comparison, numeric bool) ([]string, [][]string) {
	header := []string{"ELEMENTS"}
	for _, c := range table.Columns {
		header = append(header, c.Label+" ns/op")
		if !c.Baseline {
			header = append(header, c.Label+" change")
		}
	}
	var rows [][]string
	for _, row := range table.Rows {
		record := []string{strconv.Itoa(row.Elements)}
		for i, c := range table.Columns {
			cell := row.Cells[i]
			record = append(record, timing(cell.NanosecondsPerOp, numeric))
			if !c.Baseline {
				change := "—"
				if numeric {
					change = ""
				}
				if cell.ChangePercent.Available {
					if numeric {
						change = fmt.Sprintf("%.1f", cell.ChangePercent.Value)
					} else {
						change = fmt.Sprintf("%+.1f%%", cell.ChangePercent.Value)
					}
				}
				record = append(record, change)
			}
		}
		rows = append(rows, record)
	}
	return header, rows
}

func timing(metric result.Metric, numeric bool) string {
	if metric.Available {
		return fmt.Sprintf("%.2f", metric.Value)
	}
	if numeric {
		return ""
	}
	return "—"
}

func markdownRow(w io.Writer, fields []string) {
	escaped := make([]string, len(fields))
	for i, value := range fields {
		escaped[i] = markdownCell(value)
	}
	fmt.Fprintln(w, "| "+strings.Join(escaped, " | ")+" |")
}

func markdownCell(value string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;",
		"\\", "\\\\", "|", "\\|", "\r", " ", "\n", " ").Replace(value)
}
