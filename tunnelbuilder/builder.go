package tunnelbuilder

import (
	"encoding/json"
	"fmt"
	"gcp-tunneler/utils"
	"os/user"
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
	sessionNames := map[string]bool{}

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

		freePort, err := utils.GetFreePort()
		if err != nil {
			log.Error().Err(err).Msg("couldn't get free port")
			return
		}

		currentUser, err := user.Current()
		if err != nil {
			log.Error().Err(err).Msg("couldn't get current user")
		}

		gcloudCMD := buildGCloudCommand(instance, freePort)

		utils.CreateTMUXTunnelSession(gcloudCMD, instance.Name)

		utils.WaitForSSHSession(currentUser.Username, freePort)

	}
}

func buildGCloudCommand(instance Instance, freePort int) (gcloudCMD string) {
	gcloudCMD = fmt.Sprintf(
		"gcloud compute start-iap-tunnel %s 22 --local-host-port=localhost:%s --project=%s --zone=%s",
		instance.Name,
		strconv.Itoa(freePort),
		instance.Project,
		instance.Zone,
	)

	return gcloudCMD
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
