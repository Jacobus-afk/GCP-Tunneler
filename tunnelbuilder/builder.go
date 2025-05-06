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

type SSHConnection struct {
	Connected bool
	Port      string
	Username  string
}

func goWorker(jobs chan [2]string, results chan SSHConnection) {
	for {
		j, more := <-jobs
		if !more {
			return
		}
		resp := utils.WaitForSSHSession(j[0], j[1])

		sshConnection := SSHConnection{Connected: resp, Port: j[1], Username: j[0]}

		results <- sshConnection
	}
}

func BuildTunnelAndSSH(resourceNames string) {
	resourceList := strings.Split(resourceNames, "\n")
	numJobs := len(resourceList)
	jobs := make(chan [2]string, numJobs)
	results := make(chan SSHConnection, numJobs)


	instanceList, sessionNames := buildTunnelCommands(resourceList)

	sessionName, err := utils.PromptForSessionName(sessionNames)
	if err != nil {
		log.Fatal().Err(err).Msg("halting build tunnel process")
	}
	log.Debug().Msg(sessionName)

	_ = sessionName
	
	for range numJobs {
		go goWorker(jobs, results)
	}

	createTunnels(instanceList, jobs)

	close(jobs)


}

func createTunnels(instanceList []Instance, jobs chan [2]string) {
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

		jobs <- [2]string{currentUser.Username, strconv.Itoa(freePort)}
	}
}

func buildTunnelCommands(resourceList []string) ([]Instance, map[string]bool) {
	sessionNames := map[string]bool{}

	instanceList := []Instance{}
	// resourceList := strings.Split(resourceNames, "\n")

	// numJobs := len(resourceList)
	// jobs := make(chan [2]string, numJobs)
	// results := make(chan SSHConnection, numJobs)

	for _, entry := range resourceList {
		// log.Debug().Msg(entry)
		instance := getTunnelDetails(entry)
		instanceList = append(instanceList, instance)
		sessionNames[instance.InstanceGroup] = true
	}

	return instanceList, sessionNames

	// sessionName, err := utils.PromptForSessionName(sessionNames)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("halting build tunnel process")
	// }
	// log.Debug().Msg(sessionName)
	//
	// _ = sessionName

	// for range numJobs {
	// 	go goWorker(jobs, results)
	// }

	// for _, instance := range instanceList {
	// 	log.Info().Interface("instance", instance).Msg("")
	//
	// 	freePort, err := utils.GetFreePort()
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("couldn't get free port")
	// 		return
	// 	}
	//
	// 	currentUser, err := user.Current()
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("couldn't get current user")
	// 	}
	//
	// 	gcloudCMD := buildGCloudCommand(instance, freePort)
	//
	// 	utils.CreateTMUXTunnelSession(gcloudCMD, instance.Name)
	//
	// 	jobs <- [2]string{currentUser.Username, strconv.Itoa(freePort)}
	// }
	// close(jobs)

	for range numJobs {
		sshConnection := <-results

		log.Info().Interface("SSH Connection", sshConnection).Msg("")

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
