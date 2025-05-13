package main

import (
	"context"
	"encoding/json"
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

func run() (string, error) {
	cfg := config.GetConfig()

	configureLogger()

	reloadConfig := parseCmdLineArgs()
	configGCPResourceFileExists := utils.CheckIfFileExists(cfg.InstanceFilename)

	if reloadConfig || !configGCPResourceFileExists {
		projectDataList := populateGCPResources()

		log.Info().
			Str("config_file", cfg.InstanceFilename).
			Msg("writing GCP resource details to file...")

		jsonData, jsonErr := json.MarshalIndent(projectDataList, "", "  ")
		if jsonErr != nil {
			return "error marshaling to JSON", jsonErr
		}

		if writeErr := os.WriteFile(cfg.InstanceFilename, jsonData, 0644); writeErr != nil {
			return "couldn't write GCP resource details to file", writeErr
		}

	}

	resourceNames := menu.HandleFZFMenu()
	sessionName, sessErr := tunnelbuilder.BuildTunnelAndSSH(resourceNames)
	if sessErr != nil {
		return "error building tunnels", sessErr
	}

	switchErr := utils.SwitchToCreatedSession(sessionName)
	if switchErr != nil {
		return "couldn't switch to tmux session", switchErr
	}
	return "", nil
}

func main() {
	msg, err := run()
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
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
