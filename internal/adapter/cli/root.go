package cli

import (
	"github.com/spf13/cobra"
	"protty/internal/infrastructure/service"
)

type RootCommand struct {
	cobraCmd        *cobra.Command
	reverseProxySvc *service.ReverseProxyService
}

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{}

	rootCommand.cobraCmd = &cobra.Command{
		Use:   "protty",
		Short: "Protty is a HTTP proxy written in Go that redirects requests to a remote host",
	}

	rootCommand.cobraCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCommand.cobraCmd.SetHelpTemplate(
		rootCommand.cobraCmd.HelpTemplate() +
			"\n*Use environment variables (for example, REMOTE_URI) or request headers (for example, X-PROTTY-REMOTE-URI) to configure settings. The settings will be applied in the following priority: environment variables -> command flags -> request headers\n")

	return rootCommand
}

func (c *RootCommand) GetCobraCommand() *cobra.Command {
	return c.cobraCmd
}
