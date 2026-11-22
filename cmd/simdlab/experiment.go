package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/agent/labdispatch"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/labtools"
	"github.com/akolybelnikov/go-simd/internal/labworker"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

const experimentInstructions = `Propose one go-simd benchmark using plan_experiment. Use the explicitly requested catalog group, strategy IDs, element sizes, repetitions and benchtime. If details are missing or the request names unsupported capabilities, ask for clarification. The CLI displays the resolved plan for human confirmation and controls execution. Never claim measurements have been taken.`

const experimentRunTimeout = 2 * time.Minute

type experimentExecutor struct {
	runner   *labworker.Runner
	root     string
	progress func(labworker.Manifest)
}

func (e experimentExecutor) Execute(ctx context.Context, request labworker.Request) (labworker.RunResult, error) {
	return e.runner.RunWithProgress(ctx, e.root, request, e.progress)
}

func runExperiment(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("experiment", flag.ContinueOnError)
	flags.SetOutput(stderr)
	preview := flags.Bool("preview", false, "enable the local experiment preview")
	endpoint := flags.String("endpoint", "http://127.0.0.1:11434", "Ollama server base URL")
	model := flags.String("model", "", "installed Ollama model (explicit selection; experimental)")
	think := flags.Bool("think", true, "request Ollama's separate thinking channel")
	root := flags.String("artifacts", filepath.Join("artifacts", "experiments"), "local artifact directory")
	group := flags.String("group", "", "exact catalog group")
	variants := flags.String("variants", "", "comma-separated strategy IDs")
	sizes := flags.String("sizes", "", "comma-separated element counts")
	count := flags.Int("count", 0, "repetitions per case")
	benchtime := flags.String("benchtime", "", "Go benchmark duration or fixed iterations")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	question := strings.TrimSpace(strings.Join(flags.Args(), " "))
	var explicit bool
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "group", "variants", "sizes", "count", "benchtime":
			explicit = true
		}
	})
	if !*preview || len(question) > maxQuestionBytes || *root == "" || explicit == (question != "") {
		fmt.Fprintln(stderr, "experiment requires -preview and either a request in words or all five exact plan flags")
		return 2
	}
	var client agent.ModelClient
	var request *labapi.PlanRequest
	if explicit {
		if *group == "" || *variants == "" || *sizes == "" || *count <= 0 || *benchtime == "" || len(*group)+len(*sizes)+len(*variants)+len(*benchtime) > maxQuestionBytes {
			fmt.Fprintln(stderr, "supply -group, -variants, -sizes, -count, and -benchtime for an exact plan")
			return 2
		}
		request = &labapi.PlanRequest{Group: *group, Variants: strings.Split(*variants, ","), Count: *count, Benchtime: *benchtime}
		for _, part := range strings.Split(*sizes, ",") {
			size, err := strconv.Atoi(part)
			if err != nil {
				fmt.Fprintln(stderr, "sizes must be comma-separated integers")
				return 2
			}
			request.Sizes = append(request.Sizes, size)
		}
	} else {
		modelClient, err := newLocalModelClient(*endpoint, *model, experimentRunTimeout, *think)
		if err != nil {
			fmt.Fprintln(stderr, terminalText(err.Error()))
			return 2
		}
		client = modelClient
		fmt.Fprintln(stderr, "Experimental Ollama model:", terminalText(*model))
	}
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	runner, err := labworker.NewRunner(executable, "_worker")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	interrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	executor := experimentExecutor{runner: runner, root: *root, progress: func(m labworker.Manifest) {
		fmt.Fprintf(stderr, "%s: %s (%d observations saved)\n", m.JobID, m.Status, m.Observations)
	}}
	return experiment(interrupt, client, executor, question, request, stdin, stdout, stderr)
}

