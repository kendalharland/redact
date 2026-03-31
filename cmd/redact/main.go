// Command redact removes sensitive data from text.
package main

import (
	"fmt"
	"os"

	"github.com/kendalharland/redact/internal/cli"
	"github.com/kendalharland/redact/internal/evals"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "evals" {
		// Run evaluation subcommand
		if err := runEvals(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Parse flags
	opts, err := cli.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Run the application
	app := cli.NewApp(opts.Model)
	if err := app.Run(opts, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runEvals(args []string) error {
	opts, err := evals.ParseFlags(args)
	if err != nil {
		return err
	}

	runner := evals.NewRunner(opts.Model)
	return runner.Run(opts, os.Stdout, os.Stderr)
}
