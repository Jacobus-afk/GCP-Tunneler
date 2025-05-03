package tunnelbuilder

import (
	"encoding/json"
	"gcp-tunneler/utils"
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
	instanceList := []Instance{}
	resourceList := strings.SplitSeq(resourceNames, "\n")

	for entry := range resourceList {
		// log.Debug().Msg(entry)
		instance := getTunnelDetails(entry)
		instanceList = append(instanceList, instance)
	}

	for _, entry := range instanceList {
		log.Info().Interface("instance", entry).Msg("")
	}
}

func getTunnelDetails(resourceName string) Instance {
	var instance Instance
	rawJSON := utils.RunCommand("./scripts/resource_builder.sh", resourceName)

	err := json.Unmarshal([]byte(rawJSON),&instance)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't get tunnel details")
	}

	return instance
}
