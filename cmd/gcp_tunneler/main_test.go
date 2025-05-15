package main

import (
	"gcp-tunneler/internal/config"
	gcptunneler "gcp-tunneler/internal/gcp_api"
	"os"
	"testing"
)

type TestDoubleConfiguration struct {
	GetConfigFn            func() *config.Config
	CheckIfFileExistsFn    func(string) bool
	PopulateGCPResourcesFn func() []gcptunneler.ProjectData
	ParseCmdLineArgsFn     func() bool
	WriteFileFn            func(string, []byte, os.FileMode) error
}

var _ Configuration = (*TestDoubleConfiguration)(nil)

func (t *TestDoubleConfiguration) GetConfig() *config.Config {
	return t.GetConfigFn()
}

func (t *TestDoubleConfiguration) CheckIfFileExists(path string) bool {
	return t.CheckIfFileExistsFn(path)
}

func (t *TestDoubleConfiguration) PopulateGCPResources() []gcptunneler.ProjectData {
	return t.PopulateGCPResourcesFn()
}

func (t *TestDoubleConfiguration) ParseCmdLineArgs() bool {
	return t.ParseCmdLineArgsFn()
}

func (t *TestDoubleConfiguration) WriteFile(name string, data []byte, perm os.FileMode) error {
	return t.WriteFileFn(name, data, perm)
}

func TestLoadConfigurationShouldntWriteToFile(t *testing.T) {
	writeFileCalled := false
	mockConfig := &TestDoubleConfiguration{
		GetConfigFn: func() *config.Config {
			return &config.Config{InstanceFilename: "test.json"}
		},
		CheckIfFileExistsFn: func(string) bool {
			return true
		},
		PopulateGCPResourcesFn: func() []gcptunneler.ProjectData {
			return []gcptunneler.ProjectData{{Project: "Poep"}}
		},
		ParseCmdLineArgsFn: func() bool {
			return false
		},
		WriteFileFn: func(string, []byte, os.FileMode) error {
			writeFileCalled = true
			return nil
		},
	}

	app := &Application{
		Config: mockConfig,
	}

	err := app.loadConfiguration()

	if writeFileCalled || err != nil {
		t.Errorf("loadConfiguration() shouldn't write to file: %v", err)
	}
}
