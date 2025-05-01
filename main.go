package main

import (
	"context"
	// "encoding/json"
	"errors"
	"flag"
	"gcp-tunneler/config"
	gcptunneler "gcp-tunneler/v3"
	"os"
	"os/exec"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

		// gcptunneler.MatchInstancesWithHosts(ctx, "")
		gcptunneler.MatchInstancesWithHosts(ctx, "")

		// projects := gcptunneler.ListProjects(ctx)
		// projectDataList := gcptunneler.GetInstancesByProject(ctx, projects)
		// jsonData, err := json.MarshalIndent(projectDataList, "", "  ")
		// if err != nil {
		// 	log.Fatal().Err(err).Msg("Error marshaling to JSON: %v")
		// }
		//
		// os.WriteFile(cfg.InstanceFilename, jsonData, 0644)


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

	selectedProject := runCommand("./scripts/project_select.sh", cfg.InstanceFilename)
	log.Print(selectedProject)

	selectedInstance := runCommand(
		"./scripts/instance_select.sh",
		cfg.InstanceFilename, selectedProject,
	)
	log.Print(selectedInstance)

}

func runCommand(cmdName string, cmdArgs ...string) string {
	cmd := exec.Command(cmdName, cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running command")
	}

	return (string(out))
}
