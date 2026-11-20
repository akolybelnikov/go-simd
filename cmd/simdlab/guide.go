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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/analysis"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labservice"
	"github.com/akolybelnikov/go-simd/internal/labtools"
)

const guideExample = "Compare ordinary and vectorized addition for 1,024 numbers."
const guideScope = "The lab times individual operations on generated arrays on the measuring machine. It cannot benchmark your files, a different machine, or an end-to-end pipeline."
const guideFallback = "Try: " + guideExample + " To run the starter directly, use simdlab run addition."

func runGuide(service *labservice.Deterministic, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("ask", flag.ContinueOnError)
	flags.SetOutput(stderr)
	endpoint := flags.String("endpoint", "http://127.0.0.1:11434", "Ollama server for additional phrasings")
	model := flags.String("model", "qwen3:4b", "installed Ollama model")
	think := flags.Bool("think", false, "request Ollama's separate thinking channel")
	timeout := flags.Duration("timeout", 30*time.Second, "model deadline (maximum 5m)")
	root := flags.String("artifacts", filepath.Join("artifacts", "experiments"), "local artifact directory")
	path := flags.String("evidence", "", "open saved measurements for follow-up questions")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *timeout <= 0 || *timeout > 5*time.Minute || *root == "" {
		fmt.Fprintln(stderr, "usage: simdlab ask [options]; enter questions after the prompt")
		return 2
	}
	client, err := newLocalModelClient(*endpoint, *model, *timeout, *think)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	available, err := service.Catalog("all")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	session := guideSession{
		operations: available.Operations, model: client, timeout: *timeout,
		newTools: func() (*labtools.Toolset, error) { return newComparisonTools(*root) },
	}
	if *path != "" {
		absolute, err := filepath.Abs(*path)
		if err != nil {
			fmt.Fprintln(stderr, terminalText(err.Error()))
			return 1
		}
		file, err := os.Open(absolute)
		if err != nil {
			fmt.Fprintln(stderr, terminalText(err.Error()))
			return 1
		}
		inspection, err := service.Inspect(file, evidence.DefaultReadLimits())
		file.Close()
		if err != nil {
			fmt.Fprintln(stderr, terminalText(err.Error()))
			return 1
		}
		session.saved = savedComparison{inspection: inspection, path: absolute}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := session.run(ctx, stdin, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, terminalText(err.Error()))
		return 1
	}
	return 0
}

type guideSession struct {
	operations []labapi.OperationDescriptor
	model      agent.ModelClient
	timeout    time.Duration
	newTools   func() (*labtools.Toolset, error)
	saved      savedComparison
	generated  bool // Set for runs executed in this session.
}

func (s *guideSession) run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	if _, err := fmt.Fprintln(stdout, "Ask about an operation, compare two methods, or explain saved measurements. Type quit to leave.\nTry: "+guideExample); err != nil {
		return err
	}
	if s.saved.path != "" {
		if _, err := fmt.Fprintln(stdout, terminalText("Opened measurements: "+s.saved.path)); err != nil {
			return err
		}
	}
	reader := bufio.NewScanner(stdin)
	reader.Buffer(make([]byte, 1024), maxQuestionBytes+2)
	var previous string
	var pending guideRequest
	for {
		question, err := promptLine(ctx, reader, stdout, "> ")
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		switch strings.ToLower(question) {
		case "":
			continue
		case "quit", "exit":
			return nil
		case "yes", "y", "no", "n":
			if _, err := fmt.Fprintln(stdout, "There is no comparison waiting for approval."); err != nil {
				return err
			}
			continue
		}

		request, matched := matchGuideRequest(question)
		if !matched {
			answer := strings.ToLower(strings.TrimRight(question, ".?!"))
			switch {
			case pending.Action == "clarify" && s.operation(answer) != nil:
				request, matched = guideRequest{Action: "compare", Group: answer}, true
			case pending.Action == "compare":
				if size, ok := guideSize(answer); ok {
					request, matched = pending, true
					request.Size = size
				}
			}
		}
		pending = guideRequest{}
		if !matched {
			if _, err := fmt.Fprintln(stdout, "Asking Ollama to resolve the request…"); err != nil {
				return err
			}
			planning, cancel := context.WithTimeout(ctx, s.timeout)
			request, err = routeGuideRequest(planning, s.model, s.operations, question, previous, s.saved.path != "")
			cancel()
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				message := "The assistant could not resolve that request with Ollama."
				if errors.Is(err, context.DeadlineExceeded) {
					message = "Ollama did not respond in time."
				}
				if _, err := fmt.Fprintln(stdout, message+" "+guideFallback); err != nil {
					return err
				}
				if s.saved.path != "" {
					if _, err := fmt.Fprintln(stdout, savedResultOffer); err != nil {
						return err
					}
				}
				previous = ""
				continue
			}
			switch request.Action {
			case "result", "repeats", "result_repeats", "details":
				// A model can recognize a result question, but it cannot establish
				// that the saved run covers all the user's qualifiers.
				request = guideRequest{Action: "result_target"}
			}
		}

		clarify, err := s.handle(ctx, reader, request, stdout, stderr)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		previous = ""
		if clarify {
			previous = question
			pending = request
		}
	}
}

