package cmd

import (
	"context"
	"os"
	"sort"

	"github.com/LarsEckart/ports/render"
	"github.com/LarsEckart/ports/scanner"
	"github.com/urfave/cli/v3"
)

func PSCmd() *cli.Command {
	return &cli.Command{
		Name:  "ps",
		Usage: "Show running processes, not just listening ports",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show all processes, not just dev-ish ones",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			processes, err := scanner.GetAllProcesses(ctx)
			if err != nil {
				return err
			}

			showAll := cmd.Bool("all")
			if !showAll {
				processes = filterDevProcesses(processes)
				processes = scanner.CollapseDockerProcesses(processes)
			}

			sort.Slice(processes, func(i, j int) bool {
				return processes[i].CPU > processes[j].CPU
			})
			render.DisplayProcessTable(os.Stdout, processes, !showAll)
			return nil
		},
	}
}
