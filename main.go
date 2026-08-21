package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/LarsEckart/ports/cmd"
)

type exitCoder interface {
	error
	ExitCode() int
}

func main() {
	app := cmd.NewApp()
	if err := app.Run(context.Background(), os.Args); err != nil {
		if ec, ok := errors.AsType[exitCoder](err); ok {
			if msg := err.Error(); msg != "" {
				fmt.Fprintln(os.Stderr, msg)
			}
			os.Exit(ec.ExitCode())
		}

		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
