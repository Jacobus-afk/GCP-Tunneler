package run

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

type Configuration interface {
	GetConfig() *config.Config
	CheckIfFileExists(path string) bool
	PopulateGCPResources() []gcptunneler.ProjectData
	ParseCmdLineArgs() bool
	WriteFile(name string, data []byte, perm os.FileMode) error
}

// type DefaultDependencies struct {
// 	configuration Configuration
// }
//
// func (d *DefaultDependencies) GetConfig() *config.Config {
// 	return d.configuration.GetConfig()
// }
//
// func (d *DefaultDependencies) CheckIfFileExists(path string) bool {
// 	return d.configuration.CheckIfFileExists(path)
// }
//
// func (d *DefaultDependencies) PopulateGCPResources() []gcptunneler.ProjectData {
// 	return d.configuration.PopulateGCPResources()
// }
//
// func (d *DefaultDependencies) ParseCmdLineArgs() bool {
// 	return d.configuration.ParseCmdLineArgs()
// }
//
// func (d *DefaultDependencies) WriteFile(name string, data []byte, perm os.FileMode) error {
// 	return d.configuration.WriteFile(name, data, perm)
// }

type RealConfiguration struct{}

func (r *RealConfiguration) GetConfig() *config.Config {
	return config.GetConfig()
}

func (r *RealConfiguration) CheckIfFileExists(path string) bool {
	return utils.CheckIfFileExists(path)
}

func (r *RealConfiguration) PopulateGCPResources() []gcptunneler.ProjectData {
	return populateGCPResources()
}

func (r *RealConfiguration) ParseCmdLineArgs() bool {
	return parseCmdLineArgs()
}

func (r *RealConfiguration) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

type Application struct {
	Config Configuration
}

// func LoadConfiguration(defaultDeps *DefaultDependencies) error {
func (app *Application) WriteResourceDetailsToFile(reloadCfgFlag bool, envCfg *config.Config) error {
	configGCPResourceFileExists := app.Config.CheckIfFileExists(envCfg.GCPResourceDetailsFilename)

	reloadConfig := app.Config.ParseCmdLineArgs()

	if reloadConfig || !configGCPResourceFileExists {
		projectDataList := app.Config.PopulateGCPResources()

		log.Info().
			Str("config_file", envCfg.GCPResourceDetailsFilename).
			Msg("writing GCP resource details to file...")

		jsonData, jsonErr := json.MarshalIndent(projectDataList, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("error marshaling to JSON: %w", jsonErr)
		}

		if writeErr := app.Config.WriteFile(envCfg.GCPResourceDetailsFilename, jsonData, 0644); writeErr != nil {
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

func (app *Application) Run() error {
	configureLogger()

	// realConfig := &RealConfiguration{}
	//
	// defaultDeps := &DefaultDependencies{configuration: realConfig}

	cfgErr := app.LoadConfiguration()
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
