// simdlab is the shared deterministic laboratory and future agent entry point.
// The _worker command is private: parent processes provide its bounded request.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labservice"
	"github.com/akolybelnikov/go-simd/internal/labworker"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

const maxPlanRequestBytes = 1 << 20

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "_worker" {
		if err := labworker.Serve(context.Background(), stdin, stdout, labapi.DefaultPlanPolicy()); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	if len(args) == 1 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help") {
		printUsage(stdout)
		return 0
	}
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	service, err := labservice.NewDeterministic(platform.Current(), labapi.DefaultPlanPolicy())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	switch args[0] {
	case "catalog":
		return runCatalog(service, args[1:], stdout, stderr)
	case "plan":
		return runPlan(service, args[1:], stdin, stdout, stderr)
	case "inspect":
		return runInspect(service, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runCatalog(service *labservice.Deterministic, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("catalog", flag.ContinueOnError)
	flags.SetOutput(stderr)
	group := flags.String("group", "all", "operation group or all")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		if err == nil {
			fmt.Fprintln(stderr, "catalog accepts no positional arguments")
		}
		return 2
	}
	response, err := service.Catalog(*group)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(stdout, stderr, response)
}

func runPlan(service *labservice.Deterministic, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: simdlab plan < request.json")
		return 2
	}
	var request labapi.PlanRequest
	if err := decodeBounded(stdin, maxPlanRequestBytes, &request); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	plan, err := service.Plan(request)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(stdout, stderr, plan)
}

func runInspect(service *labservice.Deterministic, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: simdlab inspect <measurements.json>")
		return 2
	}
	file, err := os.Open(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer file.Close()
	inspection, err := service.Inspect(file, evidence.DefaultReadLimits())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeJSON(stdout, stderr, inspection)
}

func decodeBounded(reader io.Reader, maximum int64, destination any) error {
	data, err := io.ReadAll(io.LimitReader(reader, maximum+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > maximum {
		return fmt.Errorf("JSON request exceeds %d bytes", maximum)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode JSON request: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return fmt.Errorf("JSON request must contain exactly one value")
	}
	return nil
}

func writeJSON(stdout, stderr io.Writer, value any) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: simdlab <command>")
	fmt.Fprintln(writer, "  catalog [-group group]     print structured capabilities")
	fmt.Fprintln(writer, "  plan                       read a JSON plan request from stdin")
	fmt.Fprintln(writer, "  inspect <measurements.json> validate and inspect saved evidence")
}
