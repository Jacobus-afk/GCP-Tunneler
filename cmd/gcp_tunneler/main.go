package main

import (
	"flag"
	"gcp-tunneler/internal/config"
	"gcp-tunneler/internal/run"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	configureLogger()

	// TODO: add flags for other configurable values?
	reloadCfgFlag := flag.Bool(
		"reload-config",
		false,
		"whether or not to repopulate the instance list from GCP",
	)

	flag.Parse()
	envCfg := config.GetConfig()

	if envCfg.Develop.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Debug().Msg("Debugging: enabled")

	app := &run.Application{
		Config: &run.RealConfiguration{},
	}
	err := app.Run(*reloadCfgFlag, envCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("application failed")
	}
}

func configureLogger() {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}
