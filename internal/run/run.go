package run

import (
	"context"
	"encoding/json"
	"fmt"
	"gcp-tunneler/internal/config"
	"gcp-tunneler/internal/menu"
	"gcp-tunneler/internal/tunnelbuilder"
	"gcp-tunneler/internal/utils"
	"os"

	gcptunneler "gcp-tunneler/internal/gcp_api"

	"github.com/rs/zerolog/log"
)

type Configuration interface {
	CheckIfFileExists(path string) bool
	PopulateGCPResources() []gcptunneler.ProjectData
	WriteFile(name string, data []byte, perm os.FileMode) error
}

type RealConfiguration struct{}

func (r *RealConfiguration) CheckIfFileExists(path string) bool {
	return utils.CheckIfFileExists(path)
}

func (r *RealConfiguration) PopulateGCPResources() []gcptunneler.ProjectData {
	return populateGCPResources()
}

func (r *RealConfiguration) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

type Application struct {
	Config Configuration
}

func (app *Application) WriteResourceDetailsToFile(reloadCfgFlag bool, envCfg *config.ConfigV2) error {
	gcpResourceDetailsPath := envCfg.GetGCPResourceDetailsPath()
	configGCPResourceFileExists := app.Config.CheckIfFileExists(gcpResourceDetailsPath)

	if reloadCfgFlag || !configGCPResourceFileExists {
		projectDataList := app.Config.PopulateGCPResources()

		log.Info().
			Str("config_file", gcpResourceDetailsPath).
			Msg("writing GCP resource details to file...")

		jsonData, jsonErr := json.MarshalIndent(projectDataList, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("error marshaling to JSON: %w", jsonErr)
		}

		if writeErr := app.Config.WriteFile(gcpResourceDetailsPath, jsonData, 0644); writeErr != nil {
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

func (app *Application) Run(reloadCfgFlag bool, envCfg *config.ConfigV2) error {
	cfgErr := app.WriteResourceDetailsToFile(reloadCfgFlag, envCfg)
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
