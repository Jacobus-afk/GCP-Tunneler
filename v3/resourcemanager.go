package gcptunneler

import (
	"context"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	resourcemanagerpb "cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/iterator"
)

func ListProjects(ctx context.Context) []string {
	projectList := []string{}

	projectsClient, _ := resourcemanager.NewProjectsClient(ctx)
	defer projectsClient.Close()

	req := &resourcemanagerpb.SearchProjectsRequest{}

	it := projectsClient.SearchProjects(ctx, req)

	for {
		project, err := it.Next()
		if err == iterator.Done {
			break
		}

		projectList = append(projectList, project.ProjectId)
	}
	return projectList
}
