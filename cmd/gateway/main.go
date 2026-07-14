package main

import (
	"os"

	"github.com/rejzzzz/goGate/internal/app"
)

func main() {
	configPath := "configs/gateway.yaml"
	if envPath := os.Getenv("GATEWAY_CONFIG"); envPath != "" {
		configPath = envPath
	}

	app.Run(configPath)
}
