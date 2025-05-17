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
}

type ConfigV2 struct {
	Instances                  Instances `koanf:"instances"`
	GCPResourceDetailsFilename string    `koanf:"gcp_resource_details_filename"`
	SSH                        SSH       `koanf:"ssh"`
}

var (
	config *ConfigV2
	once   sync.Once
)

func NewUpdatedConfig() *ConfigV2 {
	return &ConfigV2{
		Instances:                  Instances{Excluded: []string{}, Included: []string{}},
		GCPResourceDetailsFilename: ResourceDetailsFilename,
		SSH:                        SSH{Timeout: 12},
	}
}

func GetConfig() *ConfigV2 {
	once.Do(func() {
		config = NewUpdatedConfig()

		_ = godotenv.Overload()

		k := koanf.New(".")

		if err := loadConfigFromFile(k); err != nil {
			log.Warn().Err(err).Msg("issue loading config from file")
		}

		// this overrides TOML values if they exist
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
			log.Error().Err(envErr).Msg("error loading environment variables")
		}

		// fmt.Println("\nKeys in Koanf after loading:")
		// for _, key := range k.Keys() {
		// 	fmt.Printf("%s: %v\n", key, k.Get(key))
		// }

		if err := k.Unmarshal("", &config); err != nil {
			log.Error().Err(err).Msg("error unmarshalling config")
		}

		// log.Debug().Interface("config", config).Msg("")
	})
	return config
}

func loadConfigFromFile(k *koanf.Koanf) error {
	configPath, cfgPathErr := getConfigFilePath()
	if cfgPathErr != nil {
		return cfgPathErr
	}
	if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
		return fmt.Errorf("couldn't load from path %s: %w", configPath, err)
	}
	return nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("couldn't get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "gcp-tunneler")
	configPath := filepath.Join(configDir, ConfigFilename)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("path %s doesn't exist: %w", configPath, err)
	}

	return configPath, nil
}
