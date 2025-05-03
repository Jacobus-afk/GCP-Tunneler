package menu

import (
	"gcp-tunneler/config"
	"os/exec"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
)

type ResourceType int

const (
	ScriptsDir = "./scripts/"
	ConfigDir  = "./"
)

var (
	ConfigPath = path.Join(ConfigDir, config.GetConfig().InstanceFilename)

	SelectProjectScript  = path.Join(ScriptsDir, "project_select.sh")
	SelectViewScript     = path.Join(ScriptsDir, "view_select.sh")
	SelectBackendScript  = path.Join(ScriptsDir, "backend_select.sh")
	SelectInstanceScript = path.Join(ScriptsDir, "instance_select.sh")
)

const (
	BackendResource ResourceType = iota
	InstanceResource
)

func (r ResourceType) String() string {
	return [...]string{"backends", "instances"}[r]
}

func Menu() {
	selectedProject := selectProject()

	selectedView := selectView(selectedProject)

	if selectedView == BackendResource.String() {
		selectedBackend := selectBackend(selectedProject)
		log.Print(selectedBackend)

	} else if selectedView == InstanceResource.String() {
		selectedInstance := selectInstance(selectedProject)
		log.Print(selectedInstance)
	}
}

func selectProject() string {
	selectedProject := runCommand(SelectProjectScript, ConfigPath)
	log.Print(selectedProject)

	return selectedProject
}

func selectView(selectedProject string) string {
	selectedView := runCommand(
		SelectViewScript,
		ConfigPath,
		selectedProject,
	)
	log.Print(selectedView)

	return selectedView
}

func selectBackend(selectedProject string) string {
	selectedBackend := runCommand(
		SelectBackendScript,
		ConfigPath, selectedProject,
	)
	return selectedBackend
}

func selectInstance(selectedProject string) string {
	selectedInstance := runCommand(
		SelectInstanceScript,
		ConfigPath, selectedProject,
	)

	return selectedInstance
}

func runCommand(cmdName string, cmdArgs ...string) string {
	cmd := exec.Command(cmdName, cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg("Error running command")
	}

	return strings.TrimSpace(string(out))
}
