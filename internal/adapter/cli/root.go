package cli

import (
	"github.com/spf13/cobra"
)

type RootCommand struct {
	cobraCmd *cobra.Command
}

const asciiLogo = `
      ###################
    #######################
   #########################     ____    ____     ___    _____   _____  __   __
   #######      ###  #######    |  _ \  |  _ \   / _ \  |_   _| |_   _| \ \ / /
   #######    ###### #######    | |_) | | |_) | | | | |   | |     | |    \ V /
   #######  #######  #######    |  __/  |  _ <  | |_| |   | |     | |     | |
   ####### ######    #######    |_|     |_| \_\  \___/    |_|     |_|     |_|
   #######  ###      #######
    #####             #####
`

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{}

	rootCommand.cobraCmd = &cobra.Command{
		Use:   "protty",
		Short: asciiLogo + "\n" + "HTTP proxy interceptor with on the fly request/response transforming capabilities",
	}

	rootCommand.cobraCmd.CompletionOptions.HiddenDefaultCmd = true

	return rootCommand
}

func (c *RootCommand) GetCobraCommand() *cobra.Command {
	return c.cobraCmd
}
