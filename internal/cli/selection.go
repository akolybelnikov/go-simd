// Package cli shares the human command surface between executable entry points.
package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

type Selection struct {
	group, legacy         string
	addition, computation bool
	List                  bool
}

func (s *Selection) Register(flags *flag.FlagSet) {
	flags.StringVar(&s.group, "group", "all", "benchmark group: addition, computation, or all")
	flags.StringVar(&s.legacy, "function", "", "deprecated alias: add, compute, or all")
	flags.BoolVar(&s.addition, "addition", false, "alias for -group addition")
	flags.BoolVar(&s.computation, "computation", false, "alias for -group computation")
	flags.BoolVar(&s.List, "list", false, "list implemented kernels and runtime capabilities without benchmarking")
}

func (s Selection) Resolve(flags *flag.FlagSet) (string, error) {
	if flags.NArg() != 0 {
		return "", fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	var selections []string
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "group":
			selections = append(selections, s.group)
		case "function":
			selections = append(selections, s.legacy)
		case "addition":
			if s.addition {
				selections = append(selections, "addition")
			}
		case "computation":
			if s.computation {
				selections = append(selections, "computation")
			}
		}
	})
	group := "all"
	for i, selected := range selections {
		switch selected {
		case "add":
			selected = "addition"
		case "compute":
			selected = "computation"
		}
		if _, ok := catalog.ByGroup(selected); !ok && selected != "all" {
			return "", fmt.Errorf("unknown group %q (choose addition, computation, or all)", selected)
		}
		if i > 0 && selected != group {
			return "", fmt.Errorf("conflicting group selections cannot be used together")
		}
		group = selected
	}
	return group, nil
}
