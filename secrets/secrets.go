package secrets

import (
	"context"
	"errors"

	openbao "github.com/openbao/openbao/api/v2"

	"github.com/Shoowa/vamos/config"
)

func ReadConfig(cfg *config.Config) *openbao.Config {
	url := cfg.Secrets.Openbao.ReadConfig()
	clientConfig := openbao.DefaultConfig()
	clientConfig.Address = url
	return clientConfig
}

func BuildClient(obCfg *openbao.Config, cfg *config.Config) (*openbao.Client, error) {
	client, err := openbao.NewClient(obCfg)
	if err != nil {
		return nil, err
	}

	client.SetToken(cfg.Secrets.Openbao.Token)
	return client, nil
}

func ReadSecret(c *openbao.Client, secretPath string) (string, error) {
	secret, secretErr := c.KVv2("secret").Get(context.Background(), secretPath)
	if secretErr != nil {
		return "", secretErr
	}

	v, ok := secret.Data["password"].(string)
	if !ok {
		return "", errors.New("Type assertion failed on value of PW field.")
	}
	return v, nil
}

func BuildAndRead(cfg *config.Config, secretPath string) (string, error) {
	oCfg := ReadConfig(cfg)

	c, cErr := BuildClient(oCfg, cfg)
	if cErr != nil {
		return "", cErr
	}

	pw, pwErr := ReadSecret(c, secretPath)
	if pwErr != nil {
		return "", pwErr
	}

	return pw, nil
}

func ReadPathAndKey(c *openbao.Client, secretPath, key string) (string, error) {
	secret, secretErr := c.KVv2("secret").Get(context.Background(), secretPath)
	if secretErr != nil {
		return "", secretErr
	}

	v, ok := secret.Data[key].(string)
	if !ok {
		return "", errors.New("Type assertion failed on value of PW field.")
	}
	return v, nil
}
