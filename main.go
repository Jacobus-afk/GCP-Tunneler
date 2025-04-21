package main

import (
	"context"
	"encoding/json"
	"gcp-tunneler/config"
	"log"
	"os"

	gcptunneler "gcp-tunneler/v3"

	fzf "github.com/junegunn/fzf/src"
)

func main() {
	config.GetConfig()

	ctx := context.Background()

	projects := gcptunneler.ListProjects(ctx)
	//
	// fmt.Println(projects)

	// for _, project := range projects {
	// 	fmt.Println(project)
	// 	instances := gcptunneler.ListInstances(ctx, project)
	// 	for _, instance := range instances {
	// 		fmt.Println(instance)
	// 	}
	//
	// }

	projectDataList := gcptunneler.GetInstancesByProject(ctx, projects)

	// for _, data := range projectDataList {
	// 	log.Println(data)
	// }

	jsonData, err := json.MarshalIndent(projectDataList, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}

	// log.Println(string(jsonData))

	os.WriteFile("instances.json", jsonData, 0644)

	log.Println("")

	// -------------------------------------------------------------------

	inputChan := make(chan string)
	go func() {
		for _, p := range projectDataList {
			inputChan <- p.Project
		}
		close(inputChan)
	}()

	outputChan := make(chan string)
	go func() {
		for s := range outputChan {
			log.Println("Got: ", s)
		}
	}()

	exit := func(code int, err error) {
		if err != nil {
			log.Println(err.Error())
		}
		os.Exit(code)
	}

	options, err := fzf.ParseOptions(
		true,
		[]string{"--multi", "--reverse", "--border", "--height=40%"},
	)
	if err != nil {
		exit(fzf.ExitError, err)
	}

	options.Input = inputChan
	options.Output = outputChan

	code, err := fzf.Run(options)
	exit(code, err)
}
