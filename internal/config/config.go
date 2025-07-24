package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

var AppVersion string

// Read from either the dev or prod file.
func Read() *Config {
	eFile := "internal/config/prod.yml"
	if os.Getenv("APP_ENV") == "DEV" {
		eFile = "internal/config/dev.yml"
	}

	ymlFile, errFile := os.ReadFile(eFile)
	if errFile != nil {
		panic("No config file!")
	}

	var config Config
	err := yaml.Unmarshal(ymlFile, &config)
	if err != nil {
		panic(err.Error())
	}

	config.Version = AppVersion
	return &config
}

type Config struct {
	Logger  *Logger `yaml:"logger"`
	Version string  `yaml:"version"`
}

type Logger struct {
	Level string `yaml:"level"`
}