func (s *guideSession) handle(ctx context.Context, reader *bufio.Scanner, request guideRequest, stdout, stderr io.Writer) (bool, error) {
	switch request.Action {
	case "unsupported":
		_, err := fmt.Fprintln(stdout, guideScope+" "+guideFallback)
		return false, err
	case "invalid_size":
		_, err := fmt.Fprintf(stdout, "Choose a whole number from 1 to %d values per input array.\n", labapi.MaxExperimentSize)
		return false, err
	case "inference":
		_, err := fmt.Fprintln(stdout, "The lab reports observed timings and repeat ranges. It does not estimate statistical significance or confidence intervals.")
		return false, err
	case "clarify":
		_, err := fmt.Fprintln(stdout, "Which operation would you like to compare? You can name an operation such as addition or summing an array.")
		return true, err
	case "result", "repeats", "result_repeats", "details", "result_target":
		return false, s.explain(request, stdout)
	}
	operation := s.operation(request.Group)
	if operation == nil {
		_, err := fmt.Fprintln(stdout, "Which catalog operation should I use? For example, addition or summing an array.")
		return true, err
	}
	if request.Action == "find" {
		return false, discoverOperation(stdout, *operation)
	}
	size := request.Size
	if size == 0 {
		size = starterSize
	}
	if size < 1 || size > operation.MaxSize {
		_, err := fmt.Fprintf(stdout, "Which size should I use? Choose a whole number from 1 to %d values per input array.\n", operation.MaxSize)
		return true, err
	}
	methods, err := guideMethods(*operation, request.Methods)
	if err != nil {
		if _, err := fmt.Fprintln(stdout, "Which two supported methods should I compare? Include the ordinary scalar method so I can give a measured difference."); err != nil {
			return false, err
		}
		return true, discoverOperation(stdout, *operation)
	}
	tools, err := s.newTools()
	if err != nil {
		return false, err
	}
	plan, err := tools.ProposeExperiment(comparisonRequest(operation.Group, methods, size))
	if err != nil {
		_, writeErr := fmt.Fprintln(stdout, terminalText(err.Error()))
		return true, writeErr
	}
	if _, err := io.WriteString(stdout, terminalText(comparisonPlanText(plan))); err != nil {
		return false, err
	}
	review, cancel := context.WithTimeout(ctx, labtools.ExperimentPlanTTL)
	defer cancel()
	for {
		reply, err := promptLine(review, reader, stdout, "Run this comparison? [yes/no] ")
		if errors.Is(err, io.EOF) || errors.Is(err, context.DeadlineExceeded) {
			_, writeErr := fmt.Fprintln(stdout, "Comparison canceled.")
			if writeErr != nil {
				return false, writeErr
			}
			// A timeout can leave the terminal read blocked. End the session.
			return false, io.EOF
		}
		if err != nil {
			return false, err
		}
		switch strings.ToLower(reply) {
		case "yes", "y":
			saved, err := measureComparison(ctx, tools, plan, stdout, stderr)
			if saved.path != "" {
				s.saved, s.generated = saved, true
			}
			if err != nil {
				if ctx.Err() != nil {
					return false, fmt.Errorf("comparison canceled: %w", ctx.Err())
				}
				_, err = fmt.Fprintln(stderr, terminalText(err.Error()))
			}
			return false, err
		case "no", "n", "":
			_, err := fmt.Fprintln(stdout, "Comparison canceled.")
			return false, err
		case "quit", "exit":
			if _, err := fmt.Fprintln(stdout, "Comparison canceled."); err != nil {
				return false, err
			}
			return false, io.EOF
		}
	}
}

const savedResultOffer = "To see the saved run separately, ask: Explain the saved result."

