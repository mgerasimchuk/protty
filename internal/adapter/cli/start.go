package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"protty/internal/infrastructure/config"
	"protty/internal/infrastructure/service"
)

type StartCommand struct {
	cobraCmd        *cobra.Command
	reverseProxySvc *service.ReverseProxyService
	cfg             *config.StartCommandConfig
}

func NewStartCommand(cfg *config.StartCommandConfig, reverseProxySvc *service.ReverseProxyService, parentCommand *cobra.Command) *StartCommand {
	startCommand := &StartCommand{
		reverseProxySvc: reverseProxySvc,
		cfg:             cfg,
	}

	startCommand.cobraCmd = &cobra.Command{
		Use:     "start",
		Short:   "Start the proxy",
		PreRunE: startCommand.preRunE,
		RunE:    startCommand.runE,
	}

	parentCommand.AddCommand(startCommand.GetCobraCommand())

	startCommand.cobraCmd.Flags().SortFlags = false
	startCommand.cobraCmd.Flags().StringVar(buildFlagArgs(cfg.LogLevel))
	startCommand.cobraCmd.Flags().IntVar(buildFlagArgs(cfg.LocalPort))
	startCommand.cobraCmd.Flags().StringVar(buildFlagArgs(cfg.RemoteURI))
	startCommand.cobraCmd.Flags().Float64Var(buildFlagArgs(cfg.ThrottleRateLimit))
	startCommand.cobraCmd.Flags().StringVar(buildFlagArgs(cfg.ThrottleHost))

	startCommand.cobraCmd.Example = fmt.Sprintf("  %s --%s https://www.githubstatus.com --%s 2",
		startCommand.cobraCmd.CommandPath(), cfg.RemoteURI.GetFlagName(), cfg.ThrottleRateLimit.GetFlagName())

	startCommand.cobraCmd.IsAdditionalHelpTopicCommand()

	startCommand.cobraCmd.SetHelpTemplate(
		startCommand.cobraCmd.HelpTemplate() +
			"\n*Use CLI flags, environment variables or request headers to configure settings. " +
			"The settings will be applied in the following priority: environment variables -> CLI flags -> request headers\n")

	return startCommand
}

func (c *StartCommand) GetCobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *StartCommand) preRunE(cmd *cobra.Command, args []string) error {
	if err := c.cfg.ReadEnv(); err != nil {
		return err
	}

	// We should parse the flags only after load env variables
	return cmd.ParseFlags(os.Args[1:])
}

func (c *StartCommand) runE(cmd *cobra.Command, args []string) error {
	if err := c.cfg.Validate(); err != nil {
		return err
	}
	return c.reverseProxySvc.Start(c.cfg)
}

func buildFlagArgs[T config.OptionValueType](o config.Option[T]) (*T, string, T, string) {
	description := o.Description + fmt.Sprintf(" | Env variable alias: %s | Request header alias: %s", o.GetEnvName(), o.GetHeaderName())
	return &o.Value, o.GetFlagName(), o.Value, description
}
