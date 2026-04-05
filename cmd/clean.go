package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/LarsEckart/ports/render"
	"github.com/LarsEckart/ports/scanner"
	"github.com/urfave/cli/v3"
)

func CleanCmd() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "Kill orphaned or zombie dev processes",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Skip confirmation",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Use SIGKILL instead of SIGTERM",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			orphaned, err := scanner.FindOrphanedProcesses(ctx)
			if err != nil {
				return err
			}
			if len(orphaned) == 0 {
				render.DisplayCleanResults(os.Stdout, nil, nil, nil)
				return nil
			}

			if !cmd.Bool("yes") {
				fmt.Println()
				fmt.Printf("Found %d orphaned/zombie process(es):\n", len(orphaned))
				for _, port := range orphaned {
					fmt.Printf("  - :%d — %s (PID %d)\n", port.Port, port.ProcessName, port.PID)
				}
				fmt.Println()
				ok, err := confirm("Kill all? [y/N] ")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Println("Aborted.")
					return nil
				}
			}

			killed := make([]int, 0, len(orphaned))
			failed := make([]int, 0)
			for _, port := range orphaned {
				if err := scanner.KillProcess(port.PID, cmd.Bool("force")); err != nil {
					failed = append(failed, port.PID)
					continue
				}
				killed = append(killed, port.PID)
			}

			render.DisplayCleanResults(os.Stdout, orphaned, killed, failed)
			if len(failed) > 0 {
				return exitWith("", 1)
			}
			return nil
		},
	}
}
