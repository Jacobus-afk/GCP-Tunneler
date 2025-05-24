package main

import (
	"flag"
	"gcp-tunneler/internal/config"
	"gcp-tunneler/internal/menu"
	"gcp-tunneler/internal/run"
	"gcp-tunneler/internal/tunnelbuilder"
	"gcp-tunneler/internal/utils"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var version = "undefined"

func main() {
	configureLogger()
	// tos
	// TODO: add flags for other configurable values?
	reloadCfgFlag := flag.Bool(
		"reload-config",
		false,
		"whether or not to repopulate the instance list from GCP",
	)

	verFlag := flag.Bool(
		"version",
		false,
		"print the current program version",
	)

	flag.Parse()

	if *verFlag {
		printVersion()
		os.Exit(0)
	}

	envCfg := config.GetConfig()

	if envCfg.Develop.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Debug().Msg("Debugging: enabled")

	app := &run.Application{
		Config:         &run.RealConfiguration{},
		TunnelBuilder:  &tunnelbuilder.Builder{},
		MenuHandler:    &menu.Menu{},
		SessionHandler: &utils.Session{},
	}
	err := app.Run(*reloadCfgFlag, envCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("application failed")
	}
}

func printVersion() {
	_, _ = os.Stdout.WriteString(version + "\n")
}

func configureLogger() {
	// needed, or not?
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}
