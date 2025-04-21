package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"gcp-tunneler/config"
	gcptunneler "gcp-tunneler/v3"
	"os"
	"os/exec"

	// "github.com/rs/zerolog"
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
			log.Fatal().Err(err).Msg("Error marshaling to JSON: %v")
		}

		os.WriteFile(cfg.InstanceFilename, jsonData, 0644)

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

	selectedProject := runCommand("./project_select.sh", cfg.InstanceFilename)
	log.Print(selectedProject)

	selectedInstance := runCommand(
		"./instance_select.sh",
		cfg.InstanceFilename, selectedProject,
	)
	log.Print(selectedInstance)
	// -------------------------------------------------------------------

	// inputChan := make(chan string)
	// go func() {
	// 	for _, p := range projectDataList {
	// 		inputChan <- p.Project
	// 	}
	// 	close(inputChan)
	// }()
	//
	// outputChan := make(chan string)
	// go func() {
	// 	for s := range outputChan {
	// 		log.Println("Got: ", s)
	// 	}
	// }()
	//
	// exit := func(code int, err error) {
	// 	if err != nil {
	// 		log.Println(err.Error())
	// 	}
	// 	os.Exit(code)
	// }
	//
	// options, err := fzf.ParseOptions(
	// 	true,
	// 	[]string{"--multi", "--reverse", "--border", "--height=40%"},
	// )
	// if err != nil {
	// 	exit(fzf.ExitError, err)
	// }
	//
	// options.Input = inputChan
	// options.Output = outputChan
	//
	// code, err := fzf.Run(options)
	// exit(code, err)
}

func runCommand(cmdName string, cmdArgs ...string) string {
	cmd := exec.Command(cmdName, cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running command")
	}

	return (string(out))
}
