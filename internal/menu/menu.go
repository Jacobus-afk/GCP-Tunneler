package menu

import (
	"gcp-tunneler/internal/config"
	"gcp-tunneler/internal/utils"

	"strings"
)


type Menu struct{}

func (m *Menu) RunMenu() string {
	var (
		nextMenu        MenuSelection
		responseFZF     string
		selectedProject string
		selectedView    string
	)

	currentMenu := ProjectMenu

	for {
		switch currentMenu {

		case ProjectMenu:
			selectedView = ""
			responseFZF = selectProject()
			selectedProject = responseFZF

		case ViewMenu:
			responseFZF = selectView(selectedProject)
			selectedView = responseFZF

		case ResourcesMenu:
			if selectedView == BackendResource.String() {
				responseFZF = selectBackend(selectedProject)
			} else if selectedView == InstanceResource.String() {
				responseFZF = selectInstance(selectedProject)
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

func selectProject() string {
	configPath := config.GetConfig().GetGCPResourceDetailsPath()
	selectProjectScript := config.GetScriptConfig().SelectProjectScript
	selectedProject := utils.CommandCombinedOutput(selectProjectScript, configPath)

	return selectedProject
}

func selectView(selectedProject string) string {
	configPath := config.GetConfig().GetGCPResourceDetailsPath()
	selectViewScript := config.GetScriptConfig().SelectViewScript
	selectedView := utils.CommandCombinedOutput(
		selectViewScript,
		configPath,
		selectedProject,
	)

	return selectedView
}

func selectBackend(selectedProject string) string {
	configPath := config.GetConfig().GetGCPResourceDetailsPath()
	selectBackendScript := config.GetScriptConfig().SelectBackendScript
	selectedBackend := utils.CommandCombinedOutput(
		selectBackendScript,
		configPath, selectedProject,
	)
	return selectedBackend
}

func selectInstance(selectedProject string) string {
	configPath := config.GetConfig().GetGCPResourceDetailsPath()
	selectInstanceScript := config.GetScriptConfig().SelectInstanceScript
	selectedInstance := utils.CommandCombinedOutput(
		selectInstanceScript,
		configPath, selectedProject,
	)

	return selectedInstance
}
