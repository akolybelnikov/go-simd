package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/labservice"
)

func runUI(service *labservice.Deterministic, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("ui", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("artifacts", filepath.Join("artifacts", "experiments"), "local artifact directory")
	noOpen := flags.Bool("no-open", false, "print the private browser URL without opening it")
	useOllama := flags.Bool("ollama", false, "enable Ollama for the experimental comparison preview API")
	model := flags.String("model", "qwen3:4b", "installed local Ollama model")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *root == "" {
		fmt.Fprintln(stderr, "usage: simdlab ui [options]")
		return 2
	}

	var client agent.ModelClient
	if *useOllama {
		local, err := newLocalModelClient("http://127.0.0.1:11434", *model, uiModelTimeout, false)
		if err != nil {
			fmt.Fprintln(stderr, terminalText(err.Error()))
			return 2
		}
		client = local
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer listener.Close()
	app, err := newLocalUI(listener.Addr().String(), *root, service, client)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := &http.Server{
		Handler: app, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second,
		WriteTimeout: 150 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()

	url := app.origin + "/#" + app.token
	fmt.Fprintf(stdout, "Open the local lab in your browser:\n%s\nPress Ctrl+C to stop.\n", url)
	if !*noOpen {
		if err := openLocalUIBrowser(url); err != nil {
			fmt.Fprintln(stderr, "Could not open a browser automatically. Use the private URL above.")
		}
	}

	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
		}
	case err := <-done:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	return 0
}

func openLocalUIBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url)
	case "darwin":
		command = exec.Command("open", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	command.Stdin = nil
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}
