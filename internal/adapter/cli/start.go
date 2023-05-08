package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"protty/internal/infrastructure/config"
	"protty/internal/infrastructure/service"
	"strings"
	"text/template"
)

type StartCommand struct {
	cobraCmd        *cobra.Command
	reverseProxySvc *service.ReverseProxyService
	cfg             *config.StartCommandConfig
	errSetFromEnv   error
}

func NewStartCommand(cfg *config.StartCommandConfig, reverseProxySvc *service.ReverseProxyService, parentCommand *cobra.Command) *StartCommand {
	startCommand := &StartCommand{
		reverseProxySvc: reverseProxySvc,
		cfg:             cfg,
		errSetFromEnv:   cfg.SetFromEnv(), // parse in this place, cos we should do it before command flag parsing (handling of the error places in the runE func)
	}

	startCommand.cobraCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the proxy",
		RunE:  startCommand.runE,
	}

	parentCommand.AddCommand(startCommand.GetCobraCommand())

	startCommand.cobraCmd.Flags().SortFlags = false
	startCommand.cobraCmd.Flags().StringVar(buildFlagArgs(&cfg.LogLevel))
	startCommand.cobraCmd.Flags().IntVar(buildFlagArgs(&cfg.LocalPort))
	startCommand.cobraCmd.Flags().StringVar(buildFlagArgs(&cfg.RemoteURI))
	startCommand.cobraCmd.Flags().Float64Var(buildFlagArgs(&cfg.ThrottleRateLimit))
	startCommand.cobraCmd.Flags().StringArrayVar(buildFlagArgs(&cfg.AdditionalRequestHeaders))
	startCommand.cobraCmd.Flags().StringArrayVar(buildFlagArgs(&cfg.TransformRequestBodySED))
	startCommand.cobraCmd.Flags().StringArrayVar(buildFlagArgs(&cfg.TransformRequestBodyJQ))
	startCommand.cobraCmd.Flags().StringArrayVar(buildFlagArgs(&cfg.AdditionalResponseHeaders))
	startCommand.cobraCmd.Flags().StringArrayVar(buildFlagArgs(&cfg.TransformResponseBodySED))
	startCommand.cobraCmd.Flags().StringArrayVar(buildFlagArgs(&cfg.TransformResponseBodyJQ))

	startCommand.cobraCmd.Example = startCommand.getExamples()
	startCommand.cobraCmd.SetHelpTemplate(
		startCommand.cobraCmd.HelpTemplate() +
			"\n*Use CLI flags, environment variables or request headers to configure settings. " +
			"The settings will be applied in the following priority: environment variables -> CLI flags -> request headers\n")

	return startCommand
}

func (c *StartCommand) GetCobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *StartCommand) runE(cmd *cobra.Command, args []string) error {
	if c.errSetFromEnv != nil {
		return c.errSetFromEnv
	}
	if err := c.cfg.Validate(); err != nil {
		return err
	}
	return c.reverseProxySvc.Start(c.cfg)
}

func (c *StartCommand) getExamples() string {
	textTemplate := `  # Start the proxy with default values
  {{ .Cmd.CommandPath }}
  
  # Start the proxy with specific log level
  {{ .Cmd.CommandPath }} --{{ .Cfg.LogLevel.GetFlagName }} info

  # Start the proxy with a specific local port
  {{ .Cmd.CommandPath }} --{{ .Cfg.LocalPort.GetFlagName }} 8080
  
  # Start the proxy with a specific remote URI and specific throttle rate limit 
  {{ .Cmd.CommandPath }} --{{ .Cfg.RemoteURI.GetFlagName }} https://www.githubstatus.com --{{ .Cfg.ThrottleRateLimit.GetFlagName }} 2

  # Start the proxy with a specific additional request headers
  {{ .Cmd.CommandPath }} --{{ .Cfg.AdditionalRequestHeaders.GetFlagName }} 'Authorization: Bearer authtoken-with:any:symbols' --{{ .Cfg.AdditionalRequestHeaders.GetFlagName }} 'X-Another-One: another-value'

  # Start the proxy with a specific SED expression for response transformation
  {{ .Cmd.CommandPath }} --{{ .Cfg.TransformResponseBodySED.GetFlagName }} 's|old|new|g'

  # Start the proxy with a specific SED expressions pipeline for response transformation
  {{ .Cmd.CommandPath }} --{{ .Cfg.TransformResponseBodySED.GetFlagName }} 's|old|new-stage-1|g' --{{ .Cfg.TransformResponseBodySED.GetFlagName }} 's|new-stage-1|new-stage-2|g'

  # Start the proxy with a specific SED expressions pipeline for response transformation (configured with env)
  {{ .Cfg.TransformResponseBodySED.GetEnvName }}_0='s|old|new-stage-1|g' {{ .Cfg.TransformResponseBodySED.GetEnvName }}_1='s|new-stage-1|new-stage-2|g' {{ .Cmd.CommandPath }}

  # Start the proxy with a specific JQ expressions pipeline for response transformation
  {{ .Cmd.CommandPath }} --{{ .Cfg.TransformResponseBodyJQ.GetFlagName }} '.[] | .id'`

	t, b := new(template.Template), new(strings.Builder)
	err := template.Must(t.Parse(textTemplate)).Execute(b, struct {
		Cmd *cobra.Command
		Cfg *config.StartCommandConfig
	}{c.cobraCmd, c.cfg})
	if err != nil {
		panic(err)
	}
	return b.String()
}

func buildFlagArgs[T config.OptionValueType](o *config.Option[T]) (*T, string, T, string) {
	o.MarkAsAddedToCLI()
	description := o.Description + fmt.Sprintf(" | Env variable alias: %s | Request header alias: %s", o.GetEnvName(), o.GetHeaderName())
	return &o.Value, o.GetFlagName(), o.Value, description
}
