package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

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

func loadConfiguration() error {
	cfg := config.GetConfig()


	reloadConfig := parseCmdLineArgs()
	configGCPResourceFileExists := utils.CheckIfFileExists(cfg.InstanceFilename)

	if reloadConfig || !configGCPResourceFileExists {
		projectDataList := populateGCPResources()

		log.Info().
			Str("config_file", cfg.InstanceFilename).
			Msg("writing GCP resource details to file...")

		jsonData, jsonErr := json.MarshalIndent(projectDataList, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("error marshaling to JSON: %w", jsonErr)
		}

		if writeErr := os.WriteFile(cfg.InstanceFilename, jsonData, 0644); writeErr != nil {
			return fmt.Errorf("couldn't write GCP resource details to file: %w", writeErr)
		}

	}
	return nil
}

func selectResources() string {
	resourceNames := menu.HandleFZFMenu()
	return resourceNames
}

func connectToResources(resourceNames string) (string, error) {
	sessionName, err := tunnelbuilder.BuildTunnelAndSSH(resourceNames)
	if err != nil {
		return "", fmt.Errorf("error building tunnels: %w", err)
	}
	return sessionName, nil
}

func parseCmdLineArgs() (reloadConfig bool) {
	flag.BoolVar(
		&reloadConfig,
		"reload-config",
		false,
		"whether or not to repopulate the instance list from GCP",
	)
	flag.Parse()

	return reloadConfig
}

func configureLogger() {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func populateGCPResources() []gcptunneler.ProjectData {
	ctx := context.Background()

	projects := gcptunneler.ListProjects(ctx)
	projectDataList := gcptunneler.GetInstancesByProject(ctx, projects)

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

	return projectDataList
}

func run() error {
	configureLogger()

	cfgErr := loadConfiguration()
	if cfgErr != nil {
		return cfgErr
	}

	resourceNames := selectResources()
	sessionName, sessErr := connectToResources(resourceNames)
	if sessErr != nil {
		return sessErr
	}

	switchErr := utils.SwitchToCreatedSession(sessionName)
	if switchErr != nil {
		return fmt.Errorf("couldn't switch to tmux session: %w", switchErr)
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal().Err(err).Msg("application failed")
	}
}
