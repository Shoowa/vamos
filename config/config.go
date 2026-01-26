package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AppVersion can be defined during the build command. It will appear in logs.
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

// Config holds various smaller structs needed to define desired behavior in the
// application and various dependencies.
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

// Logger expects debug to be enabled or disabled.
type Logger struct {
	// Value can be TRUE or FALSE. And the logger will use either DEBUG or WARN.
	Debug bool `json:"debug"`
}

// Secrets is currently oriented toward Openbao.
type Secrets struct {
	// An Openbao struct containing a TlsSecret struct.
	Openbao Openbao `json:"openbao"`
}

// Openbao struct expects an OPENBAO_TOKEN value, a site, and TLS config.
type Openbao struct {
	// Token can obtain a value by invoking the ReadToken method.
	Token  string `json:"token"`
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   string `json:"port"`
	// TlsSecret expects local file paths when used inside the Openbao struct.
	TlsClient *TlsSecret `json:"tls_client"`
}

// ReadConfig is the rare method on a struct in this configuration file. It
// simply assembles a string needed by the client to contact a server.
func (o *Openbao) ReadConfig() string {
	return fmt.Sprintf(
		"%v://%v:%v",
		o.Scheme, o.Host, o.Port,
	)
}

// ReadToken is the rare method on a struct in this configuration file. It reads
// an environ variable named OPENBAO_TOKEN.
func (o *Openbao) ReadToken() {
	o.Token = os.Getenv("OPENBAO_TOKEN")
}

// Data holds a list of Rdb structs.
type Data struct {
	Relational []Rdb `json:"relational"`
}

// Rdb represents a Postgres connection.
type Rdb struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
	// Database represents one of the many "databases" residing in a Postgres
	// server. Their jargon is unintuitive.
	Database string `json:"database"`
	// Sslmode toggles TLS connections.
	Sslmode bool `json:"sslmode"`
	// Secret is an Openbao HTTP endpoint.
	Secret string `json:"secret"`
	// SecretKey is an Openbao JSON data field returned from the endpoint.
	SecretKey string `json:"secret_key"`
}

// TlsSecret can be used by the application either as a server or a client for
// mutual TLS inside the same network. The fields often represent HTTP endpoints
// on an Openbao server, and the subsequent JSON key in the returned data. The
// Openbao client relies on the paths in these fields to read CAs, certs, & keys
// hidden in an Openbao server.
//
// This struct is also used to represent local file paths when configuring an
// Openbao client. I overloaded this struct. It can represent either an Openbao
// HTTP endpoint or a local file path.
//
// First, an Openbao client is configured with a local CA, cert, & key. Second,
// the application uses the Openbao client to contact the Openbao server. The
// server stores other CAs, certs, & keys that can be read with the SkeletonKey.
// Third, other clients such as the Redis client and the Postgres client use
// those CAs, certs, & keys hidden in Openbao. Then those clients open secure
// connections to those databases.
type TlsSecret struct {
	// CaPath is where to find an intermediate Certificate Authority.
	CaPath string `json:"ca_path"`
	// CertPath is where to find a x509 certificate.
	CertPath string `json:"cert_path"`
	// CertField is where to find a x509 certificate when reading from an
	// Openbao server.
	CertField string `json:"cert_field"`
	// KeyPath is where to find a x509 key.
	KeyPath string `json:"key_path"`
	// KeyField is where to find a x509 key when reading from an Openbao server.
	KeyField string `json:"key_field"`
}

// RateLimiter configures a token bucket rate limiter for all routes. Each token
// represents a single HTTP request. Every second, the "average" amount of
// tokens is added to the bucket. And the router is allowed to spend the "burst"
// amount per second.
type RateLimiter struct {
	// Active toggles the global rate limiter on and off.
	Active bool `json:"active"`
	// Avergae is the amount of tokens refilled per second.
	Average float64 `json:"average"`
	// Burst is the maximum amount of tokens spent per second.
	Burst int `json:"burst"`
}

// HttpServer expects a CA, x509 cert, & key as a server. And a x509 cert & key
// as a client for an internal network.
type HttpServer struct {
	Port string `json:"port"`
	// TimeoutRead is the amount of seconds allowed to read an entire request,
	// including the body.
	TimeoutRead int `json:"timeout_read"`
	// TimeoutWrite is the amount of seconds allowed to write an entire
	// response, beginning from a TLS handshake to Time To First Byte.
	TimeoutWrite int `json:"timeout_write"`
	// TimeoutIdle is the amount of seconds to hold an idle connection.
	TimeoutIdle int `json:"timeout_idle"`
	// SecretCA is an HTTP endpoint on an Openbao server holding an intermediate
	// CA.
	SecretCA string `json:"secret_ca"`
	// SecretCAKey is a JSON key in Openbao data holding the intermediate CA.
	SecretCAKey string `json:"secret_ca_key"`
	// TlsServer is configuration for the application to become a secure server.
	TlsServer *TlsSecret `json:"tls_server"`
	// TlsClient is configuration for the application to become a secure client.
	TlsClient *TlsSecret `json:"tls_client"`
	// GlobalRateLimiter is an optional rate limiter.
	GlobalRateLimiter *RateLimiter `json:"global_rate_limiter"`
}

// Health configures the thresholds for various healthchecks.
type Health struct {
	// PingDbTimer is the amount of seconds between each Postgres ping.
	PingDbTimer int `json:"ping_db_timer"`
	// HeapTimer is the amount of seconds between each evaluation of heap size.
	HeapTimer int `json:"heap_timer"`
	// HeapSize is the desired maximum amount of megabytes.
	HeapSize uint64 `json:"heap_size"`
	// RoutTimer is the amount of seconds between each measurement of the amount
	// of goroutines.
	RoutTimer int `json:"rout_timer"`
	// RoutinesPerCore is the desired maximum amount of goroutines that can be
	// assigned to each processor. In reality, no code pins a fixed amount of
	// goroutines to each processor. This field simply encourages a developer to
	// consider a workload with this perspective.
	RoutinesPerCore int `json:"routines_per_core"`
}

// Test points the test executables to the test data.
type Test struct {
	// DbPosition is an index into the Data.Relational array.
	DbPosition int `json:"db_position"`
	// FakeData is a local file path to identify .sql scripts.
	FakeData string `json:"fake_data"`
}

// Metrics enables or disables various types of runtime metrics.
type Metrics struct {
	// GarbageCollection toggles the Prometheus metrics that gather GC data.
	GarbageCollection bool `json:"garbage_collection"`
	// Memory toggles the Prometheus metrics that gather memory data.
	Memory bool `json:"memory"`
	// Schedular toggles the Prometheus metrics that gather schedular data.
	Scheduler bool `json:"scheduler"`
	// Cpu toggles the metrics that gather runtime data matching
	// "^/cpu/classes/.*"
	Cpu bool `json:"cpu"`
	// Lock toggles the metrics that gather runtime data matching
	// "^/sync/mutex/.*"
	Lock bool `json:"lock"`
	// Process toggles the Prometheus metrics in
	// collectors.NewProcessCollector(opt)
	Process bool `json:"process"`
}

// Cache currently represents a Redis server.
type Cache struct {
	Host    string `json:"host"`
	Port    string `json:"port"`
	Db      int    `json:"db"`
	User    string `json:"user"`
	Sslmode bool   `json:"sslmode"`
	// Secret is a Openbao HTTP endpoint.
	Secret string `json:"secret"`
	// SecretKey is a Openbao JSON data field.
	SecretKey string `json:"secret_key"`
}
