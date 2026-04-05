package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

const (
	exitCodeFailure = 1
	exitCodeUsage   = 2
)

const rootHelpTemplateNoGlobals = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}} {{if .VisibleFlags}}[options]{{end}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}
{{- if len .Authors}}

AUTHOR{{template "authorsTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{template "visibleCommandCategoryTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{template "copyrightTemplate" .}}{{end}}
`

const commandHelpTemplateNoGlobals = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{template "usageTemplate" .}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleFlagCategories}}

OPTIONS:{{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}
`

type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string {
	return e.msg
}

func (e *exitError) ExitCode() int {
	return e.code
}

func exitWith(message string, code int) error {
	return &exitError{code: code, msg: message}
}

func usageErrorWithHelp(ctx context.Context, cmd *cli.Command, message string) error {
	fmt.Fprintf(cmd.Root().ErrWriter, "Incorrect Usage: %s\n\n", message)

	if err := showUsageHelp(ctx, cmd); err != nil {
		return err
	}

	return exitWith("", exitCodeUsage)
}

func onUsageError(ctx context.Context, cmd *cli.Command, err error, _ bool) error {
	return usageErrorWithHelp(ctx, cmd, err.Error())
}

func showUsageHelp(ctx context.Context, cmd *cli.Command) error {
	if cmd.Root() == cmd {
		return cli.ShowRootCommandHelp(cmd)
	}

	return cli.ShowSubcommandHelp(cmd)
}

func confirm(prompt string) (bool, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(text))
	return answer == "y" || answer == "yes", nil
}
