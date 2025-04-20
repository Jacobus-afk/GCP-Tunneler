package main

import (
	"context"
	"log"

	"gcp-tunneler/config"
	gcptunneler "gcp-tunneler/v3"
)

func main() {
	config.GetConfig()

	ctx := context.Background()

	projects := gcptunneler.ListProjects(ctx)
	//
	// fmt.Println(projects)

	instances := gcptunneler.ListInstances(ctx, projects[0])
	for _, instance := range instances {
		fmt.Println(instance)
	}
	fmt.Println("")
}

