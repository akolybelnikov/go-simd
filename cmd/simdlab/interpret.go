package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labinterpret"
	"github.com/akolybelnikov/go-simd/internal/labservice"
)

func runInterpret(service *labservice.Deterministic, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("interpret", flag.ContinueOnError)
	flags.SetOutput(stderr)
	preview := flags.Bool("preview", false, "enable model-assisted interpretation")
	structured := flags.Bool("json", false, "emit structured JSON including evidence IDs")
	endpoint := flags.String("endpoint", "http://127.0.0.1:11434", "Ollama server base URL")
	model := flags.String("model", "", "installed Ollama model (explicit selection; experimental)")
	think := flags.Bool("think", true, "request Ollama's separate thinking channel")
	timeout := flags.Duration("timeout", 2*time.Minute, "model deadline (maximum 5m)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if !*preview || flags.NArg() == 0 || *timeout <= 0 || *timeout > 5*time.Minute {
		fmt.Fprintln(stderr, "interpret requires -preview, a capture file, and a timeout greater than zero and at most 5m")
		return 2
	}
	question := strings.TrimSpace(strings.Join(flags.Args()[1:], " "))
	if question == "" {
		question = "Summarize the measured kernel timings and any compatible scalar comparisons."
	}
	if len(question) > maxQuestionBytes {
		fmt.Fprintf(stderr, "question exceeds %d bytes\n", maxQuestionBytes)
		return 2
	}
	file, err := os.Open(flags.Arg(0))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	facts, err := service.Inspect(file, evidence.DefaultReadLimits())
	file.Close()
	if err != nil {
		fmt.Fprintln(stderr, terminalText(err.Error()))
		return 1
	}
	client, err := newLocalModelClient(*endpoint, *model, *timeout, *think)
	if err != nil {
		fmt.Fprintln(stderr, terminalText(err.Error()))
		return 2
	}
	fmt.Fprintln(stderr, "Experimental Ollama model:", terminalText(*model))
	interrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(interrupt, *timeout)
	defer cancel()
	answer, modelErr := labinterpret.Interpret(ctx, client, facts, question)
	var code int
	if *structured {
		code = writeJSON(stdout, stderr, answer)
	} else {
		code = writeText(stdout, stderr, interpretationText(answer))
	}
	if code != 0 {
		return 1
	}
	if modelErr != nil {
		fmt.Fprintln(stderr, "Could not validate an evidence selection. Use simdlab inspect for the deterministic summary.")
		return 1
	}
	return 0
}

func interpretationText(answer labinterpret.Answer) string {
	var text strings.Builder
	if answer.Decision == "answer" && len(answer.Facts) > 0 {
		fmt.Fprintln(&text, "What the measurements show")
		for _, fact := range answer.Facts {
			fmt.Fprintln(&text, "-", fact.Display)
		}
	} else {
		fmt.Fprintln(&text, "The saved evidence is insufficient to answer this question.")
	}
	if len(answer.Caveats) > 0 {
		fmt.Fprintln(&text, "\nWhat to keep in mind")
		for _, caveat := range answer.Caveats {
			fmt.Fprintln(&text, "-", caveat.Message)
		}
	}
	if answer.MeasurementNotice != "" {
		fmt.Fprintln(&text, "\n"+answer.MeasurementNotice)
	}
	return text.String()
}
