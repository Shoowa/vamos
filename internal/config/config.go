package config

import (
	"fmt"
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
	Logger     *Logger     `yaml:"logger"`
	Version    string      `yaml:"version"`
	Secrets    *Secrets    `yaml:"secrets"`
	Data       *Data       `yaml:"data"`
	HttpServer *HttpServer `yaml:"httpserver"`
}

type Logger struct {
	Level string `yaml:"level"`
}

type Secrets struct {
	Openbao Openbao `yaml:"openbao"`
}

type Openbao struct {
	Token  string `yaml:"token"`
	Scheme string `yaml:"scheme"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
}

func (o *Openbao) ReadConfig() string {
	o.Token = os.Getenv("OPENBAO_TOKEN")
	return fmt.Sprintf(
		"%v://%v:%v",
		o.Scheme, o.Host, o.Port,
	)
}

type Data struct {
	Relational []Rdb `yaml:"relational"`
}

type Rdb struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Database string `yaml:"database"`
	Sslmode  string `yaml:"sslmode"`
	Secret   string `yaml:"secret"`
}

type HttpServer struct {
	Port         string `yaml:"port"`
	TimeoutRead  int    `yaml:"timeout_read"`
	TimeoutWrite int    `yaml:"timeout_write"`
	TimeoutIdle  int    `yaml:"timeout_idle"`
}
