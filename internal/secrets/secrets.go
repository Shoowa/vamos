package secrets

import (
	openbao "github.com/openbao/openbao/api/v2"

	"vamos/internal/config"
)

func ReadConfig(cfg *config.Config) *openbao.Config {
	url := cfg.Secrets.Openbao.ReadConfig()
	clientConfig := openbao.DefaultConfig()
	clientConfig.Address = url
	return clientConfig
}
