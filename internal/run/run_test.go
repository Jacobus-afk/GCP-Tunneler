package run

import (
	"gcp-tunneler/internal/config"
	gcptunneler "gcp-tunneler/internal/gcp_api"
	"os"
	"testing"
)

type TestDoubleConfiguration struct {
	CheckIfFileExistsFn    func(string) bool
	PopulateGCPResourcesFn func() []gcptunneler.ProjectData
	WriteFileFn            func(string, []byte, os.FileMode) error
}

// verify test interface is still the same as actual one
var _ Configuration = (*TestDoubleConfiguration)(nil)

func (t *TestDoubleConfiguration) CheckIfFileExists(path string) bool {
	return t.CheckIfFileExistsFn(path)
}

func (t *TestDoubleConfiguration) PopulateGCPResources() []gcptunneler.ProjectData {
	return t.PopulateGCPResourcesFn()
}

func (t *TestDoubleConfiguration) WriteFile(name string, data []byte, perm os.FileMode) error {
	return t.WriteFileFn(name, data, perm)
}

func TestProgramShouldntWriteToFile(t *testing.T) {
	writeFileCalled := false
	envCfg := config.ConfigV2{GCPResourceDetailsFilename: "test.json"}
	reloadCfgFlag := false
	mockConfig := &TestDoubleConfiguration{
		CheckIfFileExistsFn: func(string) bool {
			return true
		},
		PopulateGCPResourcesFn: func() []gcptunneler.ProjectData {
			return []gcptunneler.ProjectData{{Project: "Poep"}}
		},
		WriteFileFn: func(string, []byte, os.FileMode) error {
			writeFileCalled = true
			return nil
		},
	}

	app := &Application{
		Config: mockConfig,
	}

	err := app.writeResourceDetailsToFile(reloadCfgFlag, &envCfg)

	if writeFileCalled || err != nil {
		t.Errorf("loadConfiguration() shouldn't write to file: %v", err)
	}
}
