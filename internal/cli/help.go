package cli

import (
	"flag"
	"fmt"
	"io"
)

func Help(w io.Writer, flags *flag.FlagSet, description string, examples ...string) {
	fmt.Fprintf(w, "Usage: %s [flags]\n\n%s\n\nOptions:\n", flags.Name(), description)
	flags.PrintDefaults()
	fmt.Fprintln(w, "\nExamples:")
	for _, example := range examples {
		fmt.Fprintln(w, "  "+example)
	}
}
