package gcptunneler

import (
	"context"
	"fmt"
	"path"

	// "fmt"
	"gcp-tunneler/config"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

type ProjectData struct {
	Project   string         `json:"project"`
	Instances map[string]InstanceData `json:"instances"`
	InstanceGroups map[string]InstanceGroupData `json:"instance_groups"`
	Zones map[string]ZoneData `json:"zones"`
}

type InstanceData struct {
	Name string `json:"name"`
	Zone string `json:"zone"`
	InstanceGroup string `json:"instance_group"`
}

type InstanceGroupData struct {
	Name string `json:"name"`
	Zone string `json:"zone"`
	Instances []string `json:"instances"`
}

type ZoneData struct {
	Name string `json:"name"`
	InstanceGroups []string `json:"instance_groups"`
}

type BackendData struct {
	Region string `json:"region"`
	InstanceGroups []InstanceGroupData `json:"instance_groups"`
}

type HostData struct {
	Host           []string       `json:"host"`
	DefaultService string         `json:"default_service"`
}

func GetInstancesByProject(ctx context.Context, projects []string, instMap map[string]InstanceData) []ProjectData {
	projectDataList := []ProjectData{}
	numJobs := len(projects)

	jobs := make(chan string, numJobs)
	results := make(chan ProjectData, numJobs)

	for range 5 {
		go worker(ctx, jobs, results, instMap)
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

func worker(ctx context.Context, jobs <-chan string, results chan<- ProjectData, instMap map[string]InstanceData) {
	for j := range jobs {
		ListInstances(ctx, j, instMap)
		projectData := ProjectData{Project: j, Instances: instMap}
		results <- projectData
	}
}

func MatchInstancesWithHosts(ctx context.Context, projectID string) {
	instGroupMap := map[string]InstanceGroupData{}
	instMap := map[string]InstanceData{}
	zoneMap  := map[string]ZoneData{}
	ListInstances(ctx, projectID, instMap)
	ListZonalInstanceGroups(ctx, projectID, instGroupMap, instMap, zoneMap)

	backendDataMap, _ := buildBackendServiceMap(ctx, projectID, instGroupMap)

	// hostMap := ListURLMapsWithRules(ctx, projectID)
	// fmt.Println(hostMap)
	// _ = hostMap
	// _ = instanceList
	// for _, entry := range instMap {
	// 	log.Info().Interface("InstanceData",entry).Msg("")
	// }
	// for _, entry := range instGroupMap {
	// 	log.Info().Interface("InstanceGroupData",entry).Msg("")
	// }
	// for _, entry := range zoneMap {
	// 	log.Info().Interface("ZoneData",entry).Msg("")
	// }
	for be, entry := range backendDataMap {
		log.Info().Interface("BackendData",entry).Msg(be)
	}
}

func buildBackendServiceMap(
	ctx context.Context,
	projectID string,
	instGroupMap map[string]InstanceGroupData,
	// zoneMap map[string]ZoneData,
) (map[string]BackendData, error) {
	backendDataMap := map[string]BackendData{}

	backendServicesClient, err := compute.NewBackendServicesRESTClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create backend services client")
	}
	defer backendServicesClient.Close()

	// Map to store backend service -> instance groups
	// backendServiceMap := make(BackendServiceMap)

	// List backend services
	req := &computepb.AggregatedListBackendServicesRequest{
		Project: projectID,
	}

	it := backendServicesClient.AggregatedList(ctx, req)
	for {
		backendService, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal().Err(err).Msg("error listing backend services")
		}

		// serviceURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/backendServices/%s", 
		// 	projectID, *backendService.Name)

		// Extract instance group URLs from backends
		// fmt.Println(backendService.Key)
		for _, service := range backendService.Value.BackendServices {
			// fmt.Println("  ", *service.Name)
			instGroupsList := []InstanceGroupData{}
			for _, backend := range service.Backends {
				instanceGroups := ""
				// zones := ""
				splitGroup := strings.Split(*backend.Group, "/")
				for idx, entry := range splitGroup {
					// if entry == "zones" {
					// 	zones = splitGroup[idx + 1]
					// }
					if entry == "instanceGroups" {
						instanceGroups = splitGroup[idx+1]
					}
				}
				if instGroupEntity, exists := instGroupMap[instanceGroups]; exists {
					instGroupsList = append(instGroupsList, instGroupEntity)
					// fmt.Println("GOT INST GROUP: ", backendService.Key)
				} else {
					continue
				}
				// fmt.Println("    zones: ", zones)
				// fmt.Println("    instanceGroups: ", instanceGroups, instGroupsList)
			}
			if len(instGroupsList) != 0 {
				backendDataMap[*service.Name] = BackendData{
					Region:         backendService.Key,
					InstanceGroups: instGroupsList,
				}
			}

			// if backend.Backends != nil {
			// 	// backendServiceMap[serviceURL] = append(backendServiceMap[serviceURL], *backend.Group)
			// 	beGrp := path.Base(*backend.Group)
			// 	fmt.Println("    ", beGrp)
			// }
		}
	}

	return backendDataMap, nil
}

func ListZonalInstanceGroups(
	ctx context.Context,
	projectID string,
	instanceGroupMap map[string]InstanceGroupData,
	instMap map[string]InstanceData,
	zoneMap map[string]ZoneData,
) {

	instanceGroupsClient, err := compute.NewInstanceGroupsRESTClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create instance groups client")
	}
	defer instanceGroupsClient.Close()

	// Use AggregatedList to get all instance groups across zones
	req := &computepb.AggregatedListInstanceGroupsRequest{
		Project: projectID,
	}

	it := instanceGroupsClient.AggregatedList(ctx, req)

	// fmt.Println("Listing zonal instance groups:")
	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal().Err(err).Msg("error listing instance groups")
		}

		// Skip if no instance groups in this scope
		if len(pair.Value.InstanceGroups) == 0 {
			continue
		}

		// Extract zone from scope key
		// fmt.Println(pair.Key)
		scope := pair.Key
		if !strings.HasPrefix(scope, "zones/") {
			// fmt.Println(pair.Value)
			continue // Skip non-zonal scopes
		}
		
		zone := strings.TrimPrefix(scope, "zones/")
		// fmt.Printf("\nZone: %s\n", zone)

		instGroupList := []string{}

		// Process the instance groups in this zone
		for _, group := range pair.Value.InstanceGroups {
			if group.Name == nil {
				continue
			}
			instGroupName := *group.Name
			
			// fmt.Printf("  Instance Group: %s\n", instGroupName)



			




			listReq := &computepb.ListInstancesInstanceGroupsRequest{
				Project:                        projectID,
				Zone:                           zone,
				InstanceGroup:                  *group.Name,
			}

			instanceList := []string{}

			for resp, err := range instanceGroupsClient.ListInstances(ctx, listReq).All() {
			if err != nil {
				fmt.Printf("    Error listing instances: %v\n", err)
				continue
			}
				instance_name := path.Base(*resp.Instance)
				if instEntity, exists := instMap[instance_name]; exists {
					instEntity.InstanceGroup = instGroupName
					instMap[instance_name] = instEntity
				} else {
					continue
				}
				// fmt.Println("    ", instance_name)
				instanceList = append(instanceList, instance_name)
			}

			if len(instanceList) == 0 {
				continue
			}

			// instGroup := InstanceGroupData{Name: instGroupName, Zone: zone, Instances: instanceList}

			instanceGroupMap[instGroupName] = InstanceGroupData{
				Name:      instGroupName,
				Zone:      zone,
				Instances: instanceList,
			}

			instGroupList = append(instGroupList, instGroupName)

		}

		zoneMap[zone] = ZoneData{Name: zone, InstanceGroups: instGroupList}
	}

}


