package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var AppVersion string

// Read from either the dev or prod file.
func Read() *Config {
	eFile := "config/prod.json"
	if os.Getenv("APP_ENV") == "DEV" {
		eFile = "config/dev.json"
	}

	jsonFile, errFile := os.ReadFile(eFile)
	if errFile != nil {
		dirName := os.Getenv("PROJECT_NAME")
		if dirName != "" {
			os.Setenv("APP_ENV", "DEV")
			wd, _ := os.Getwd()
			for !strings.HasSuffix(wd, dirName) {
				wd = filepath.Dir(wd)
			}
			chdirErr := os.Chdir(wd)
			if chdirErr != nil {
				panic(chdirErr.Error())
			}
			jsonFile, errFile = os.ReadFile("config/dev.json")
		}
		if errFile != nil {
			panic("No config file!")
		}
	}

	var config Config
	err := json.Unmarshal(jsonFile, &config)
	if err != nil {
		panic(err.Error())
	}

	config.Version = AppVersion
	return &config
}

type Config struct {
	Logger     *Logger     `json:"logger"`
	Version    string      `json:"version"`
	Secrets    *Secrets    `json:"secrets"`
	Data       *Data       `json:"data"`
	HttpServer *HttpServer `json:"httpserver"`
	Health     *Health     `json:"health"`
	Test       *Test       `json:"test"`
	Metrics    *Metrics    `json:"metrics"`
	Cache      *Cache      `json:"cache"`
}

type Logger struct {
	Level string `json:"level"`
}

type Secrets struct {
	Openbao Openbao `json:"openbao"`
}

type Openbao struct {
	Token  string `json:"token"`
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   string `json:"port"`
}

func (o *Openbao) ReadConfig() string {
	return fmt.Sprintf(
		"%v://%v:%v",
		o.Scheme, o.Host, o.Port,
	)
}

func (o *Openbao) ReadToken() {
	o.Token = os.Getenv("OPENBAO_TOKEN")
}

type Data struct {
	Relational []Rdb `json:"relational"`
}

type Rdb struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	User      string `json:"user"`
	Database  string `json:"database"`
	Sslmode   string `json:"sslmode"`
	Secret    string `json:"secret"`
	SecretKey string `json:"secret_key"`
}

type HttpServer struct {
	Port         string `json:"port"`
	TimeoutRead  int    `json:"timeout_read"`
	TimeoutWrite int    `json:"timeout_write"`
	TimeoutIdle  int    `json:"timeout_idle"`
	Certificate  string `json:"certificate"`
	Key          string `json:"key"`
}

type Health struct {
	PingDbTimer     int    `json:"ping_db_timer"`
	HeapTimer       int    `json:"heap_timer"`
	HeapSize        uint64 `json:"heap_size"`
	RoutTimer       int    `json:"rout_timer"`
	RoutinesPerCore int    `json:"routines_per_core"`
}

type Test struct {
	DbPosition int    `json:"db_position"`
	FakeData   string `json:"fake_data"`
}

type Metrics struct {
	GarbageCollection bool `json:"garbage_collection"`
	Memory            bool `json:"memory"`
	Scheduler         bool `json:"scheduler"`
	Cpu               bool `json:"cpu"`
	Lock              bool `json:"lock"`
	Process           bool `json:"process"`
}

type Cache struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	Db        int    `json:"db"`
	Secret    string `json:"secret"`
	SecretKey string `json:"secret_key"`
}
