package main

import (
	"log"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/pardnchiu/PodRun/internal/command"
	"github.com/pardnchiu/PodRun/internal/utils"
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("failed to load .env",
			slog.String("error", err.Error()))
	}
}

func main() {
	if err := utils.CheckRelyPackages(); err != nil {
		log.Fatalf("missing required packages: %s", err)
	}

	if _, err := utils.CheckENV(); err != nil {
		log.Fatalf("missing required environment: %s", err)
	}

	cmd, err := command.New()
	if err != nil {
		log.Fatalf("failed to create command: %s", err)
	}

	if err := utils.SSHTest(); err != nil {
		log.Fatalf("failed to connect to remote server: %s", err)
	}

	switch cmd.RemoteArgs[0] {
	case "domain":
	case "deploy":
	default:
		result, err := cmd.ComposeCMD()
		if err != nil {
			slog.Error("failed to connect to remote server",
				"err", err)
		}
		slog.Info("", "result", result)
		// case "rm":
		// case "ports":
		// case "export":
	}
}
