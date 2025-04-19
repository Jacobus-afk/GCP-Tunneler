package main

import (
	"context"
	"fmt"

	gcptunneler "gcp-tunneler/v3"
)

func main() {
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

