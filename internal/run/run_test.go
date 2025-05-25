package run

import (
	"errors"
	"gcp-tunneler/internal/config"
	gcptunneler "gcp-tunneler/internal/gcp_api"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

type TestDoubleConfiguration struct {
	CheckIfFileExistsFn    func(string) bool
	PopulateGCPResourcesFn func() []gcptunneler.ProjectData
	WriteFileFn            func(string, []byte, os.FileMode) error
	MarshalIndentFn        func(any, string, string) ([]byte, error)
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

func (t *TestDoubleConfiguration) MarshalIndent(
	v any,
	prefix string,
	indent string,
) ([]byte, error) {
	return t.MarshalIndentFn(v, "", "  ")
}

func TestWriteResourceDetailsToFile(t *testing.T) {
	// Disable all logging
	zerolog.SetGlobalLevel(zerolog.Disabled)

	writeDetailsToFileTests := map[string]struct {
		// got
		reloadCfgFlag     bool
		fileExists        bool
		marshallIndentErr error
		// want
		writeFileFnCalled bool
	}{
		"reload flag false and file exists": {
			reloadCfgFlag:     false,
			fileExists:        true,
			writeFileFnCalled: false,
			marshallIndentErr: nil,
		},
		"reload flag true and file exists": {
			reloadCfgFlag:     true,
			fileExists:        true,
			writeFileFnCalled: true,
			marshallIndentErr: nil,
		},
		"reload flag false and file doesn't exist": {
			reloadCfgFlag:     false,
			fileExists:        false,
			writeFileFnCalled: true,
			marshallIndentErr: nil,
		},
		"reload flag false and file doesn't exist but json marshal error": {
			reloadCfgFlag:     false,
			fileExists:        false,
			writeFileFnCalled: false,
			marshallIndentErr: errors.New("test error"),
		},
	}

	for name, test := range writeDetailsToFileTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			writeFileCalled := false
			envCfg := config.ConfigV2{GCPResourceDetailsFilename: "test.json"}
			mockConfig := &TestDoubleConfiguration{
				CheckIfFileExistsFn: func(string) bool {
					return test.fileExists
				},
				PopulateGCPResourcesFn: func() []gcptunneler.ProjectData {
					return []gcptunneler.ProjectData{{Project: "Poep"}}
				},
				MarshalIndentFn: func(any, string, string) ([]byte, error) {
					return []byte{}, test.marshallIndentErr
				},
				WriteFileFn: func(string, []byte, os.FileMode) error {
					writeFileCalled = true
					return test.marshallIndentErr
				},
			}

			app := &Application{
				Config: mockConfig,
			}
			err := app.writeResourceDetailsToFile(test.reloadCfgFlag, &envCfg)

			if writeFileCalled != test.writeFileFnCalled {
				t.Errorf(
					"test failed for test case %+v, result: writeFileCalled: %v",
					test,
					writeFileCalled,
				)
			}
			if err != test.marshallIndentErr {
				if !strings.Contains(err.Error(), test.marshallIndentErr.Error()) {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// func TestProgramShouldntWriteToFile(t *testing.T) {
// 	writeFileCalled := false
// 	envCfg := config.ConfigV2{GCPResourceDetailsFilename: "test.json"}
// 	reloadCfgFlag := false
// 	mockConfig := &TestDoubleConfiguration{
// 		CheckIfFileExistsFn: func(string) bool {
// 			return true
// 		},
// 		PopulateGCPResourcesFn: func() []gcptunneler.ProjectData {
// 			return []gcptunneler.ProjectData{{Project: "Poep"}}
// 		},
// 		WriteFileFn: func(string, []byte, os.FileMode) error {
// 			writeFileCalled = true
// 			return nil
// 		},
// 	}
//
// 	app := &Application{
// 		Config: mockConfig,
// 	}
//
// 	err := app.writeResourceDetailsToFile(reloadCfgFlag, &envCfg)
//
// 	if writeFileCalled || err != nil {
// 		t.Errorf("loadConfiguration() shouldn't write to file: %v", err)
// 	}
// }
