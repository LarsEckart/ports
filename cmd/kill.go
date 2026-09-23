package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/LarsEckart/ports/render"
	"github.com/LarsEckart/ports/scanner"
	"github.com/urfave/cli/v3"
)

type killMode uint8

const (
	killAuto killMode = iota
	killByPID
	killByPort
)

func KillCmd() *cli.Command {
	return &cli.Command{
		Name:               "kill",
		Usage:              "Kill by port or PID",
		UsageText:          "ports kill [options] <port|pid> [port|pid...]",
		CustomHelpTemplate: commandHelpTemplateNoGlobals,
		OnUsageError:       onUsageError,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Use SIGKILL instead of SIGTERM",
			},
			&cli.BoolFlag{
				Name:  "pid",
				Local: true,
				Usage: "Interpret every argument as a PID",
			},
			&cli.BoolFlag{
				Name:  "port",
				Local: true,
				Usage: "Interpret every argument as a port",
			},
		},
		Action: killAction,
	}
}

func killAction(ctx context.Context, cmd *cli.Command) error {
	mode, err := killSelection(ctx, cmd)
	if err != nil {
		return err
	}
	args := cmd.Args().Slice()
	force := cmd.Bool("force")
	var anyFailed bool
	fmt.Println()
	for _, arg := range args {
		failed, err := killArgument(ctx, arg, mode, force)
		if err != nil {
			return err
		}
		anyFailed = anyFailed || failed
	}
	fmt.Println()
	if anyFailed {
		return exitWith("", exitCodeFailure)
	}
	return nil
}

func killSelection(ctx context.Context, cmd *cli.Command) (killMode, error) {
	if cmd.Args().Len() == 0 {
		return killAuto, usageErrorWithHelp(ctx, cmd, "usage: ports kill [-f|--force] [--pid|--port] <port|pid> [port|pid...]")
	}
	if cmd.Bool("pid") && cmd.Bool("port") {
		return killAuto, usageErrorWithHelp(ctx, cmd, "choose only one of --pid or --port")
	}
	if cmd.Bool("pid") {
		return killByPID, nil
	}
	if cmd.Bool("port") {
		return killByPort, nil
	}
	return killAuto, nil
}

func killArgument(ctx context.Context, arg string, mode killMode, force bool) (bool, error) {
	n, err := strconv.Atoi(arg)
	if err != nil {
		render.DisplayInvalidKillArgument(os.Stdout, arg)
		return true, nil
	}

	var target *scanner.KillTarget
	switch mode {
	case killByPID:
		target = scanner.ResolveKillPID(n)
	case killByPort:
		target, err = scanner.ResolveKillPort(ctx, n)
	default:
		target, err = scanner.ResolveKillTarget(ctx, n)
	}
	if errors.Is(err, scanner.ErrKillTargetAmbiguous) {
		render.DisplayAmbiguousKillTarget(os.Stdout, n)
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if target == nil {
		render.DisplayMissingKillTarget(os.Stdout, n, mode == killByPID, mode == killByPort)
		return true, nil
	}

	render.DisplayKilling(os.Stdout, target)
	if err := scanner.KillProcess(target.PID, force); err != nil {
		render.DisplayKillResult(os.Stdout, target, force, false)
		return true, nil
	}
	render.DisplayKillResult(os.Stdout, target, force, true)
	return false, nil
}
