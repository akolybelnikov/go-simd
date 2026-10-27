package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

// Listing consumes metadata only: no fixture allocation, kernel binding,
// calibration, benchmark invocation, Go subprocess, or artifact creation.
func PrintCatalog(w io.Writer, group string, machine platform.Info) error {
	var out strings.Builder
	fmt.Fprintln(&out, "Implemented laboratory catalogue (no benchmarks run)")
	fmt.Fprintf(&out, "go: %s\nplatform: %s/%s\nSIMD bits: %d\nSIMD float32 lanes: %d\nemulated: %t\nGOMAXPROCS: %d\n",
		machine.GoVersion, machine.OS, machine.Arch, machine.Bits, machine.Lanes, machine.Emulated, machine.Procs)
	for _, op := range catalog.All() {
		if group != "all" && group != op.Group {
			continue
		}
		fmt.Fprintf(&out, "\n%s [%s/%s]\ngroup: %s\nbenchmark: %s\ncategory: %s\nsemantics: %s\nexpression: %s\nscalar reference: %s\nexecution: %s\n",
			op.Title, op.ID, op.TypeID, op.Group, op.BenchmarkName, op.Category, op.ArithmeticID, op.Expression, op.ReferenceVariantID, op.ExecutionMode)
		var inputs []string
		for _, value := range op.Inputs {
			inputs = append(inputs, describeValue(value))
		}
		fmt.Fprintf(&out, "inputs: %s\noutput: %s\nscalar parameters: %d\n", strings.Join(inputs, ", "), describeValue(op.Output), len(op.Parameters))
		bytes, err := op.LogicalBytes(1)
		if err != nil {
			return err
		}
		fmt.Fprintf(&out, "logical traffic at n=1: %d bytes/call (effective, not DRAM)\n", bytes)
		table := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(table, "LABEL\tSTRATEGY ID\tMODE\tWORKER POLICY")
		for _, variant := range op.Variants {
			mode := "scalar"
			if variant.UsesSIMD {
				mode = "portable hardware"
				if machine.Emulated {
					mode = "portable emulation"
				}
			}
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", variant.Label, variant.ID, mode, variant.Workers)
		}
		if err := table.Flush(); err != nil {
			return err
		}
	}
	fmt.Fprintln(&out, "\nOnly registered operation/type pairs are listed; no implicit casts or ISA-only extensions.")
	n, err := io.WriteString(w, out.String())
	if err == nil && n != out.Len() {
		return io.ErrShortWrite
	}
	return err
}

func describeValue(value catalog.Value) string {
	if value.Shape == catalog.Array {
		return value.Name + ":" + value.TypeID + "[]"
	}
	return value.Name + ":" + value.TypeID
}
