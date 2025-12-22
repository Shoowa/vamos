package secrets

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"

	openbao "github.com/openbao/openbao/api/v2"

	"github.com/Shoowa/vamos/config"
)

type SkeletonKey struct {
	Openbao *openbao.Client
}

func (sk *SkeletonKey) Create(cfg *config.Config) {
	cfg.Secrets.Openbao.ReadToken()
	clientConfig := ReadConfig(cfg)
	client, err := BuildClient(clientConfig, cfg.Secrets.Openbao.Token)
	if err != nil {
		panic(err.Error())
	}
	sk.Openbao = client
}

func ReadConfig(cfg *config.Config) *openbao.Config {
	url := cfg.Secrets.Openbao.ReadConfig()
	clientConfig := openbao.DefaultConfig()
	clientConfig.Address = url
	return clientConfig
}

func BuildClient(obCfg *openbao.Config, token string) (*openbao.Client, error) {
	client, err := openbao.NewClient(obCfg)
	if err != nil {
		return nil, err
	}

	client.SetToken(token)
	return client, nil
}

func (sk *SkeletonKey) ReadPathAndKey(secretPath, key string) (string, error) {
	secret, secretErr := sk.Openbao.KVv2("secret").Get(context.Background(), secretPath)
	if secretErr != nil {
		return "", secretErr
	}

	v, ok := secret.Data[key].(string)
	if !ok {
		return "", errors.New("Type assertion failed on the value.")
	}
	return v, nil
}

func (sk *SkeletonKey) ReadTlsCertAndKey(cfg *config.Config, keyField, certField string) (*tls.Certificate, error) {
	cert64, certErr := sk.ReadPathAndKey(cfg.HttpServer.Certificate, certField)
	if certErr != nil {
		return nil, certErr
	}

	key64, keyErr := sk.ReadPathAndKey(cfg.HttpServer.Key, keyField)
	if keyErr != nil {
		return nil, keyErr
	}

	cert, decodeCertErr := base64.StdEncoding.DecodeString(cert64)
	if decodeCertErr != nil {
		return nil, decodeCertErr
	}

	key, decodeKeyErr := base64.StdEncoding.DecodeString(key64)
	if decodeKeyErr != nil {
		return nil, decodeKeyErr
	}

	pair, X509Err := tls.X509KeyPair([]byte(cert), []byte(key))
	if X509Err != nil {
		return nil, X509Err
	}

	return &pair, nil
}
