// Package cli shares the human command surface between executable entry points.
package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
)

type Selection struct {
	Input                 input.Options
	group, legacy         string
	addition, computation bool
	List                  bool
}

func (s *Selection) Register(flags *flag.FlagSet) {
	s.Input = input.DefaultOptions()
	flags.Var(float32Flag{&s.Input.Alpha}, "alpha", "float32 scale/affine multiplier (broadcast per call)")
	flags.Var(float32Flag{&s.Input.Beta}, "beta", "float32 affine offset (unfused)")
	flags.Var(float32Flag{&s.Input.Threshold}, "threshold", "float32 threshold for threshold/selection")
	flags.Var(float32Flag{&s.Input.Lower}, "lower", "float32 exclusive lower range bound")
	flags.Var(float32Flag{&s.Input.Upper}, "upper", "float32 exclusive upper range bound")
	flags.StringVar(&s.Input.Pattern, "pattern", s.Input.Pattern, "predicate fixture: none, all, alternating, clustered, shuffled")
	flags.Uint64Var(&s.Input.Seed, "seed", s.Input.Seed, "deterministic shuffled predicate seed")
	flags.StringVar(&s.group, "group", "all", "operation group (use -list to see available groups)")
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
			return "", fmt.Errorf("unknown group %q (use -list to see available groups)", selected)
		}
		if i > 0 && selected != group {
			return "", fmt.Errorf("conflicting group selections cannot be used together")
		}
		group = selected
	}
	if err := s.Input.Validate(); err != nil {
		return "", err
	}
	return group, nil
}
