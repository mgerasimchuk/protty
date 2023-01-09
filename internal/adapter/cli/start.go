package cli

import (
	"github.com/spf13/cobra"
	"os"
	"protty/internal/infrastructure/config"
	"protty/internal/infrastructure/service"
)

type StartCommand struct {
	cobraCmd        *cobra.Command
	reverseProxySvc *service.ReverseProxyService
	cfg             *config.Config
}

func NewStartCommand(cfg *config.Config, reverseProxySvc *service.ReverseProxyService) *StartCommand {
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

	startCommand.cobraCmd.Flags().SortFlags = false
	startCommand.cobraCmd.Flags().StringVar(&cfg.LogLevel.Value, "log-level", cfg.LogLevel.Value, "Verbosity level (panic, fatal, error, warn, info, debug, trace)")
	startCommand.cobraCmd.Flags().IntVar(&cfg.LocalPort.Value, "local-port", cfg.LocalPort.Value, "Listening port for the proxy")
	startCommand.cobraCmd.Flags().StringVar(&cfg.RemoteURI.Value, "remote-uri", cfg.RemoteURI.Value, "URI of the remote resource")
	startCommand.cobraCmd.Flags().Float64Var(&cfg.ThrottleRateLimit.Value, "throttle-rate-limit", cfg.ThrottleRateLimit.Value, "How many requests can be send to the remote resource per second")
	startCommand.cobraCmd.Flags().StringVar(&cfg.ThrottleHost.Value, "throttle-host", cfg.ThrottleHost.Value, "On which host, the throttle rate limit should be applied")

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
