package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LarsEckart/ports/render"
	"github.com/LarsEckart/ports/scanner"
	"github.com/urfave/cli/v3"
)

func WatchCmd() *cli.Command {
	return &cli.Command{
		Name:               "watch",
		Usage:              "Monitor port changes in real time",
		CustomHelpTemplate: commandHelpTemplateNoGlobals,
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:  "interval",
				Usage: "Polling interval",
				Value: 2 * time.Second,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := rejectUnsupportedAllFlag(cmd); err != nil {
				return err
			}

			watchCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()

			render.DisplayWatchHeader(os.Stdout)
			return scanner.WatchPorts(watchCtx, cmd.Duration("interval"), func(eventType string, info scanner.PortInfo) {
				render.DisplayWatchEvent(os.Stdout, eventType, info)
			})
		},
	}
}
