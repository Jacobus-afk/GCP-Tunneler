package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"

	// "gcp-tunneler/config"
	"gcp-tunneler/internal/config"
	gcptunneler "gcp-tunneler/internal/gcp_api"
	"gcp-tunneler/internal/menu"
	"gcp-tunneler/internal/tunnelbuilder"
	"gcp-tunneler/internal/utils"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ResourceType int

const (
	BackendResource ResourceType = iota
	InstanceResource
)

func (r ResourceType) String() string {
	return [...]string{"backends", "instances"}[r]
}

func main() {
	cfg := config.GetConfig()
	var reloadConfig bool
	flag.BoolVar(
		&reloadConfig,
		"reload-config",
		false,
		"whether or not to repopulate the instance list from GCP",
	)
	flag.Parse()
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// If instanceFile doesn't exist, or --reload-config flag
	if _, err := os.Stat(cfg.InstanceFilename); errors.Is(err, os.ErrNotExist) {
		reloadConfig = true
	}

	if reloadConfig {
		ctx := context.Background()

		projects := gcptunneler.ListProjects(ctx)
		projectDataList := gcptunneler.GetInstancesByProject(ctx, projects)
		jsonData, err := json.MarshalIndent(projectDataList, "", "  ")
		if err != nil {
			log.Fatal().Err(err).Msg("error marshaling to JSON")
		}

		log.Info().Str("config_file", cfg.InstanceFilename).Msg("Writing configuration to file...")

		os.WriteFile(cfg.InstanceFilename, jsonData, 0644)

		// ------------DEBUG stuff
		// for _, project := range projects {
		// 	fmt.Println(project)
		// 	instances := gcptunneler.ListInstances(ctx, project)
		// 	for _, instance := range instances {
		// 		fmt.Println(instance)
		// 	}
		//
		// }

		// for _, data := range projectDataList {
		// 	log.Println(data)
		// }

		// log.Println(string(jsonData))
	}

	resourceNames := menu.HandleFZFMenu()
	sessionName, err := tunnelbuilder.BuildTunnelAndSSH(resourceNames)
	if err != nil {
		log.Fatal().Err(err).Msg("error building tunnels")
	}

	switchErr := utils.SwitchToCreatedSession(sessionName)
	if switchErr != nil {
		log.Fatal().Err(switchErr).Msg("couldn't switch to tmux session")
	}

}
