package gcptunneler

import (
	"context"
	"gcp-tunneler/config"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

type ProjectData struct {
	Project   string         `json:"project"`
	Instances []InstanceData `json:"instances"`
}

type InstanceData struct {
	Name string `json:"name"`
	Zone string `json:"zone"`
}

func GetInstancesByProject(ctx context.Context, projects []string) []ProjectData {
	projectDataList := []ProjectData{}
	numJobs := len(projects)

	jobs := make(chan string, numJobs)
	results := make(chan ProjectData, numJobs)

	for range 5 {
		go worker(ctx, jobs, results)
	}

	for _, project := range projects {
		jobs <- project
	}
	close(jobs)

	for range numJobs {
		projectData := <-results

		projectDataList = append(projectDataList, projectData)
	}
	return projectDataList
}

func worker(ctx context.Context, jobs <-chan string, results chan<- ProjectData) {
	for j := range jobs {
		instanceList := ListInstances(ctx, j)
		projectData := ProjectData{Project: j, Instances: instanceList}
		results <- projectData
	}
}

func ListInstances(ctx context.Context, projectID string) []InstanceData {
	instanceList := []InstanceData{}
	instancesClient, _ := compute.NewInstancesRESTClient(ctx)
	defer instancesClient.Close()

	filterStr := "status = RUNNING"
	req := &computepb.AggregatedListInstancesRequest{
		Project: projectID,
		Filter:  &filterStr,
	}

	it := instancesClient.AggregatedList(ctx, req)

	for {
		zone, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Log and return the partial list we have so far
			log.Error().Err(err).Msg("Error accessing instances in project " + projectID)
			return instanceList
		}

		if len(zone.Value.Instances) == 0 {
			continue
		}

		zoneKey := strings.TrimPrefix(zone.Key, "zones/")

		for _, instance := range zone.Value.Instances {

			// if checkExclusions(instance) {
			// 	continue
			// }

			if !checkInclusions(instance) {
				continue
			}

			instanceList = append(instanceList, InstanceData{*instance.Name, zoneKey})
		}
	}
	return instanceList
}

func checkInclusions(instance *computepb.Instance) bool {
	instanceInclusions := config.GetConfig().Inclusions
	instanceName := *instance.Name
	for _, pattern := range instanceInclusions {
		if strings.Contains(instanceName, pattern) {
			return true
		}
	}
	return false
}

func checkExclusions(instance *computepb.Instance) bool {
	instanceExclusions := config.GetConfig().Exclusions
	instanceName := *instance.Name
	for _, pattern := range instanceExclusions {
		if strings.Contains(instanceName, pattern) {
			return true
		}
	}
	return false
}
