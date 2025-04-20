package config

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Exclusions []string
	Inclusions []string
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file", err)
		}

		exclusions_env := os.Getenv("GCPT_EXCLUDED_INSTANCES")
		inclusions_env := os.Getenv("GCPT_INCLUDED_INSTANCES")

		inclusions := envSplitter(inclusions_env)
		exclusions := envSplitter(exclusions_env)
		
		instance = &Config{
			Exclusions: exclusions,
			Inclusions: inclusions,
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
