package app

import (
	"github.com/sirupsen/logrus"
	"protty/internal/adapter/cli"
	"protty/internal/infrastructure/config"
	"protty/internal/infrastructure/service"
)

func Start(cfg *config.StartCommandConfig) {
	logger := logrus.New()
	logger.SetLevel(cfg.GetLogLevelLogrus())
	logger.SetFormatter(&logrus.JSONFormatter{})

	rootCmd := cli.NewRootCommand()
	reverseProxySvc := service.NewReverseProxyService(logger)
	cli.NewStartCommand(cfg, reverseProxySvc, rootCmd.GetCobraCommand())
	_ = rootCmd.GetCobraCommand().Execute()
}
