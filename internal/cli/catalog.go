package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

// Listing consumes metadata only: no fixture allocation, kernel binding,
// calibration, benchmark invocation, Go subprocess, or artifact creation.
func PrintCatalog(w io.Writer, group string, machine platform.Info) error {
	return PrintCatalogWithOptions(w, group, machine, input.DefaultOptions())
}

func PrintCatalogWithOptions(w io.Writer, group string, machine platform.Info, options input.Options) error {
	var out strings.Builder
	fmt.Fprintln(&out, "Implemented laboratory catalogue (no benchmarks run)")
	fmt.Fprintf(&out, "go: %s\nplatform: %s/%s\nSIMD bits: %d\nSIMD float32 lanes: %d\nemulated: %t\nGOMAXPROCS: %d\n",
		machine.GoVersion, machine.OS, machine.Arch, machine.Bits, machine.Lanes, machine.Emulated, machine.Procs)
	if machine.FMA32Path != "" {
		fmt.Fprintf(&out, "FMA32 SIMD path: %s\n", machine.FMA32Path)
	}
	if group == "all" {
		fmt.Fprintln(&out)
		table := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(table, "GROUP\tOPERATION / TYPE\tSTRATEGIES\tLOGICAL BYTES AT n=1")
		for _, op := range catalog.All() {
			bytes, err := op.LogicalBytes(1)
			if err != nil {
				return err
			}
			fmt.Fprintf(table, "%s\t[%s/%s]\t%d\t%d\n", op.Group, op.ID, op.TypeID, len(op.Variants), bytes)
		}
		if err := table.Flush(); err != nil {
			return err
		}
		fmt.Fprintln(&out, "\nUse -list -group <group> for expressions, parameters, strategy labels and semantics.")
		fmt.Fprintln(&out, "Logical traffic is effective work accounting, not measured DRAM traffic.")
		return writeCatalog(w, out.String())
	}
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
		if op.ScalarResult() {
			fmt.Fprintf(&out, "fixture: %s\nlogical traffic model: %d*n + 4 bytes/call\nnumerics: %s\n", options.Generator(op), 4*len(op.Inputs), op.NumericalPolicy)
		}
		if op.Parameterized() {
			for _, p := range op.Parameters {
				fmt.Fprintf(&out, "  %s: %g (float32 bits 0x%08x)\n", p.Name, options.Float(p.Name), options.Bits(p.Name))
			}
			if op.ID == catalog.AffineFMA {
				fmt.Fprintf(&out, "fixture: %s\nsetup: native vector path broadcasts once per call; software path uses scalar parameters\n", options.Generator(op))
			} else {
				fmt.Fprintf(&out, "fixture: %s\nsetup: scalar parameters prepared; SIMD broadcasts once per kernel call\n", options.Generator(op))
			}
		}
		if op.NumericalPolicy != "" && !op.ScalarResult() {
			fmt.Fprintln(&out, "numerics: "+op.NumericalPolicy)
		}
		if op.ID == catalog.Clamp {
			fmt.Fprintln(&out, "clamp pattern: ordered clipped positions; remaining values include endpoints, zeros and unchanged NaN payloads")
			fmt.Fprintln(&out, "SIMDMinMax includes outside-bound/tie/NaN repairs in timing; logical traffic 8*n bytes/call")
		}
		if op.ID == catalog.Copy {
			fmt.Fprintf(&out, "fixture: %s\nlogical traffic model: 8*n bytes/call; disjoint float32 input/output\n", options.Generator(op))
			fmt.Fprintln(&out, "controls: Scalar is a source loop; GoCopy uses compiler/runtime Go copy; source labels do not prove machine instruction selection")
		}
		bytes, err := op.LogicalBytes(1)
		if err != nil {
			return err
		}
		fmt.Fprintf(&out, "logical traffic at n=1: %d bytes/call (effective, not DRAM)\n", bytes)
		if op.ID == catalog.PredicateGTPacked {
			fmt.Fprintln(&out, "logical traffic model: 4*n+ceil(n/8) bytes/call; packed LSB-first output")
			fmt.Fprintln(&out, "SIMD boundary: partial load, compare, select 0/1, convert/extract, pack valid bits, byte writes; all timed")
		} else if op.ByteResult() {
			fmt.Fprintln(&out, "logical traffic model: 5*n bytes/call; uint8 0/1 output")
			fmt.Fprintln(&out, "SIMD boundary: partial load, compare, select 0/1, convert/extract, valid byte writes; all timed")
		}
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
			if op.ID == catalog.DotFMA {
				mode = "software FMA32"
				if variant.UsesSIMD {
					switch machine.FMA32Path {
					case "vector-fma32/v1":
						mode = "vector FMA32 / software tail"
					case "software-fma32/v1":
						mode = "software striped FMA32"
					default:
						mode = "unknown FMA32 gate"
					}
				}
			}
			if op.ID == catalog.Copy && !variant.UsesSIMD {
				mode = "source loop (compiler-selected)"
				if variant.ID == catalog.GoCopy {
					mode = "Go builtin (compiler/runtime)"
				}
			}
			if op.ID == catalog.AffineFMA {
				mode = "software FMA32"
				if variant.UsesSIMD {
					switch machine.FMA32Path {
					case "vector-fma32/v1":
						mode = "verified vector FMA32"
					case "software-fma32/v1":
					default:
						mode = "unknown FMA32 path"
					}
				}
			}
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", variant.Label, variant.ID, mode, variant.Workers)
		}
		if err := table.Flush(); err != nil {
			return err
		}
	}
	fmt.Fprintln(&out, "\nOnly registered operation/type pairs are listed; no implicit casts or ISA-only extensions.")
	return writeCatalog(w, out.String())
}

func writeCatalog(w io.Writer, value string) error {
	n, err := io.WriteString(w, value)
	if err == nil && n != len(value) {
		return io.ErrShortWrite
	}
	return err
}

func describeValue(value catalog.Value) string {
	if value.Shape == catalog.PackedBits {
		return value.Name + ":" + value.TypeID + "[ceil(n/8)]"
	}
	if value.Shape == catalog.Array {
		return value.Name + ":" + value.TypeID + "[]"
	}
	return value.Name + ":" + value.TypeID
}