func (s *guideSession) explain(request guideRequest, stdout io.Writer) error {
	if s.saved.path == "" {
		_, err := fmt.Fprintln(stdout, "No saved measurements yet. "+guideFallback)
		return err
	}
	if request.Action == "result_target" {
		_, err := fmt.Fprintln(stdout, "I couldn't match that question to the saved run. "+savedResultOffer)
		return err
	}
	inspection := s.saved.inspection
	if request.Group != "" {
		operation, _ := catalog.ByGroup(request.Group)
		if !slices.ContainsFunc(inspection.Series, func(series analysis.SeriesFact) bool {
			return series.OperationID == operation.ID
		}) {
			_, err := fmt.Fprintf(stdout, "The saved run has no measurements for %s. %s\n", operation.Title, savedResultOffer)
			return err
		}
	}
	if request.Size != 0 {
		inspection = resultAtSize(inspection, request.Size)
		if len(inspection.Series) == 0 {
			_, err := fmt.Fprintf(stdout, "The saved run has no measurements at %d values per array. %s\n", request.Size, savedResultOffer)
			return err
		}
	}
	var answer string
	switch request.Action {
	case "repeats":
		answer = repeatSummary(inspection) + shortResultCaveats(inspection)
	case "details":
		answer = inspectionText(inspection)
	default:
		answer = comparisonSummary(inspection)
		if request.Action == "result_repeats" {
			answer += repeatSummary(inspection)
		}
		if s.generated {
			answer += "This run used generated data on this machine.\n"
		} else {
			answer += "These timings describe the measuring host recorded in the file.\n"
		}
	}
	_, err := io.WriteString(stdout, terminalText(answer+"Measurements: "+s.saved.path+"\n"))
	return err
}

func resultAtSize(inspection analysis.Interpretation, size int) analysis.Interpretation {
	inspection.Series = slices.DeleteFunc(slices.Clone(inspection.Series), func(series analysis.SeriesFact) bool {
		return series.Elements != size
	})
	inspection.Comparisons = slices.DeleteFunc(slices.Clone(inspection.Comparisons), func(fact analysis.ComparisonFact) bool {
		return fact.Elements != size
	})
	return inspection
}

func (s *guideSession) operation(group string) *labapi.OperationDescriptor {
	return findGuideOperation(s.operations, group)
}

func findGuideOperation(operations []labapi.OperationDescriptor, group string) *labapi.OperationDescriptor {
	for i := range operations {
		if operations[i].Group == group {
			return &operations[i]
		}
	}
	return nil
}

func discoverOperation(stdout io.Writer, operation labapi.OperationDescriptor) error {
	var text strings.Builder
	fmt.Fprintf(&text, "%s · %s inputs\n", operation.Title, operation.TypeID)
	for _, method := range operation.Strategies {
		fmt.Fprintln(&text, "  "+methodName(operation.Group, method.ID, method.Label))
	}
	fmt.Fprintf(&text, "Values per input array: any whole number from 1 to %d.\n", operation.MaxSize)
	fmt.Fprintf(&text, "Suggested sweep sizes: %s\n", supportedSizesText(operation.SupportedSizes))
	_, err := io.WriteString(stdout, text.String())
	return err
}

func supportedSizesText(sizes []int) string {
	values := make([]string, 0, len(sizes))
	for _, size := range sizes {
		values = append(values, strconv.Itoa(size))
	}
	return strings.Join(values, ", ")
}

func guideMethods(operation labapi.OperationDescriptor, requested []string) ([]string, error) {
	if len(requested) == 0 && operation.Group == "addition" {
		requested = additionStarter().Variants
	}
	if len(requested) == 0 {
		for _, method := range operation.Strategies {
			if method.UsesSIMD && method.WorkerPolicy == catalog.Serial {
				requested = []string{catalog.Scalar, method.ID}
				break
			}
		}
	}
	if len(requested) != 2 || requested[0] == requested[1] || !slices.Contains(requested, catalog.Scalar) {
		return nil, errors.New("comparison requires the scalar baseline and one other method")
	}
	for _, id := range requested {
		if !slices.ContainsFunc(operation.Strategies, func(method labapi.StrategyDescriptor) bool { return method.ID == id }) {
			return nil, errors.New("unknown method")
		}
	}
	return requested, nil
}

func promptLine(ctx context.Context, reader *bufio.Scanner, stdout io.Writer, prompt string) (string, error) {
	if _, err := io.WriteString(stdout, prompt); err != nil {
		return "", err
	}
	type response struct {
		line string
		err  error
	}
	read := make(chan response, 1)
	go func() {
		if reader.Scan() {
			read <- response{line: strings.TrimSpace(reader.Text())}
			return
		}
		err := reader.Err()
		if err == nil {
			err = io.EOF
		}
		read <- response{err: err}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-read:
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return result.line, result.err
	}
}