func ListURLMapsWithRules(ctx context.Context, projectID string) map[string]HostData {
	hostMap := map[string]HostData{}

	// Create UrlMaps client
	urlMapsClient, err := compute.NewUrlMapsRESTClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create URL Maps client")
	}
	defer urlMapsClient.Close()


	// List global URL Maps (used by external HTTP(S) load balancers)
	req := &computepb.ListUrlMapsRequest{
		Project: projectID,
	}

	it := urlMapsClient.List(ctx, req)
	for {
		urlMap, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal().Err(err).Msg("error listing URL Maps")
		}

		for _, pathMatcher := range urlMap.PathMatchers {
			if pathMatcher.Name == nil || pathMatcher.DefaultService == nil {
				continue
			}
			defService := path.Base(*pathMatcher.DefaultService)
			hostMap[*pathMatcher.Name] = HostData{DefaultService: defService}
		}

		for _, hostRule := range urlMap.HostRules {
			hData, ok := hostMap[*hostRule.PathMatcher]
			if ok {
				hData.Host = hostRule.Hosts
				hostMap[*hostRule.PathMatcher] = hData
			} else {
				hostMap[*hostRule.PathMatcher] = HostData{Host: hostRule.Hosts}
			}
		}


	}
	log.Info().Msg("___________HOSTS____________")
	fmt.Println("")

	return hostMap
}

func ListInstances(ctx context.Context, projectID string, instMap map[string]InstanceData) {
 	// instanceList := []InstanceData{}
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
			log.Fatal().Err(err).Msg("Error accessing instances in project " + projectID)
			// return instanceList
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

			instName := *instance.Name
			// instanceList = append(instanceList, InstanceData{*instance.Name, zoneKey})
			instMap[instName] = InstanceData{Name: instName, Zone: zoneKey}
		}
	}
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
