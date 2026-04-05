package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/LarsEckart/ports/render"
	"github.com/LarsEckart/ports/scanner"
	"github.com/urfave/cli/v3"
)

func NewApp() *cli.Command {
	return &cli.Command{
		Name:  "ports",
		Usage: "See what is listening on your machine",
		Description: `A fast Go CLI for developers who want to know what is bound to their ports.

Examples:
  ports
  ports --all
  ports 3000
  ports ps
  ports kill 3000
  ports clean --yes
  ports watch`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show all listening ports, not just dev-ish ones",
			},
		},
		Commands: []*cli.Command{
			PSCmd(),
			KillCmd(),
			CleanCmd(),
			WatchCmd(),
		},
		Action: rootAction,
	}
}

func rootAction(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	showAll := cmd.Bool("all")

	if len(args) == 0 {
		ports, err := scanner.GetListeningPorts(ctx, false)
		if err != nil {
			return err
		}
		if !showAll {
			ports = filterDevPorts(ports)
		}
		render.DisplayPortTable(os.Stdout, ports, !showAll)
		return nil
	}

	if len(args) == 1 {
		port, err := strconv.Atoi(args[0])
		if err == nil {
			info, err := scanner.GetPortDetails(ctx, port)
			if err != nil {
				return err
			}
			render.DisplayPortDetail(os.Stdout, info)
			if info == nil {
				return exitWith("", 1)
			}
			return nil
		}
	}

	return exitWith(fmt.Sprintf("unknown command or argument: %s", args[0]), 1)
}

func filterDevPorts(ports []scanner.PortInfo) []scanner.PortInfo {
	filtered := make([]scanner.PortInfo, 0, len(ports))
	for _, port := range ports {
		if scanner.IsDevProcess(port.ProcessName, port.Command) {
			filtered = append(filtered, port)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Port < filtered[j].Port })
	return filtered
}

func filterDevProcesses(processes []scanner.ProcessInfo) []scanner.ProcessInfo {
	filtered := make([]scanner.ProcessInfo, 0, len(processes))
	for _, process := range processes {
		if scanner.IsDevProcess(process.ProcessName, process.Command) {
			filtered = append(filtered, process)
		}
	}
	return filtered
}
