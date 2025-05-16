package gcptunneler

import (
	"context"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	resourcemanagerpb "cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

func ListProjects(ctx context.Context) []string {
	log.Info().Msg("Getting list of GCP projects...")
	projectList := []string{}

	projectsClient, _ := resourcemanager.NewProjectsClient(ctx)
	defer func() { _ = projectsClient.Close()}()

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
