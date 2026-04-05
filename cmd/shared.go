package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

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

func rejectUnsupportedAllFlag(cmd *cli.Command) error {
	if cmd.Bool("all") {
		return exitWith("--all is only supported for 'ports' and 'ports ps'", 1)
	}
	return nil
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
