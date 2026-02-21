package secrets

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"

	openbao "github.com/openbao/openbao/api/v2"

	"github.com/Shoowa/vamos/config"
)

// SkeletonKey wrap around an Openbao client. Perhaps it can wrap around other
// clients that read different storage.
type SkeletonKey struct {
	Openbao *openbao.Client
}

// Create is a method of the SkeletonKey. It is hardcoded for the Openbao
// client. It basically adds a token to an Openbao client, and adds the client
// to the SkeletonKey.
func (sk *SkeletonKey) Create(cfg *config.Config) {
	cfg.Secrets.Openbao.ReadToken()
	clientConfig := readConfig(cfg)
	client, err := buildClient(clientConfig, cfg.Secrets.Openbao.Token)
	if err != nil {
		panic(err.Error())
	}
	sk.Openbao = client
}

func readConfig(cfg *config.Config) *openbao.Config {
	url := cfg.Secrets.Openbao.ReadConfig()
	clientConfig := openbao.DefaultConfig()
	clientConfig.Address = url

	// If the config file possesses an entry for a X509 certificate in the OpenBao
	// portion, then configure the OpenBao client for TLS.
	if cfg.Secrets.Openbao.TlsClient.CertPath != "" {
		tls := openbao.TLSConfig{}
		tls.ClientCert = cfg.Secrets.Openbao.TlsClient.CertPath
		tls.ClientKey = cfg.Secrets.Openbao.TlsClient.KeyPath
		tls.CACert = cfg.Secrets.Openbao.TlsClient.CaPath
		tls.Insecure = false
		clientConfig.ConfigureTLS(&tls)
	}
	return clientConfig
}

func buildClient(obCfg *openbao.Config, token string) (*openbao.Client, error) {
	client, err := openbao.NewClient(obCfg)
	if err != nil {
		return nil, err
	}

	client.SetToken(token)
	return client, nil
}

// ReadPathAndKey expects an Openbao endpoint, and a JSON key.
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

// ReadTlsCertAndKey expects a custom struct named TlsSecret in the Config file.
// It will assemble a x509 certificate and key that is stored in Openbao as a
// base64 values.
func (sk *SkeletonKey) ReadTlsCertAndKey(tlsInfo *config.TlsSecret) (*tls.Certificate, error) {
	cert64, certErr := sk.ReadPathAndKey(tlsInfo.CertPath, tlsInfo.CertField)
	if certErr != nil {
		return nil, certErr
	}

	key64, keyErr := sk.ReadPathAndKey(tlsInfo.KeyPath, tlsInfo.KeyField)
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

// ReadIntermediateCA expects a custom struct named HttpServer in the Config
// file. It will read a base64 encoded value from Openbao, and return bytes.
func (sk *SkeletonKey) ReadIntermediateCA(cfg *config.HttpServer) ([]byte, error) {
	ca64, ca64Err := sk.ReadPathAndKey(cfg.SecretCA, cfg.SecretCAKey)
	if ca64Err != nil {
		return nil, ca64Err
	}

	cert, decodeCertErr := base64.StdEncoding.DecodeString(ca64)
	if decodeCertErr != nil {
		return nil, decodeCertErr
	}

	return cert, nil
}

// CreateCertPool expects a custom struct named HttpServer in the Config file.
// It will read a base64 encoded value from Openbao, then use that certificate
// to configure a certPool.
func (sk *SkeletonKey) CreateCertPool(cfg *config.HttpServer) (*x509.CertPool, error) {
	ca, caErr := sk.ReadIntermediateCA(cfg)
	if caErr != nil {
		return nil, caErr
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(ca)
	return certPool, nil
}

// ConfigureTLSwithCA expects a custom struct named HttpServer in the Config
// file. It will assemble a tls.Config with a CA, cert, and TLS 1.3
func (sk *SkeletonKey) ConfigureTLSwithCA(cfg *config.HttpServer) (*tls.Config, error) {
	clientCert, ccErr := sk.ReadTlsCertAndKey(cfg.TlsClient)
	if ccErr != nil {
		return nil, ccErr
	}

	certPool, cpErr := sk.CreateCertPool(cfg)
	if cpErr != nil {
		return nil, cpErr
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{*clientCert},
		RootCAs: certPool,
	}, nil
}

// LogicalRead expects an Openbao endpoint to GET.
func (sk *SkeletonKey) LogicalRead(secretPath string) (*openbao.Secret, error) {
	logicalClient := sk.Openbao.Logical()
	secret, secretErr := logicalClient.Read(secretPath)
	if secretErr != nil {
		return nil, secretErr
	}

	return secret, nil
}

type payload map[string]any

// LogicalWrite expects an Openbao endpoint and a map of data to PUT.
func (sk *SkeletonKey) LogicalWrite(path string, data payload) (*openbao.Secret, error) {
	logicalClient := sk.Openbao.Logical()
	secret, secretErr := logicalClient.Write(path, data)
	if secretErr != nil {
		return nil, secretErr
	}

	return secret, nil
}
