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

	// for _, project := range projects {
	// 	fmt.Println(project)
	// 	instances := gcptunneler.ListInstances(ctx, project)
	// 	for _, instance := range instances {
	// 		fmt.Println(instance)
	// 	}
	//
	// }

	projectDataList := gcptunneler.GetInstancesByProject(ctx, projects)

	for _, data := range projectDataList {
		log.Println(data)
	}

	log.Println("")
}
