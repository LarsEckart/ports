package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/LarsEckart/ports/scanner"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"
)

func KillCmd() *cli.Command {
	return &cli.Command{
		Name:  "kill",
		Usage: "Kill by port or PID",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Use SIGKILL instead of SIGTERM",
			},
			&cli.BoolFlag{
				Name:  "pid",
				Usage: "Interpret every argument as a PID",
			},
			&cli.BoolFlag{
				Name:  "port",
				Usage: "Interpret every argument as a port",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) == 0 {
				return exitWith("usage: ports kill [-f|--force] [--pid|--port] <port|pid> [port|pid...]", 1)
			}

			if cmd.Bool("pid") && cmd.Bool("port") {
				return exitWith("choose only one of --pid or --port", 1)
			}

			force := cmd.Bool("force")
			byPID := cmd.Bool("pid")
			byPort := cmd.Bool("port")
			var anyFailed bool
			signal := "SIGTERM"
			if force {
				signal = "SIGKILL"
			}

			white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			green := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
			red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

			fmt.Println()
			for _, arg := range args {
				n, err := strconv.Atoi(arg)
				if err != nil {
					fmt.Println(red.Render("  ✕ " + fmt.Sprintf("%q is not a valid port or PID", arg)))
					anyFailed = true
					continue
				}

				var resolved *scanner.KillTarget
				switch {
				case byPID:
					resolved = scanner.ResolveKillPID(n)
				case byPort:
					resolved, err = scanner.ResolveKillPort(ctx, n)
				default:
					resolved, err = scanner.ResolveKillTarget(ctx, n)
				}
				if err != nil {
					if errors.Is(err, scanner.ErrKillTargetAmbiguous) {
						fmt.Println(red.Render(fmt.Sprintf("  ✕ %d matches both a listening port and a PID; use --port or --pid", n)))
						anyFailed = true
						continue
					}
					return err
				}
				if resolved == nil {
					switch {
					case byPort:
						fmt.Println(red.Render(fmt.Sprintf("  ✕ No listener on :%d", n)))
					case byPID || n > 65535:
						fmt.Println(red.Render(fmt.Sprintf("  ✕ No process with PID %d", n)))
					default:
						fmt.Println(red.Render(fmt.Sprintf("  ✕ No listener on :%d and no process with PID %d", n, n)))
					}
					anyFailed = true
					continue
				}

				label := fmt.Sprintf("PID %d", resolved.PID)
				if resolved.Via == "port" && resolved.Info != nil {
					label = fmt.Sprintf(":%d — %s (PID %d)", resolved.Port, resolved.Info.ProcessName, resolved.PID)
				}

				fmt.Println(white.Render("  Killing " + label))
				if err := scanner.KillProcess(resolved.PID, force); err != nil {
					fmt.Println(red.Render(fmt.Sprintf("  ✕ Failed to send %s to %s", signal, label)))
					anyFailed = true
					continue
				}

				fmt.Println(green.Render(fmt.Sprintf("  ✓ Sent %s to %s", signal, label)))
			}
			fmt.Println()

			if anyFailed {
				return exitWith("", 1)
			}
			return nil
		},
	}
}
