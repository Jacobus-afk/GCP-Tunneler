package main

import (
	"flag"
	"gcp-tunneler/internal/config"
	"gcp-tunneler/internal/run"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var version = "undefined"

func main() {
	configureLogger()

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
		Config: &run.RealConfiguration{},
	}
	err := app.Run(*reloadCfgFlag, envCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("application failed")
	}
}

func printVersion()  {
	os.Stdout.WriteString(version + "\n")
}

func configureLogger() {
	// undecided if needed
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}
