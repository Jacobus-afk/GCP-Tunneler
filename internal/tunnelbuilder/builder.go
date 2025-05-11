package tunnelbuilder

import (
	"encoding/json"
	"fmt"
	"gcp-tunneler/internal/utils"
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

func goWorker(jobs chan [2]string, results chan utils.SSHConnection) {
	for {
		j, more := <-jobs
		if !more {
			return
		}
		resp := utils.WaitForSSHSession(j[0], j[1])

		sshConnection := utils.SSHConnection{Connected: resp, Port: j[1], Username: j[0]}

		results <- sshConnection
	}
}

func BuildTunnelAndSSH(resourcesInput string) (string, error) {
	resourceList := strings.Split(resourcesInput, "\n")

	instanceList, possibleSessionNames := buildTunnelCommands(resourceList)

	sessionName, err := utils.PromptForSessionName(possibleSessionNames)
	if err != nil {
		return "", fmt.Errorf("failed to get session name: %w", err)
	}

	connections := createConcurrentTunnelConnections(instanceList)

	if err := setupTMUXEnvironment(connections, sessionName); err != nil {
		return "", fmt.Errorf("failed to setup TMUX Environment: %w", err)
	}

	return sessionName, nil
}

func createConcurrentTunnelConnections(instanceList []Instance) []utils.SSHConnection {
	numJobs := len(instanceList)
	jobs := make(chan [2]string, numJobs)
	results := make(chan utils.SSHConnection, numJobs)

	for range numJobs {
		go goWorker(jobs, results)
	}

	createTunnels(instanceList, jobs)

	close(jobs)

	connections := make([]utils.SSHConnection, 0, numJobs)
	for range numJobs {
		sshConnection := <-results
		connections = append(connections, sshConnection)
	}

	return connections
}

func setupTMUXEnvironment(
	connections []utils.SSHConnection,
	sessionName string,
) error {
	for _, sshConnection := range connections {
		// sshConnection := <-connections
		log.Info().Interface("SSH Connection", sshConnection).Msg("")
		err := utils.CreateTMUXSSHSession(sshConnection, sessionName)
		if err != nil {
			return fmt.Errorf("failed to create TMUX SSH session: %w", err)
		}
	}
	utils.ArrangeLayout(sessionName)

	return nil
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

	for _, entry := range resourceList {
		// log.Debug().Msg(entry)
		instance := getTunnelDetails(entry)
		instanceList = append(instanceList, instance)
		sessionNames[instance.InstanceGroup] = true
	}

	return instanceList, sessionNames
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
