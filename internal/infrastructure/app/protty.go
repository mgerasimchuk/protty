package app

import (
	"context"
	"github.com/sirupsen/logrus"
	"protty/internal/adapter/cli"
	"protty/internal/infrastructure/config"
	"protty/internal/infrastructure/service"
)

type ProttyApp struct {
	logger          *logrus.Logger
	rootCmd         *cli.RootCommand
	startCmd        *cli.StartCommand
	reverseProxySvc *service.ReverseProxyService
}

func NewProttyApp(cfg *config.StartCommandConfig) *ProttyApp {
	app := &ProttyApp{}
	app.logger = logrus.New()
	app.logger.SetLevel(cfg.GetLogLevelLogrus())
	app.logger.SetFormatter(&logrus.JSONFormatter{})
	app.rootCmd = cli.NewRootCommand()
	app.reverseProxySvc = service.NewReverseProxyService(app.logger)
	app.startCmd = cli.NewStartCommand(cfg, app.reverseProxySvc, app.rootCmd.GetCobraCommand())
	return app
}

func (a *ProttyApp) Start() error {
	return a.rootCmd.GetCobraCommand().Execute()
}

func (a *ProttyApp) Stop(ctx context.Context) error {
	return a.reverseProxySvc.Stop(ctx)
}
