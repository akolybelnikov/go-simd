package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"simd"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "simdreport:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("simdreport", flag.ContinueOnError)
	group := flags.String("group", "all", "addition, computation, or all")
	count := flags.Int("count", 5, "measurement repetitions")
	benchtime := flags.String("benchtime", "100ms", "duration or iteration count per measurement")
	bench := flags.String("bench", ".", "Go benchmark selection expression")
	output := flags.String("output", "artifacts", "report directory")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *count < 1 {
		return fmt.Errorf("count must be positive")
	}
	groups := []string{"addition", "computation"}
	if *group != "all" {
		if !slices.Contains(groups, *group) {
			return fmt.Errorf("unknown group %q", *group)
		}
		groups = []string{*group}
	}
	for _, name := range groups {
		if err := report(name, *bench, *benchtime, *count, *output, stdout); err != nil {
			return err
		}
	}
	return nil
}

func report(group, bench, benchtime string, count int, output string, stdout io.Writer) error {
	directory := filepath.Join(output, group)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	raw, err := os.CreateTemp(directory, "benchmark-"+time.Now().UTC().Format("20060102-150405")+"-*.raw")
	if err != nil {
		return err
	}
	args := []string{"test", "./internal/" + group, "-run=^$", "-bench=" + bench, "-benchmem", "-benchtime=" + benchtime, "-count=" + strconv.Itoa(count)}
	var vector simd.Float32s
	metadata := fmt.Sprintf("date: %s\ngo: %s\nplatform: %s/%s\nGOMAXPROCS: %d\nSIMD float32 lanes: %d\ncommand: go %s\n", time.Now().UTC().Format(time.RFC3339), runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), vector.Len(), strings.Join(args, " "))
	envOutput, err := exec.Command("go", "env", "-json", "GOEXPERIMENT", "GOAMD64", "GOARM64", "GOTOOLCHAIN").CombinedOutput()
	if err != nil {
		raw.Close()
		return fmt.Errorf("go env: %w: %s", err, envOutput)
	}
	metadata += "toolchain environment: " + string(envOutput) + "\n"
	if _, err := io.WriteString(raw, metadata); err != nil {
		raw.Close()
		return err
	}
	fmt.Fprintf(stdout, "Measuring %s (%d repetitions)...\n", group, count)
	var measurements bytes.Buffer
	command := exec.Command("go", args...)
	command.Stdout = io.MultiWriter(raw, &measurements)
	command.Stderr = command.Stdout
	runErr := command.Run()
	closeErr := raw.Close()
	if runErr != nil {
		return fmt.Errorf("benchmark failed (raw output saved to %s): %w", raw.Name(), runErr)
	}
	if closeErr != nil {
		return closeErr
	}
	// Preserve the benchmark runner's CPU identification in the readable report too.
	for _, line := range strings.Split(measurements.String(), "\n") {
		if strings.HasPrefix(line, "cpu:") {
			metadata += strings.TrimSpace(line) + "\n"
			break
		}
	}
	table, err := comparison(measurements.String(), group)
	if err != nil {
		return fmt.Errorf("%s (raw output saved to %s)", err, raw.Name())
	}
	reportPath := strings.TrimSuffix(raw.Name(), ".raw") + ".txt"
	if err := os.WriteFile(reportPath, []byte(table+"\nMeasurement environment\n"+metadata), 0644); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%s\nSaved %s and %s\n", table, reportPath, raw.Name())
	return err
}

// comparison aggregates repeated measurements by their median. Only selected
// variants appear, and percentage changes require a measured scalar baseline.
func comparison(raw, title string) (string, error) {
	samples := map[int]map[string][]float64{}
	names := []string{}
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 || !strings.HasPrefix(fields[0], "Benchmark") {
			continue
		}
		parts := strings.Split(fields[0], "/")
		if len(parts) != 3 || !strings.HasPrefix(parts[1], "n=") {
			continue
		}
		size, err := strconv.Atoi(strings.TrimPrefix(parts[1], "n="))
		if err != nil {
			return "", err
		}
		name := parts[2]
		if i := strings.LastIndex(name, "-"); i >= 0 {
			name = name[:i]
		}
		for i := 2; i+1 < len(fields); i++ {
			if fields[i+1] != "ns/op" {
				continue
			}
			value, err := strconv.ParseFloat(fields[i], 64)
			if err != nil {
				return "", err
			}
			if samples[size] == nil {
				samples[size] = map[string][]float64{}
			}
			samples[size][name] = append(samples[size][name], value)
			if !slices.Contains(names, name) {
				names = append(names, name)
			}
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if len(samples) == 0 {
		return "", fmt.Errorf("no benchmark measurements found")
	}
	sizes := make([]int, 0, len(samples))
	results := benchutil.BenchmarkResults{}
	for size, variants := range samples {
		sizes = append(sizes, size)
		results[size] = map[string]float64{}
		for name, values := range variants {
			slices.Sort(values)
			n := len(values)
			median := values[n/2]
			if n%2 == 0 {
				median = (values[n/2-1] + median) / 2
			}
			results[size][name] = median
		}
	}
	slices.SortFunc(names, func(a, b string) int {
		order := []string{"Scalar", "SIMD", "SIMDPartial", "SIMDTail", "SIMDParallel"}
		ai, bi := slices.Index(order, a), slices.Index(order, b)
		if ai < 0 {
			ai = len(order)
		}
		if bi < 0 {
			bi = len(order)
		}
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
		return strings.Compare(a, b)
	})
	slices.Sort(sizes)
	var out bytes.Buffer
	fmt.Fprintln(&out, "Repeated measurements summarized by median.")
	if err := benchutil.PrintComparison(&out, title, sizes, results, names); err != nil {
		return "", err
	}
	return out.String(), nil
}
