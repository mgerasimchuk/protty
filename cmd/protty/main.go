package main

import (
	"protty/internal/infrastructure/app"
	"protty/internal/infrastructure/config"
)

func main() {
	cfg := config.GetConfig()
	app.Start(cfg)
}