func experiment(ctx context.Context, client agent.ModelClient, executor labtools.Executor, question string, request *labapi.PlanRequest, stdin io.Reader, stdout, stderr io.Writer) int {
	policy := labapi.DefaultPlanPolicy()
	tools, err := labtools.New(labrun.Runtime(platform.Current()), policy, executor)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	var plan labapi.ExperimentPlan
	if request != nil {
		plan, err = tools.ProposeExperiment(*request)
		if err != nil {
			fmt.Fprintln(stderr, terminalText(err.Error()))
			return 2
		}
	} else {
		var code int
		plan, code = proposeModelExperiment(ctx, client, tools.ModelAuthority(), question, stderr)
		if code != 0 {
			return code
		}
	}
	if writeJSON(stderr, stderr, plan) != 0 {
		return 1
	}
	if _, err := fmt.Fprintf(stderr, "Limits: %d cases, %d repetitions per case, %d observations, %s per benchmark or %d fixed iterations, %s worker deadline.\nType 'run %s' to confirm this exact plan: ", policy.MaxCases, policy.MaxCount, policy.MaxMeasurements, policy.MaxBenchtime, policy.MaxFixedIterations, experimentRunTimeout, plan.ID); err != nil {
		return 1
	}
	review, cancelReview := context.WithTimeout(ctx, labtools.ExperimentPlanTTL)
	err = readExperimentConfirmation(review, stdin, "run "+plan.ID)
	cancelReview()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if err := ctx.Err(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := tools.ConfirmExperiment(plan.ID); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	measuring, cancelMeasurement := context.WithTimeout(ctx, experimentRunTimeout)
	defer cancelMeasurement()
	result, runErr := tools.ExecuteConfirmed(measuring, plan.ID)
	if result.Directory != "" {
		output := struct {
			labworker.Manifest
			Evidence string `json:"evidence,omitempty"`
		}{Manifest: result.Manifest}
		if result.Capture != nil {
			output.Evidence = filepath.Join(result.Directory, "measurements.json")
		}
		if writeJSON(stdout, stderr, output) != 0 {
			return 1
		}
	}
	if runErr != nil {
		fmt.Fprintln(stderr, terminalText(runErr.Error()))
		return 1
	}
	return 0
}

func proposeModelExperiment(ctx context.Context, client agent.ModelClient, authority labtools.ModelAuthority, question string, stderr io.Writer) (labapi.ExperimentPlan, int) {
	runtime, err := agent.NewRuntime(client, experimentInstructions, []agent.ToolDefinition{labdispatch.ExperimentPlanTool()})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return labapi.ExperimentPlan{}, 1
	}
	planning, cancel := context.WithTimeout(ctx, experimentRunTimeout)
	defer cancel()
	fmt.Fprintln(stderr, "Waiting for the experiment proposal…")
	reply, err := runtime.RunModelStep(planning, []agent.Message{{Role: agent.RoleUser, Content: question}})
	if err != nil {
		fmt.Fprintln(stderr, terminalText(err.Error()))
		return labapi.ExperimentPlan{}, 1
	}
	if len(reply.ToolCalls) == 0 {
		fmt.Fprintln(stderr, "The model did not submit a plan. Use the exact plan flags (experiment -h) when the inputs are known.")
		fmt.Fprintln(stderr, "Model response:", terminalText(reply.Content))
		return labapi.ExperimentPlan{}, 2
	}
	if len(reply.ToolCalls) != 1 {
		fmt.Fprintln(stderr, "expected exactly one experiment proposal")
		return labapi.ExperimentPlan{}, 1
	}
	plan, err := labdispatch.ProposeExperiment(authority, reply.ToolCalls[0])
	if err != nil {
		fmt.Fprintln(stderr, terminalText(err.Error()))
		return labapi.ExperimentPlan{}, 1
	}
	return plan, 0
}

func readExperimentConfirmation(ctx context.Context, stdin io.Reader, expected string) error {
	// Terminal reads can block until Enter. The CLI exits on cancellation even
	// if the reader has not returned yet.
	confirmed := make(chan bool, 1)
	go func() {
		line, err := bufio.NewReader(io.LimitReader(stdin, 256)).ReadString('\n')
		confirmed <- err == nil && strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r") == expected
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ok := <-confirmed:
		if !ok {
			return fmt.Errorf("experiment not confirmed")
		}
		return ctx.Err()
	}
}
