package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
)

const (
	ResourceDetailsFilename = "gcp_resource_details.json"
	ConfigFilename          = "config.toml"
)

type Instances struct {
	Excluded []string `koanf:"excluded"`
	Included []string `koanf:"included"`
}

type SSH struct {
	Timeout int `koanf:"timeout"`
	UserName string `koanf:"username"`
}

type Develop struct {
	Debug     bool   `koanf:"debug"`
	ConfigDir string `koanf:"configdir"`
}

type ConfigV2 struct {
	Instances                  Instances `koanf:"instances"`
	GCPResourceDetailsFilename string    `koanf:"gcp_resource_details_filename"`
	SSH                        SSH       `koanf:"ssh"`
	Develop                    Develop   `koanf:"develop"`
}

func (c *ConfigV2) GetGCPResourceDetailsPath() string {
	// log.Debug().Msgf("GetGCPResourceDetailsPath elements: %s, %s", c.Develop.ConfigDir, c.GCPResourceDetailsFilename)
	return filepath.Join(c.Develop.ConfigDir, c.GCPResourceDetailsFilename)
}

type ScriptConfig struct {
	SelectProjectScript   string
	SelectViewScript      string
	SelectBackendScript   string
	SelectInstanceScript  string
	ResourceBuilderScript string
}

var (
	config       *ConfigV2
	once         sync.Once
	scriptConfig *ScriptConfig
)

func NewUpdatedConfig(configDir string) *ConfigV2 {
	return &ConfigV2{
		Instances:                  Instances{Excluded: []string{}, Included: []string{}},
		GCPResourceDetailsFilename: ResourceDetailsFilename,
		SSH:                        SSH{Timeout: 12},
		Develop:                    Develop{Debug: false, ConfigDir: configDir},
	}
}

func populateScriptConfig(configDir string) {
	scriptsDir := filepath.Join(configDir, "./scripts/")

	scriptConfig = &ScriptConfig{
		SelectProjectScript:   filepath.Join(scriptsDir, "project_select.sh"),
		SelectViewScript:      filepath.Join(scriptsDir, "view_select.sh"),
		SelectBackendScript:   filepath.Join(scriptsDir, "backend_select.sh"),
		SelectInstanceScript:  filepath.Join(scriptsDir, "instance_select.sh"),
		ResourceBuilderScript: filepath.Join(scriptsDir, "resource_builder.sh"),
	}
}

func GetScriptConfig() *ScriptConfig {
	return scriptConfig
}

func GetConfig() *ConfigV2 {
	once.Do(func() {
		configDir := getConfigDir()
		config = NewUpdatedConfig(configDir)

		_ = godotenv.Overload()

		k := koanf.New(".")

		if err := loadConfigFromFile(k, configDir); err != nil {
			log.Warn().Err(err).Msg("issue loading config from file")
		}

		// this overrides TOML values if they exist
		if err := loadConfigFromEnv(k); err != nil {
			log.Warn().Err(err).Msg("issue loading config from environment variables")
		}

		populateScriptConfig(configDir)
		// envErr := k.Load(
		// 	env.ProviderWithValue("GCPT_", ".", func(s string, v string) (string, any) {
		// 		key := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "GCPT_")), "_", ".")
		//
		// 		if strings.Contains(v, ",") {
		// 			return key, strings.Split(v, ",")
		// 		}
		// 		return key, v
		// 	}),
		// 	nil,
		// )
		// if envErr != nil {
		// 	log.Error().Err(envErr).Msg("error loading environment variables")
		// }

		// log.Debug().Msg("\nKeys in Koanf after loading:\n")
		// for _, key := range k.Keys() {
		// 	log.Debug().Msgf("%s: %v\n", key, k.Get(key))
		// }

		if err := k.Unmarshal("", &config); err != nil {
			log.Warn().Err(err).Msg("error unmarshalling config")
		}

		// log.Debug().Interface("config", config).Msg("loaded configuration")
	})
	return config
}

func loadConfigFromEnv(k *koanf.Koanf) error {
	envErr := k.Load(
		env.ProviderWithValue("GCPT_", ".", func(s string, v string) (string, any) {
			key := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "GCPT_")), "_", ".")

			if strings.Contains(v, ",") {
				return key, strings.Split(v, ",")
			}
			return key, v
		}),
		nil,
	)
	if envErr != nil {
		return fmt.Errorf("couldn't load environment variables: %w", envErr)
	}
	return nil
}

func loadConfigFromFile(k *koanf.Koanf, configDir string) error {
	configPath, cfgPathErr := getConfigFilePath(configDir)
	if cfgPathErr != nil {
		return cfgPathErr
	}
	if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
		return fmt.Errorf("couldn't load from path %s: %w", configPath, err)
	}
	return nil
}

func getConfigDir() string {
	configDir := "./"

	log.Debug().Msgf("getConfigDir ProdBuild: %v", ProdBuild)

	if ProdBuild {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Warn().
				Msgf("couldn't get user home directory, falling back to current working directory: %v", err)
			return configDir
		}
		configDir = filepath.Join(homeDir, ".config", "gcp-tunneler")
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Warn().
			Msgf("path %s doesn't exist, falling back to current working directory: %v", configDir, err)
		return "./"
	}
	return configDir
}

func getConfigFilePath(configDir string) (string, error) {
	configPath := filepath.Join(configDir, ConfigFilename)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("path %s doesn't exist: %w", configPath, err)
	}

	return configPath, nil
}
