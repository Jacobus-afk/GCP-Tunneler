package menu

import (
	"gcp-tunneler/internal/config"
	"gcp-tunneler/internal/utils"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

const (
	ScriptsDir = "./internal/scripts/"
)

var (
	ConfigPath = config.GetConfig().InstanceFilename

	SelectProjectScript  = path.Join(ScriptsDir, "project_select.sh")
	SelectViewScript     = path.Join(ScriptsDir, "view_select.sh")
	SelectBackendScript  = path.Join(ScriptsDir, "backend_select.sh")
	SelectInstanceScript = path.Join(ScriptsDir, "instance_select.sh")
)

type MenuSelection int

const (
	ProjectMenu MenuSelection = iota
	ViewMenu
	ResourcesMenu
	menuCount
)

func (m MenuSelection) Previous() MenuSelection {
	currentSel := int(m)

	if currentSel-1 < 0 {
		return MenuSelection(currentSel)
	}

	return MenuSelection(currentSel - 1)
}

func (m MenuSelection) Next() MenuSelection {
	currentSel := int(m)

	if (currentSel + 1) >= int(menuCount) {
		return MenuSelection(currentSel)
	}

	return MenuSelection(currentSel + 1)
}

type ResourceType int

const (
	BackendResource ResourceType = iota
	InstanceResource
)

func (r ResourceType) String() string {
	return [...]string{"backends", "instances"}[r]
}

func HandleFZFMenu() (string){
	var (
		nextMenu         MenuSelection
		responseFZF      string
		selectedProject  string
		selectedView     string
		selectedInstance string
		selectedBackend  string
	)

	currentMenu := ProjectMenu

	for {
		switch currentMenu {

		case ProjectMenu:
			selectedView = ""
			selectedInstance = ""
			selectedBackend = ""
			responseFZF = selectProject()
			selectedProject = responseFZF

		case ViewMenu:
			responseFZF = selectView(selectedProject)
			selectedView = responseFZF

		case ResourcesMenu:
			if selectedView == BackendResource.String() {
				responseFZF = selectBackend(selectedProject)
				selectedBackend = responseFZF
				_ = selectedBackend
			} else if selectedView == InstanceResource.String() {
				responseFZF = selectInstance(selectedProject)
				selectedInstance = responseFZF
				_ = selectedInstance
			}

		}
		// log.Debug().Msg(responseFZF)

		if strings.Contains(responseFZF, "**GO_BACK**") {
			currentMenu = currentMenu.Previous()
			continue
		}

		nextMenu = currentMenu.Next()

		// we've reached the end of menus
		if nextMenu == currentMenu {
			break
		}

		currentMenu = nextMenu

	}

	return responseFZF
}

func Menu() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)

	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		log.Printf("Received interrupt %s\n", sig)
		done <- true
	}()

	HandleFZFMenu()

	select {
	case <-done:
		log.Info().Msg("Exiting..")
		return
	default:
	}
}

func selectProject() string {
	selectedProject := utils.CommandCombinedOutput(SelectProjectScript, ConfigPath)
	// log.Print(selectedProject)

	return selectedProject
}

func selectView(selectedProject string) string {
	selectedView := utils.CommandCombinedOutput(
		SelectViewScript,
		ConfigPath,
		selectedProject,
	)
	// log.Print(selectedView)

	return selectedView
}

func selectBackend(selectedProject string) string {
	selectedBackend := utils.CommandCombinedOutput(
		SelectBackendScript,
		ConfigPath, selectedProject,
	)
	return selectedBackend
}

func selectInstance(selectedProject string) string {
	selectedInstance := utils.CommandCombinedOutput(
		SelectInstanceScript,
		ConfigPath, selectedProject,
	)

	return selectedInstance
}

