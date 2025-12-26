package cache

import (
	"context"
	"fmt"

	redis "github.com/redis/go-redis/v9"

	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/secrets"
)

func readPassword(c *secrets.SkeletonKey, cfg *config.Cache) string {
	secret, secretErr := c.ReadPathAndKey(cfg.Secret, cfg.SecretKey)
	if secretErr != nil {
		return ""
	}
	return secret
}

func configure(cfg *config.Config, sec *secrets.SkeletonKey) (*redis.Options, error) {
	hostAndPort := fmt.Sprintf("%v:%v", cfg.Cache.Host, cfg.Cache.Port)

	redisTLS, rtlsErr := sec.ConfigureTLSwithCA(cfg)
	if rtlsErr != nil {
		return nil, rtlsErr
	}

	return &redis.Options{
		TLSConfig: redisTLS,
		DB:        cfg.Cache.Db,
		Addr:      hostAndPort,
		CredentialsProviderContext: func(ctx context.Context) (string, string, error) {
			return cfg.Cache.User, readPassword(sec, cfg.Cache), nil
		},
	}, nil
}

func CreateClient(cfg *config.Config, sec *secrets.SkeletonKey) (*redis.Client, error) {
	opts, confErr := configure(cfg, sec)
	if confErr != nil {
		return nil, confErr
	}

	return redis.NewClient(opts), nil
}
