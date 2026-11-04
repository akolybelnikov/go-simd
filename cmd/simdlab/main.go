// simdlab is the shared deterministic laboratory and future agent entry point.
// The _worker command is private: parent processes provide its bounded request.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labworker"
)

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
	fmt.Fprintln(stderr, "usage: simdlab <command>")
	fmt.Fprintln(stderr, "public laboratory commands are not enabled yet")
	return 2
}
