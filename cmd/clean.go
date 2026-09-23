package cmd

import (
	"context"
	"os"

	"github.com/LarsEckart/ports/render"
	"github.com/LarsEckart/ports/scanner"
	"github.com/urfave/cli/v3"
)

func CleanCmd() *cli.Command {
	return &cli.Command{
		Name:               "clean",
		Usage:              "Kill orphaned or zombie dev processes",
		CustomHelpTemplate: commandHelpTemplateNoGlobals,
		OnUsageError:       onUsageError,
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
		Action: cleanAction,
	}
}

func cleanAction(ctx context.Context, cmd *cli.Command) error {
	orphaned, err := scanner.FindOrphanedProcesses(ctx)
	if err != nil {
		return err
	}
	if len(orphaned) == 0 {
		render.DisplayCleanResults(os.Stdout, nil, nil, nil)
		return nil
	}

	if !cmd.Bool("yes") {
		render.DisplayCleanConfirmation(os.Stdout, orphaned)
		ok, err := confirm("Kill all? [y/N] ")
		if err != nil {
			return err
		}
		if !ok {
			render.DisplayCleanAborted(os.Stdout)
			return nil
		}
	}

	killed, failed := killOrphaned(orphaned, cmd.Bool("force"))
	render.DisplayCleanResults(os.Stdout, orphaned, killed, failed)
	if len(failed) > 0 {
		return exitWith("", exitCodeFailure)
	}
	return nil
}

func killOrphaned(orphaned []scanner.PortInfo, force bool) (killed, failed []int) {
	killed = make([]int, 0, len(orphaned))
	for _, port := range orphaned {
		if err := scanner.KillProcess(port.PID, force); err != nil {
			failed = append(failed, port.PID)
			continue
		}
		killed = append(killed, port.PID)
	}
	return killed, failed
}
