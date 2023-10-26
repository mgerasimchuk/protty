package cli

import (
	"github.com/spf13/cobra"
)

type RootCommand struct {
	cobraCmd *cobra.Command
}

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{}

	rootCommand.cobraCmd = &cobra.Command{
		Use:   "protty",
		Short: "Protty is a HTTP proxy written in Go that redirects requests to a remote host",
	}

	rootCommand.cobraCmd.CompletionOptions.HiddenDefaultCmd = true

	return rootCommand
}

func (c *RootCommand) GetCobraCommand() *cobra.Command {
	return c.cobraCmd
}
