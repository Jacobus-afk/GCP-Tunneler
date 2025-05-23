package tunnelbuilder

import (
	"encoding/json"
	"fmt"
	"gcp-tunneler/internal/config"
	"gcp-tunneler/internal/utils"
	"os/user"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type Builder struct{}

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

func (b *Builder) BuildTunnelAndSSH(resourcesInput string) (string, error) {
	resourceList := strings.Split(resourcesInput, "\n")

	instanceList, possibleSessionNames := buildTunnelCommands(resourceList)

	sessionName, err := utils.PromptForSessionName(possibleSessionNames)
	if err != nil {
		return "", fmt.Errorf("failed to get session name: %w", err)
	}

	connections := b.createConcurrentTunnelConnections(instanceList)

	if err := setupTMUXEnvironment(connections, sessionName); err != nil {
		return "", fmt.Errorf("failed to setup TMUX Environment: %w", err)
	}

	return sessionName, nil
}

func (b *Builder) createConcurrentTunnelConnections(instanceList []Instance) []utils.SSHConnection {
	numJobs := len(instanceList)
	jobs := make(chan [2]string, numJobs)
	results := make(chan utils.SSHConnection, numJobs)

	for range numJobs {
		go goWorker(jobs, results)
	}

	b.createTunnels(instanceList, jobs)

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

func (b *Builder) createTunnels(instanceList []Instance, jobs chan [2]string) {
	for _, instance := range instanceList {
		log.Info().Interface("instance", instance).Msg("")

		freePort, err := utils.GetFreePort()
		if err != nil {
			log.Error().Err(err).Msg("couldn't get free port")
			return
		}

		currentUsername := b.getCurrentUserName()

		gcloudCMD := buildGCloudCommand(instance, freePort)

		utils.CreateTMUXTunnelSession(gcloudCMD, instance.Name)

		jobs <- [2]string{currentUsername, strconv.Itoa(freePort)}
	}
}

func (b *Builder) getCurrentUserName() string {
	currentUserName := config.GetConfig().SSH.UserName
	if currentUserName == "" {
		currentUser, err := user.Current()
		if err != nil {
			log.Error().Err(err).Msg("couldn't get current user")
		} else {
			currentUserName = currentUser.Username
		}
	}
	return currentUserName
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
	configPath := config.GetConfig().GetGCPResourceDetailsPath()
	resourceBuilderScript := config.GetScriptConfig().ResourceBuilderScript
	rawJSON := utils.CommandCombinedOutput(
		resourceBuilderScript,
		configPath,
		resourceName,
	)

	err := json.Unmarshal([]byte(rawJSON), &instance)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't get tunnel details")
	}

	return instance
}
