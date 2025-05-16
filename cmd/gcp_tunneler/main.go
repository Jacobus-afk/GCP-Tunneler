package main

import (
	"gcp-tunneler/internal/run"

	"github.com/rs/zerolog/log"
)

func main() {
	app := &run.Application{
		Config: &run.RealConfiguration{},
	}
	err := app.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("application failed")
	}
}
