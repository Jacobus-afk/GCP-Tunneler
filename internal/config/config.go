package config

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
)

const ResourceDetailsFilename = "gcp_resource_details.json"

type Config struct {
	Exclusions                []string
	Inclusions                []string
	GCPResourceDetailsFilename          string
	SSHTimeout                int
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Error().Err(err).Msg("Error loading .env file")
		}

		exclusionsEnv := os.Getenv("GCPT_EXCLUDED_INSTANCES")
		inclusionsEnv := os.Getenv("GCPT_INCLUDED_INSTANCES")

		// resourceDetailsFilename := os.Getenv("GCPT_RESOURCE_DETAILS_FILENAME")
		sshTimeout, err := strconv.Atoi(os.Getenv("GCPT_SSH_TIMEOUT"))
		if err != nil {
			log.Error().Err(err).Msg("invalid timeout value. reverting to default")
			sshTimeout = 12
		}

		inclusions := envSplitter(inclusionsEnv)
		exclusions := envSplitter(exclusionsEnv)

		instance = &Config{
			Exclusions:                 exclusions,
			Inclusions:                 inclusions,
			GCPResourceDetailsFilename: ResourceDetailsFilename,
			SSHTimeout:                 sshTimeout,
		}
	})
	return instance
}

func envSplitter(env_string string) []string {
	if env_string == "" {
		return []string{}
	}

	string_list := strings.Split(env_string, ",")
	for i := range string_list {
		string_list[i] = strings.TrimSpace(string_list[i])
	}
	return string_list
}
