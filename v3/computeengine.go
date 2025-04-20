package gcptunneler

import (
	"context"
	// "encoding/json"
	// "fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"

	"gcp-tunneler/config"
)

type InstanceData struct {
	Name string `json:"name"`
	Zone string `json:"zone"`
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

		if len(zone.Value.Instances) == 0 {
			continue
		}

		zoneKey := strings.TrimPrefix(zone.Key, "zones/")

		// fmt.Println(zoneKey)

		for _, instance := range zone.Value.Instances {

			// if checkExclusions(instance) {
			// 	continue
			// }

			if !checkInclusions(instance) {
				continue
			}

			// fmt.Println(*instance.Name)
			instanceList = append(instanceList, InstanceData{*instance.Name, zoneKey})
		}

		// fmt.Println("----------------------------------------")
	}
	return instanceList
}

func checkInclusions(instance *computepb.Instance) bool {
	instanceInclusions := config.GetConfig().Inclusions
	instanceName := *instance.Name
	// excludeInstance := false
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
	// excludeInstance := false
	for _, pattern := range instanceExclusions {
		if strings.Contains(instanceName, pattern) {
			return true
		}
	}
	return false
}
