package tunnelbuilder

import (
	"encoding/json"
	"fmt"
	"gcp-tunneler/utils"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type Instance struct {
	Project       string `json:"project"`
	Name          string `json:"name"`
	Zone          string `json:"zone"`
	InstanceGroup string `json:"instance_group"`
}

func BuildTunnelCommands(resourceNames string) {
	sessionNames  := map[string]bool{}

	instanceList := []Instance{}
	resourceList := strings.SplitSeq(resourceNames, "\n")

	for entry := range resourceList {
		// log.Debug().Msg(entry)
		instance := getTunnelDetails(entry)
		instanceList = append(instanceList, instance)
		sessionNames[instance.InstanceGroup] = true
	}

	sessionName, err := utils.PromptForSessionName(sessionNames)
	if err != nil {
		log.Fatal().Err(err).Msg("halting build tunnel process")
	}
	log.Debug().Msg(sessionName)

	_ = sessionName

	for _, instance := range instanceList {
		log.Info().Interface("instance", instance).Msg("")

		gcloudCMD, freePort := buildGCloudCommand(instance)
		_ = freePort

		utils.CreateTMUXTunnelSession(gcloudCMD, instance.Name)
	}
}

func buildGCloudCommand(instance Instance) (gcloudCMD string, freePort int) {
	freePort, err := utils.GetFreePort()
	if err != nil {
		log.Error().Err(err).Msg("couldn't get free port")
		return
	}

	// gcloudCMD = []string{
	// 	"gcloud",
	// 	"compute",
	// 	"start-iap-tunnel",
	// 	instance.Name,
	// 	"22",
	// 	"--local-host-port=localhost:" + strconv.Itoa(freePort),
	// 	"--project=" + instance.Project,
	// 	"--zone=" + instance.Zone,
	// }
	gcloudCMD = fmt.Sprintf(
		"gcloud compute start-iap-tunnel %s 22 --local-host-port=localhost:%s --project=%s --zone=%s",
		instance.Name,
		strconv.Itoa(freePort),
		instance.Project,
		instance.Zone,
	)

	return gcloudCMD, freePort
}

func getTunnelDetails(resourceName string) Instance {
	var instance Instance
	rawJSON := utils.CommandCombinedOutput("./scripts/resource_builder.sh", resourceName)

	err := json.Unmarshal([]byte(rawJSON), &instance)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't get tunnel details")
	}

	return instance
}
