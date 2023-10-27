package main

import (
	"protty/internal/infrastructure/app"
	"protty/internal/infrastructure/config"
)

func main() {
	cfg := config.GetStartCommandConfig()
	prottyApp := app.NewProttyApp(cfg)
	// Ignore error cos it handles by cobra
	_ = prottyApp.Start()
}
